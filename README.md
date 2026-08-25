# FitCheck

Personal outfit decision engine — upload your clothes, plan trips, get AI-powered outfit recommendations from what you actually own.

## Stack

- **Go** — chi router, Vercel serverless + local server
- **SQLite** — local development
- **Supabase Postgres + Storage** — Vercel production
- **OpenAI Vision** (optional) — clothing analysis & outfit generation
- **Open-Meteo + Nominatim** — weather & geocoding (free, no key)

## Local development

```bash
export PATH="$HOME/.local/go/bin:$PATH"
cd ~/Projects/fitcheck
go run ./cmd/web
```

Open **http://localhost:8080** — uses SQLite + local `uploads/` folder.

## Deploy to Vercel

### 1. Create a Supabase project

1. Go to [supabase.com](https://supabase.com) → New project
2. Copy these from **Project Settings → API**:
   - `SUPABASE_URL`
   - `SUPABASE_ANON_KEY`
   - `SUPABASE_SERVICE_ROLE_KEY`
3. Copy **Database URL** (use the **Transaction pooler** connection string on port `6543`) as `DATABASE_URL`

### 2. Create a public storage bucket

In Supabase Dashboard → **Storage** → **New bucket**:
- Name: `closet`
- Public: **Yes**

(The app also tries to create this automatically on first boot.)

### 3. Deploy on Vercel

1. Import **https://github.com/pariharhemalatha-cyber/fitcheck** on [vercel.com](https://vercel.com)
2. Add environment variables:

| Variable | Required on Vercel | Description |
|---|---|---|
| `DATABASE_URL` | Yes | Supabase Postgres pooler URL |
| `SUPABASE_URL` | Yes | Supabase project URL |
| `SUPABASE_SERVICE_ROLE_KEY` | Yes | For storage uploads |
| `SUPABASE_ANON_KEY` | Recommended | Future auth |
| `GEMINI_API_KEY` | Recommended | Free AI vision + outfits ([get key](https://aistudio.google.com/apikey)) |
| `OPENAI_API_KEY` | Optional | Paid AI fallback |

3. Deploy — Vercel detects Go via root `main.go` + `vercel.json`

### How Vercel works

```
Browser → Vercel Go server (main.go) → chi router
              ├── Supabase Postgres (persistent data)
              ├── Supabase Storage (clothing photos)
              ├── Open-Meteo (weather)
              └── Gemini / OpenAI (optional)
```

Local dev uses SQLite + disk uploads when `DATABASE_URL` is not set.

## Features

| Screen | Route | What it does |
|---|---|---|
| Home | `/` | Style Me composer |
| Closet | `/closet` | Upload photos, AI auto-tags |
| Item detail | `/closet/items/{id}` | Edit AI tags |
| My Style | `/style` | Personal preferences |
| Plan | `/plan` | Location, dates, activities |
| Outfits | `/outfits` | Weather-aware recommendations |
| Trip | `/trip` | Packing + day-by-day outfits |
| Fit Check | `/fitcheck` | Selfie scoring + swap suggestions |

## Environment

```bash
cp .env.example .env
```

## Project structure

```
main.go               Vercel + production entrypoint
cmd/web/main.go       Local development server
internal/
  ai/                 OpenAI + fallbacks
  db/                 SQLite + Postgres migrations
  handlers/           HTTP handlers
  outfit/             Rule-based generator + packing
  server/             Router + bootstrap
  service/            AI + weather orchestration
  storage/            Local disk + Supabase Storage
  store/              SQLite + Postgres stores
  weather/            Geocoding + Open-Meteo
  web/                HTML templates
supabase/migrations/  Postgres schema reference
vercel.json           Route all traffic to Go handler
```
