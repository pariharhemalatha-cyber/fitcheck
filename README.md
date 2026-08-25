# FitCheck

Personal outfit decision engine — upload your clothes, plan trips, get AI-powered outfit recommendations from what you actually own.

## Stack

- **Go** — HTTP server (chi router)
- **SQLite** — local persistence (`fitcheck.db`)
- **HTML templates + HTMX + Tailwind CDN** — server-rendered UI
- **OpenAI Vision** (optional) — clothing analysis & outfit generation
- **Open-Meteo + Nominatim** — weather & geocoding (free, no key)
- **Supabase migration** — ready for cloud Postgres when you connect it

## Quick start

```bash
export PATH="$HOME/.local/go/bin:$PATH"
cd ~/Projects/fitcheck
go run ./cmd/web
```

Open **http://localhost:8080**

Optional: add `OPENAI_API_KEY` to `.env` for vision-based clothing analysis and smarter outfits.

## Features

| Screen | Route | What it does |
|---|---|---|
| Home | `/` | Style Me composer |
| Closet | `/closet` | Upload photos, AI auto-tags, browse by category |
| Item detail | `/closet/items/{id}` | View/edit AI tags (color, pattern, formality, season) |
| My Style | `/style` | Personal preferences (comfort, photo-ready, no-repeat) |
| Plan | `/plan` | Location, dates, activities, formality |
| Outfits | `/outfits` | Weather-aware outfit recommendations from your closet |
| Trip | `/trip` | Packing list + day-by-day outfits with intelligent reuse |
| Fit Check | `/fitcheck` | Upload selfie, get fit score + swap suggestions |

## How it works

1. **Upload clothes** → saved to `uploads/`, analyzed by AI (or heuristics without API key)
2. **Plan a trip/day** → geocodes location, fetches Open-Meteo forecast
3. **Outfit engine** → filters closet by weather/formality, builds combinations, optionally uses OpenAI
4. **Trip packing** → minimizes pieces, reuses pants/shoes, rotates tops
5. **Fit Check** → scores your selfie against selected outfit items

## Environment

```bash
cp .env.example .env
```

| Variable | Required | Description |
|---|---|---|
| `PORT` | No | Default `8080` |
| `SQLITE_PATH` | No | Default `fitcheck.db` |
| `OPENAI_API_KEY` | No | Enables vision analysis & AI outfits |
| `SUPABASE_*` | No | For future cloud auth/storage |

## Project structure

```
cmd/web/              HTTP server entrypoint
internal/
  ai/                 OpenAI client, analyze, generate, fitcheck
  config/             Environment config
  db/                 SQLite schema + CRUD
  handlers/           HTTP handlers
  outfit/             Rule-based generator + packing solver
  server/             Chi router setup
  service/            Orchestrates AI + weather + store
  storage/            Local file uploads
  store/              Store interface + SQLite implementation
  weather/            Geocoding + Open-Meteo forecast
  web/                HTML templates
supabase/migrations/  Postgres schema (for Supabase cloud)
uploads/              Clothing photos (gitignored)
fitcheck.db           SQLite database (gitignored)
```

## Development phases

- **Phase 0** ✅ Go server, 5 screens
- **Phase 1** ✅ SQLite persistence, file uploads, AI clothing analysis
- **Phase 2** ✅ Weather API, real outfit generation
- **Phase 3** ✅ Trip packing solver with real closet items
- **Phase 4** ✅ Fit Check (selfie scoring)
