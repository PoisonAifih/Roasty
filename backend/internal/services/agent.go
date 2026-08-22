package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// maxAgentSteps caps the reason/act loop. Without it a confused model can
// call tools forever and burn the API budget.
const maxAgentSteps = 8

// TraceEvent is one visible step of the agent's work. The UI renders these
// live, which is the only way a user can tell tools actually ran.
type TraceEvent struct {
	Type    string `json:"type"` // step | tool_call | tool_result | final | error
	Tool    string `json:"tool,omitempty"`
	Message string `json:"message"`
	Step    int    `json:"step,omitempty"`
}

type Agent struct {
	pool  *pgxpool.Pool
	ai    *AIClient
	scout *ScoutService
	inv   *InventoryService
	crm   *CRMService
	rag   *RAGService
}

func NewAgent(pool *pgxpool.Pool, ai *AIClient, scout *ScoutService, inv *InventoryService, crm *CRMService, rag *RAGService) *Agent {
	return &Agent{pool: pool, ai: ai, scout: scout, inv: inv, crm: crm, rag: rag}
}

const agentSystemPrompt = `You are a sourcing advisor for an Indonesian coffee roastery.

You have tools that read the roastery's own database, search the live web,
and query two knowledge stores:
- query_knowledge_base: SNI quality standards, grading, origin profiles, market context.
- find_similar_beans: semantic search over a bean sourcing catalog with supplier
  contacts. Use this whenever the user asks for a bean recommendation by taste,
  use-case, or similarity to another origin. It returns supplier names, phone
  numbers, locations, and minimum order quantities.

Use tools before answering — never invent prices, stock levels, or supplier contacts.

Workflow for sourcing questions:
1. Call find_similar_beans with a description of the desired profile or use-case.
2. Cross-check results against list_beans / score_beans for current stock and price.
3. If the user gives a budget, call build_basket for an allocation plan.
4. Give a concrete answer: which beans to buy, from which supplier, quantity in kg,
   total cost in rupiah, and direct contact info.

Prices are Indonesian rupiah per kilogram. Always reply in English.
Be specific and short. No markdown headings, no bullet symbols.`

func agentTools() []ToolDef {
	obj := func(props map[string]any, required ...string) map[string]any {
		if required == nil {
			required = []string{}
		}
		return map[string]any{
			"type":       "object",
			"properties": props,
			"required":   required,
		}
	}
	num := map[string]any{"type": "number"}
	str := map[string]any{"type": "string"}

	return []ToolDef{
		{Type: "function", Function: FunctionDef{
			Name:        "list_beans",
			Description: "List every green bean in the catalogue with price per kg, stock kg, quality score, humidity, harvest estimate and sell price. Call this first.",
			Parameters:  obj(map[string]any{}),
		}},
		{Type: "function", Function: FunctionDef{
			Name:        "score_beans",
			Description: "Rank beans by fit against a budget, target weight and sourcing channel. Returns fit score plus a price/quality/supply/margin breakdown and profit per kg.",
			Parameters: obj(map[string]any{
				"budget":    num,
				"weight_kg": num,
				"channel":   map[string]any{"type": "string", "enum": []string{"farmer", "middleman", ""}},
			}),
		}},
		{Type: "function", Function: FunctionDef{
			Name:        "check_inventory",
			Description: "Current stock, average daily sales, days of cover, trend strength and restock urgency for every origin.",
			Parameters:  obj(map[string]any{}),
		}},
		{Type: "function", Function: FunctionDef{
			Name:        "sales_history",
			Description: "Total kg sold for one origin over the last N days. Use it to check what actually sells before recommending a purchase.",
			Parameters: obj(map[string]any{
				"origin": str,
				"days":   num,
			}, "origin"),
		}},
		{Type: "function", Function: FunctionDef{
			Name:        "build_basket",
			Description: "Allocate a budget across several beans at once and return three purchase plans (max_profit, lowest_risk, balanced) with kg, rupiah cost and projected profit per line. Use this whenever the user gives a budget - a roastery buys a mix, not a single origin.",
			Parameters: obj(map[string]any{
				"budget":  num,
				"max_kg":  num,
				"channel": map[string]any{"type": "string", "enum": []string{"farmer", "middleman", ""}},
			}, "budget"),
		}},
		{Type: "function", Function: FunctionDef{
			Name:        "market_price",
			Description: "Search the live web for current market listings and suppliers for an origin. Use when you need prices outside the roastery's own records.",
			Parameters: obj(map[string]any{
				"origin":  str,
				"variety": str,
			}, "origin"),
		}},
		{Type: "function", Function: FunctionDef{
			Name:        "find_similar_beans",
			Description: "Semantic search over the bean sourcing catalog. Given a natural-language description of the desired bean (flavor profile, aroma, use-case, budget tier, processing method, or similarity to another origin), returns the closest matches with full supplier contact details: name, type (koperasi/pedagang/online), WhatsApp/phone, location, minimum order, and notes. Always use this when the user asks which bean to buy or where to source a specific type of bean.",
			Parameters: obj(map[string]any{
				"query": str,
			}, "query"),
		}},
		{Type: "function", Function: FunctionDef{
			Name:        "query_knowledge_base",
			Description: "Search the knowledge base of Indonesian coffee quality standards (SNI), origin profiles, grading systems (SNI vs SCA), processing methods, market context, and CRM patterns. Use for any question about quality grades, defect values, moisture limits, origin characteristics, harvest seasons, pricing context, or best practices.",
			Parameters: obj(map[string]any{
				"question": str,
			}, "question"),
		}},
	}
}

// Run drives the reason/act loop, emitting a trace event for every step so the
// caller can stream progress. It returns the agent's final answer.
func (ag *Agent) Run(ctx context.Context, goal string, emit func(TraceEvent)) (string, error) {
	if emit == nil {
		emit = func(TraceEvent) {}
	}
	if !ag.ai.HasKey() {
		return "", fmt.Errorf("agent needs OPENROUTER_API_KEY")
	}
	goal = strings.TrimSpace(goal)
	if goal == "" {
		return "", fmt.Errorf("goal required")
	}

	msgs := []ChatMessage{
		{Role: "system", Content: agentSystemPrompt},
		{Role: "user", Content: goal},
	}
	tools := agentTools()

	for step := 1; step <= maxAgentSteps; step++ {
		emit(TraceEvent{Type: "step", Step: step, Message: "Thinking…"})

		reply, err := ag.ai.Chat(ctx, msgs, tools)
		if err != nil {
			emit(TraceEvent{Type: "error", Message: err.Error()})
			return "", err
		}
		msgs = append(msgs, reply)

		// No tool calls means the model is done reasoning.
		if len(reply.ToolCalls) == 0 {
			final := strings.TrimSpace(reply.Content)
			if final == "" {
				final = "No recommendation produced."
			}
			emit(TraceEvent{Type: "final", Step: step, Message: final})
			return final, nil
		}

		for _, call := range reply.ToolCalls {
			emit(TraceEvent{
				Type:    "tool_call",
				Tool:    call.Function.Name,
				Step:    step,
				Message: describeCall(call),
			})

			result, err := ag.execute(ctx, call.Function.Name, call.Function.Arguments)
			if err != nil {
				// Feed the error back rather than aborting; the model can
				// recover by trying a different tool.
				result = fmt.Sprintf("error: %v", err)
			}
			emit(TraceEvent{
				Type:    "tool_result",
				Tool:    call.Function.Name,
				Step:    step,
				Message: truncate(result, 240),
			})

			msgs = append(msgs, ChatMessage{
				Role:       "tool",
				ToolCallID: call.ID,
				Name:       call.Function.Name,
				Content:    result,
			})
		}
	}

	emit(TraceEvent{Type: "error", Message: "agent hit the step limit"})
	return "", fmt.Errorf("agent exceeded %d steps without finishing", maxAgentSteps)
}

func (ag *Agent) execute(ctx context.Context, name, rawArgs string) (string, error) {
	var args map[string]any
	if strings.TrimSpace(rawArgs) != "" {
		if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
			return "", fmt.Errorf("bad arguments: %w", err)
		}
	}
	argStr := func(k string) string {
		if v, ok := args[k].(string); ok {
			return strings.TrimSpace(v)
		}
		return ""
	}
	argNum := func(k string) float64 {
		if v, ok := args[k].(float64); ok {
			return v
		}
		return 0
	}

	switch name {
	case "list_beans":
		beans, err := ag.scout.ListBeans(ctx)
		if err != nil {
			return "", err
		}
		return toJSON(beans)

	case "score_beans":
		recs, err := ag.scout.scoreOnly(ctx, RecommendInput{
			Budget:   argNum("budget"),
			WeightKg: argNum("weight_kg"),
			Channel:  argStr("channel"),
		})
		if err != nil {
			return "", err
		}
		return toJSON(recs)

	case "check_inventory":
		rows, err := ag.inv.snapshot(ctx)
		if err != nil {
			return "", err
		}
		return toJSON(rows)

	case "sales_history":
		origin := argStr("origin")
		if origin == "" {
			return "", fmt.Errorf("origin required")
		}
		days := int(argNum("days"))
		if days <= 0 {
			days = 90
		}
		var total float64
		var orders int
		err := ag.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(s.qty_kg), 0), COUNT(s.id)
			FROM sales s JOIN beans b ON b.id = s.bean_id
			WHERE b.origin ILIKE $1 AND s.sold_at >= CURRENT_DATE - $2::int`,
			"%"+origin+"%", days).Scan(&total, &orders)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`{"origin":%q,"days":%d,"total_kg":%.1f,"transactions":%d,"avg_kg_per_day":%.2f}`,
			origin, days, total, orders, total/float64(days)), nil

	case "build_basket":
		plans, err := ag.scout.BuildBaskets(ctx, BasketInput{
			Budget:  argNum("budget"),
			MaxKg:   argNum("max_kg"),
			Channel: argStr("channel"),
		})
		if err != nil {
			return "", err
		}
		return toJSON(plans)

	case "market_price":
		shops, err := ag.scout.FindShops(ctx, argStr("origin"), argStr("variety"))
		if err != nil {
			return "", err
		}
		return toJSON(shops)

	case "find_similar_beans":
		query := argStr("query")
		if query == "" {
			return "", fmt.Errorf("query required")
		}
		if ag.rag == nil {
			return "Bean catalog not available.", nil
		}
		result, err := ag.rag.FindSimilarBeans(ctx, query, 4)
		if err != nil {
			return fmt.Sprintf("error searching bean catalog: %v", err), nil
		}
		return result, nil

	case "query_knowledge_base":
		question := argStr("question")
		if question == "" {
			return "", fmt.Errorf("question required")
		}
		if ag.rag == nil {
			return "Knowledge base not available.", nil
		}
		context := ag.rag.QueryText(ctx, question)
		if context == "" {
			return "No relevant information found in knowledge base.", nil
		}
		return context, nil
	}

	return "", fmt.Errorf("unknown tool %q", name)
}

func describeCall(call ToolCall) string {
	args := strings.TrimSpace(call.Function.Arguments)
	if args == "" || args == "{}" {
		return call.Function.Name
	}
	return call.Function.Name + " " + truncate(args, 120)
}

func toJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
