package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// This file holds the mutating operations. Everything else in the package is
// read-only; these exist so the agent can act on its own recommendations
// instead of only describing them.

type RecordSaleInput struct {
	BeanID string  `json:"bean_id"`
	QtyKg  float64 `json:"qty_kg"`
	SoldAt string  `json:"sold_at"` // YYYY-MM-DD; empty means today
}

// RecordSale inserts a sale and decrements the bean's stock in one
// transaction, so stock can never drift from sales history.
func (s *InventoryService) RecordSale(ctx context.Context, in RecordSaleInput) error {
	if in.BeanID == "" {
		return fmt.Errorf("bean_id required")
	}
	if in.QtyKg <= 0 {
		return fmt.Errorf("qty_kg must be greater than zero")
	}

	soldAt := time.Now()
	if in.SoldAt != "" {
		parsed, err := time.Parse("2006-01-02", in.SoldAt)
		if err != nil {
			return fmt.Errorf("sold_at must be YYYY-MM-DD: %w", err)
		}
		soldAt = parsed
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var stock float64
	err = tx.QueryRow(ctx,
		`SELECT stock_kg FROM beans WHERE id = $1 FOR UPDATE`, in.BeanID).Scan(&stock)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("bean %s not found", in.BeanID)
	}
	if err != nil {
		return err
	}
	if stock < in.QtyKg {
		return fmt.Errorf("insufficient stock: have %.1fkg, tried to sell %.1fkg", stock, in.QtyKg)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO sales (bean_id, qty_kg, sold_at) VALUES ($1, $2, $3)`,
		in.BeanID, in.QtyKg, soldAt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE beans SET stock_kg = stock_kg - $1 WHERE id = $2`,
		in.QtyKg, in.BeanID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type AdjustStockInput struct {
	StockKg float64 `json:"stock_kg"`
	Note    string  `json:"note"`
}

// AdjustStock sets an absolute stock level and writes an audit row in the same
// transaction, so the log can never drift from the actual stock value.
func (s *InventoryService) AdjustStock(ctx context.Context, beanID string, in AdjustStockInput) error {
	if beanID == "" {
		return fmt.Errorf("bean id required")
	}
	if in.StockKg < 0 {
		return fmt.Errorf("stock_kg cannot be negative")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var oldStock float64
	err = tx.QueryRow(ctx, `SELECT stock_kg FROM beans WHERE id = $1 FOR UPDATE`, beanID).Scan(&oldStock)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("bean %s not found", beanID)
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO stock_adjustments (bean_id, old_stock, new_stock, note) VALUES ($1, $2, $3, NULLIF($4, ''))`,
		beanID, oldStock, in.StockKg, in.Note); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE beans SET stock_kg = $1 WHERE id = $2`, in.StockKg, beanID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// MarkContacted stamps today's date so follow-up suggestions stop repeating
// for a shop that has already been chased.
func (s *CRMService) MarkContacted(ctx context.Context, shopID string) error {
	if shopID == "" {
		return fmt.Errorf("shop id required")
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE coffee_shops SET last_contacted_at = CURRENT_DATE WHERE id = $1`, shopID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("shop %s not found", shopID)
	}
	return nil
}
