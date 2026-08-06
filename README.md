# Roasty — Roastery Scout AI

Decision assistant for coffee roasteries. Enter inputs on the dashboard, get scored recommendations and short AI narratives for sourcing, stock, and customer follow-up.

## Features

- **Bean Scout**: Rank green-bean options against budget and channel, with AI notes plus live online/offline shop links from web search
- **Smart Inventory**: Surface restock priorities from stock, sales, and harvest signals
- **Customer & Payment**: Flag shops that need payment or order follow-up

## Tech stack

| Layer | Stack |
|-------|--------|
| Frontend | React, TypeScript, Vite, Tailwind CSS v4, shadcn/ui |
| Backend | Go REST API |
| Database | PostgreSQL (Docker) |
| AI | OpenRouter · `deepseek/deepseek-v4-flash` |

## Prerequisites

- Docker
- Go 1.24+
- Node 20+
- OpenRouter API key

## Setup

```bash
cp .env.example .env
cp frontend/.env.example frontend/.env
```

Set `OPENROUTER_API_KEY` in root `.env`. Optional: `AI_MODEL` (default `deepseek/deepseek-v4-flash`). Frontend uses `VITE_API_URL`.

### 1. Database

```bash
docker compose up -d
```

Default Postgres host port is `5433`. To re-seed from scratch:

```bash
docker compose down -v && docker compose up -d
```

### 2. Backend

```bash
cd backend
go run ./cmd/api
```

API: `http://localhost:8014`

| Method | Path | Notes |
|--------|------|--------|
| `GET` | `/api/health` | Health check |
| `GET` | `/api/beans` | Bean catalog |
| `POST` | `/api/scout/recommend` | Body: `{ "budget": 5000000, "weight_kg": 50, "channel": "farmer" }`—budget/weight optional |
| `GET` | `/api/scout/shops` | Query: `origin`, `variety`—live web shop links |
| `GET` | `/api/inventory/suggestions` | Restock suggestions |
| `GET` | `/api/crm/follow-ups` | Customer and payment follow-ups |

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

Open the Vite URL (usually `http://localhost:5173`).
