# CS:S Surf Stats

Small full-stack app for browsing CS:S surf map stats pulled from KSF.

## Stack

- Backend: Go HTTP server + SQLite
- Frontend: React + TypeScript + Vite + Tailwind CSS
- Deployment: Docker Compose with nginx proxying the API

## Current Data Model

The app stores one `maps` table with:

- KSF map id and name
- Tier
- Added timestamp
- Completion count
- Playtime in seconds
- Derived completions per hour
- Bonus count
- Linear vs staged flag
- Notes and `updated_at`

The current API consumed by the frontend is `GET /api/maps`.

## Local Development

Requirements:

- Go 1.24+
- Node 20+
- `npm`

Install frontend and root dependencies:

```bash
npm install
cd frontend && npm install
```

Run both apps:

```bash
npm run dev
```

That starts:

- Go API on `http://localhost:8080`
- Vite frontend on its default port, with `/api` proxied to the backend

## Database Setup

The backend expects SQLite at `backend/data/surfstats.db`.

Create the database and schema once:

```bash
mkdir -p backend/data
sqlite3 backend/data/surfstats.db < backend/migrations/001_create_maps.sql
```

## Data Ingestion

There are two supported ingestion entry points:

1. Scrape map metadata and upsert it into SQLite:

```bash
cd backend
go run ./cmd/scrape-maps
```

2. Run the weekly ingestion job, which refreshes map metadata and completion counts:

```bash
cd backend
go run ./cmd/weekly-ingest
```

The weekly job uses a fixed seed Steam ID to fetch map record pages and infer total completions from the KSF player-record API.

## Docker

Build and run the full stack:

```bash
docker compose up --build
```

Services:

- Frontend on `http://localhost`
- Backend on `http://localhost:8080`

SQLite data is persisted via `./backend/data`.
