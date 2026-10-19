# Virtual Garden Management System

A full-stack virtual garden game: a Go API (Gin, GORM, PostgreSQL, Redis) and a React + TypeScript frontend (Vite, Tailwind CSS, shadcn/ui). Players create gardens, buy seeds, grow and harvest plants under changing weather, share gardens with other players, and trade on a marketplace.

![Virtual Garden System](https://github.com/user-attachments/assets/e5a769e1-416d-490e-b497-1e311bdd0498)

## 🚀 Quick Start

### Prerequisites

- **Go** 1.26 or higher; `go.mod` pins the 1.27.2 toolchain, which Go downloads automatically
- **Node.js** 22.13 or higher
- **Docker** with Compose, for PostgreSQL and Redis
- **Air** for API live reload: `make install-air`

### Run locally

```bash
git clone https://github.com/edumes/my-garden.git
cd my-garden
cp env.example .env

# PostgreSQL and Redis, using the credentials and ports from .env
docker compose up -d --wait

# API with live reload on http://localhost:8080
make dev

# In another terminal: frontend on http://localhost:5173
cd frontend
npm ci
npm run dev
```

The frontend calls the API at the relative path `/api/v1`, and Vite forwards it to `http://localhost:8080` during `npm run dev` and `npm run preview`. Set `VITE_API_URL` when the frontend is served from a different origin than the API.

If PostgreSQL or Redis already run on your machine, stop them or change `DB_PORT` and `REDIS_PORT` in `.env`; Compose publishes the containers on those ports.

Stop the services with `docker compose down`. Add `-v` to also delete the local database.

On startup the API creates or updates its tables with GORM `AutoMigrate` and seeds reference data such as plant types.

## 🛠️ Commands

### API (repository root)

| Command | Description |
|---|---|
| `make dev` | Run the API with live reload (Air) |
| `make run` | Run the API without live reload |
| `make build` | Build the server binary into `bin/server` |
| `make test` | Run the Go unit tests |
| `make test-integration` | Run the unit and PostgreSQL integration tests |
| `make test-e2e` | Run the browser smoke tests |
| `make vet` | Run `go vet` |
| `make fmt` | Format the Go code |
| `make swagger` | Regenerate the Swagger files in `docs/` |
| `make check` | Run every check CI runs (see [Testing and Validation](#-testing-and-validation)) |
| `make clean` | Remove build output |

### Frontend (`frontend/`)

| Command | Description |
|---|---|
| `npm run dev` | Start the development server on http://localhost:5173 |
| `npm run build` | Type-check and build for production into `frontend/dist` |
| `npm run preview` | Serve the production build |
| `npm run lint` | Run ESLint |
| `npm run test:e2e` | Run the browser smoke tests |

## 📚 API Reference

While the API runs, Swagger UI is available at http://localhost:8080/api/v1/swagger/index.html. The specification in `docs/swagger.json` and `docs/swagger.yaml` is generated from the handler annotations; run `make swagger` after changing an endpoint and commit the result.

`GET /health` reports whether the server is up.

## 🎮 Features

- **Accounts**: registration, login, JWT refresh and logout, profile view and update.
- **Gardens**: create, rename, and delete gardens; plant seeds, harvest, and remove plants. A server-side game tick moves plants through seed → sprout → growing → mature → harvestable.
- **Progression**: harvesting earns coins (harvest value × yield). Players have experience and levels, but no action grants experience yet.
- **Store**: browse seeds and buy them with coins. Purchases of the same seed add up in one inventory entry, and balances and seed counts can never go negative.
- **Weather**: current weather, forecast, history, and seasons. The current weather multiplies plant growth.
- **Garden sharing**: access links with permissions, expiration dates, and usage limits; joining shared gardens and managing members' permissions. Permissions are independent capabilities (view, plant, harvest, manage): manage grants all of them, any share grants view, and only the owner can delete the garden or manage its shares.
- **Marketplace**: delivery requests and plant listings paid in coins. Sales and deliveries settle atomically and apply once, even when repeated or concurrent. See [docs/MARKETPLACE.md](docs/MARKETPLACE.md).
- **Audit log**: API actions are recorded, and each player can query their own. See [docs/AUDIT_SYSTEM.md](docs/AUDIT_SYSTEM.md).

### Not implemented yet

- Real-time updates, game status, and leaderboards: `/api/v1/ws/garden/{gardenId}`, `/api/v1/game/status`, and `/api/v1/game/leaderboard` return "coming soon" placeholders.
- Achievements: seeded in the database, but no API exposes them.
- The settings page shows notification, two-factor authentication, and account deletion controls that have no backend.

## 📁 Project Structure

```
my-garden/
├── cmd/server/          # API entry point
├── internal/
│   ├── audit/           # Audit logging
│   ├── auth/            # Accounts and authentication
│   ├── config/          # Environment configuration
│   ├── database/        # Database connection
│   ├── garden/          # Gardens, plants, and sharing
│   ├── marketplace/     # Delivery requests and plant listings
│   ├── middleware/      # HTTP middleware
│   ├── migrations/      # AutoMigrate and seed data
│   ├── store/           # Seed store and inventory
│   ├── testdb/          # Throwaway databases for integration tests
│   └── weather/         # Weather system
├── pkg/
│   ├── auth/            # JWT utilities
│   ├── db/              # Database interface
│   ├── game/            # Game engine (growth and weather ticks)
│   └── utils/           # Shared helpers
├── frontend/            # React frontend (browser smoke tests in frontend/e2e)
├── docs/                # Generated Swagger files and feature docs
├── compose.yaml         # Local PostgreSQL and Redis
├── env.example          # Environment variables template
└── Makefile
```

## ✅ Testing and Validation

| Suite | Command | Needs |
|---|---|---|
| Unit | `make test` | Nothing |
| PostgreSQL integration | `make test-integration` | PostgreSQL (`docker compose up -d --wait`) |
| Browser smoke | `make test-e2e` | PostgreSQL, Redis, and Google Chrome |

**Integration tests** carry the `//go:build integration` tag and call the real HTTP API. `testdb.New(t)` from `internal/testdb` gives each test its own randomly named database on the server from the `DB_*` settings (environment first, then `.env`), applies the migrations and seed data, and drops the database when the test ends. Start new integration tests the same way.

**Browser smoke tests** live in `frontend/e2e` and run with Playwright in the installed Google Chrome. Playwright starts the API and the Vite dev server unless they are already running. Each run signs up a new player in the configured database.

One command runs the same checks as CI:

```bash
docker compose up -d --wait
(cd frontend && npm ci)   # once, or after dependency changes
make check
```

`make check` verifies Go formatting, runs `go vet`, the unit and integration tests, fails if the committed Swagger files are stale, then runs ESLint, the TypeScript type check, and the production build. GitHub Actions ([.github/workflows/ci.yml](.github/workflows/ci.yml)) runs it, with a PostgreSQL service, on every push to `main` or `dev` and on every pull request. The browser smoke tests run locally only.

## 🗺️ Roadmap

The development plan lives in [docs/plans](docs/plans/README.md).

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
