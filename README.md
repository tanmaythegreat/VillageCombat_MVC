# 🏰 Village Combat

A web-based village-building and combat game inspired by *Clash of Clans* — built from scratch using **Go**, **MVC architecture**, **JWT authentication**, **PostgreSQL**, and a **Three.js** 3D frontend.

Players build a village on a grid, manage resources, train troops, and raid other players' bases in a real-time, server-authoritative physics battle simulation.

🎮 **[Play Village Combat Live](https://villagecombat-mvc-ykep.onrender.com/Login.html)**


## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
  - [Authentication](#authentication)
  - [Database Design](#database-design)
  - [Battle Engine](#battle-engine)
  - [Frontend](#frontend)
- [Getting Started](#getting-started)
  - [Run with Docker (recommended)](#run-with-docker-recommended)
  - [Run locally](#run-locally)
- [Environment Variables](#environment-variables)
- [Database Migrations](#database-migrations)
- [API Reference](#api-reference)
  - [REST Endpoints](#rest-endpoints)
  - [WebSocket Protocol](#websocket-protocol)
- [Testing](#testing)
- [How to Play](#how-to-play)
- [Known Issues](#known-issues)
- [Roadmap](#roadmap)
- [License](#license)

---

## Features

- **Village building** on a grid — place, move, upgrade, and repair buildings with real construction timers
- **Resource economy** — Gold, Elixir, and Dark Elixir generation, storage, and collection
- **Troop training** — Barbarians, Archers, Giants, Goblins, Wall Breakers, Balloons, Minions, Wizards
- **Real-time PvP battles** — server-authoritative combat simulation streamed over WebSockets
- **Power-based matchmaking** — opponents are found based on attack/defence power range
- **Revenge system** — look up and re-attack a specific player by username
- **Battle replays & history** — past battles are recorded and can be replayed
- **JWT auth with rotating refresh tokens** — short-lived access tokens, bcrypt-hashed refresh tokens
- **3D village rendering** with Three.js, including object pooling for efficient building spawns

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26, standard `net/http` |
| Real-time | [gorilla/websocket](https://github.com/gorilla/websocket) |
| ORM / DB Driver | [GORM](https://gorm.io/) + [pgx](https://github.com/jackc/pgx) (PostgreSQL) |
| Database | PostgreSQL 15 |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Auth | Hand-rolled HMAC-SHA256 JWTs, `bcrypt` (`golang.org/x/crypto`) |
| Frontend | Vanilla JS (ES modules) + [Three.js](https://threejs.org/) |
| Testing | Go's built-in `testing` package + [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) |
| Containerization | Docker + Docker Compose |

## Project Structure

```
VillageCombat_MVC/
├── main.go                  # Entry point: runs migrations, wires routes, starts the server
├── auth/                    # JWT issuing/verification, refresh-token rotation, password rules
├── controllers/             # HTTP + WebSocket handlers (the "C" in MVC)
│   ├── web_socket.go         # WebSocket upgrade + action router (game's main event loop)
│   ├── login.go / register.go / refreshJWT.go
│   ├── create_building.go / upgrade_building.go / move_building.go / repair_building.go
│   ├── train_troop.go / collect_resource.go / collect_all_resource.go
│   ├── inicial_load.go / all_building_data.go / check_construction_work.go
├── models/                  # Data layer (the "M" in MVC) — GORM models + hand-written queries
│   ├── Models.go              # Struct definitions, table names, custom Postgres array types
│   └── Database.go            # All DB access: users, buildings, troops, resources, matchmaking
├── battle/                   # Real-time battle simulation engine
│   ├── Battle.go               # Matchmaking, troop/defense AI, damage resolution, simulation loop
│   └── replay.go               # Battle replay playback
├── db/migrations/            # Versioned SQL schema (golang-migrate, up/down pairs)
├── public/                   # Static frontend (served directly, no build step)
│   ├── Login.html / .css / .js
│   ├── Game_village.html / .css
│   ├── src/                    # Frontend JS modules (controllers/core/models/views)
│   ├── THREE/                  # Vendored Three.js + GLTFLoader
│   └── Models/                 # Sprite/texture assets for buildings & troops
├── Dockerfile
├── docker-compose.yml
├── Makefile                  # Migration helper commands
└── .env.example
```

## Architecture

The project follows an **MVC** pattern adapted for a real-time game server:

- **Models** (`models/`) own all database access and game-data structs. Nothing outside this package talks to the database directly.
- **Controllers** (`controllers/`) handle HTTP requests and WebSocket actions, translating client intent into model calls.
- **Views** are the static HTML/CSS/JS pages in `public/`, plus the JSON payloads the server streams back over the WebSocket, which the frontend renders into the 3D scene.

### Authentication

Auth is implemented by hand rather than with a JWT library:

- An `AccessToken` is `header.payload.signature`, where the signature is `HMAC-SHA256(header.payload, secret)`. It's short-lived (4 hours).
- A separate **refresh token** is generated from `crypto/rand` (never `math/rand`), returned to the client once, and only its **bcrypt hash** is stored server-side — the raw token is never persisted.
- Refreshing a session checks that the refresh token matches the stored hash **and** that the request's user agent matches the one used at login, then rotates in a brand-new access + refresh token pair.
- Login/registration compare password hashes with `bcrypt`, and a dummy `bcrypt.CompareHashAndPassword` call runs even on a "user not found" path specifically to keep the response timing constant and avoid leaking whether a username/email exists.
- Password registration enforces a minimum length plus upper/lower/digit/special-character requirements.

### Database Design

The schema is split across migrations by concern:

1. **Users & JWT** — `users`, `refresh_tokens`
2. **Static game data** — `troop_configs`, `building_configs_base`, `troop_level_stats`, `building_level_stats`, `upgrade_costs`, plus per-category stat tables (`defense_building_stats`, `resource_building_stats`, `army_building_stats`) and their `_level_stats` counterparts. Balance data (base stats + per-level stats + upgrade costs) is seeded via `000005_static_game_data_seed.up.sql`.
3. **Player state** — `placed_buildings`, `trained_troops`, `construction_tasks`, `user_data` (gold/elixir/dark elixir balances and capacities)
4. **Battle data** — `battle_history`, `battle_troop_losses`, `buildings_broken`, `user_status` (online/in-battle + attack/defence power for matchmaking), and `battle_record` (for replays)

Per-category stat tables (rather than one polymorphic table, or Postgres arrays) were a deliberate choice after running into incomplete GORM support for native array columns.

### Battle Engine

Battles are simulated **server-side** and streamed live to both players:

- `models.FindOpponent` matches players whose defence power falls within a range of the attacker's attack power.
- `battle.StartMatch` builds a `BattleState` containing every placed building, its defense stats (if it's a defensive structure), and the attacker's deployed troops.
- A per-battle goroutine runs the simulation loop (troop movement/targeting, defense building AI, damage and health tracking), broadcasting state to both the attacker and defender (if online) over their existing WebSocket connections.
- Outcomes (resources looted, troop losses, buildings destroyed) are persisted to the battle-history tables, and full battles can be replayed later via the `REPLAY` action.

### Frontend

The client is plain ES modules with no bundler/build step:

- **Login flow**: `Login.html`/`Login.js` hit `/register` and `/login`, then store the returned tokens in `localStorage` and redirect to the game.
- **Game load**: `Game_village.html` reads the token from `localStorage`, opens a WebSocket to `/ws`, authenticates, then requests placed-building data and (if not cached) the full static building/troop catalog before rendering the map.
- **Rendering**: Three.js draws the 3D village on a grid; building placement uses raycasting against the ground plane.
- **Object pooling**: instead of destroying/recreating meshes, buildings that go out of view are hidden and pooled, then re-shown and repositioned on demand — spawning is comparatively expensive, so reuse is preferred.

## Getting Started

### Run with Docker (recommended)

**Prerequisites:** Docker and Docker Compose.

```bash
git clone https://github.com/tanmaythegreat/VillageCombat_MVC.git
cd VillageCombat_MVC
cp .env.example .env   # then edit the values, especially JWT_SECRET_KEY
docker compose up --build
```

This starts a PostgreSQL container plus the app container, applies all pending migrations automatically on boot, and serves the game at **http://localhost:8080**.

### Run locally

**Prerequisites:** Go 1.26+, PostgreSQL 15, and the [`migrate` CLI](https://github.com/golang-migrate/migrate) (only needed if you want to run migrations manually via `make`).

```bash
git clone https://github.com/tanmaythegreat/VillageCombat_MVC.git
cd VillageCombat_MVC
cp .env.example .env   # edit values as needed

# make sure a local Postgres instance is running and reachable,
# then either let the app apply migrations on boot, or run them yourself:
make migrate-up

go run main.go
```

The server listens on `PORT` (default `8080`) and serves the frontend straight out of `./public`.

## Environment Variables

| Variable | Description | Example |
|---|---|---|
| `DATABASE_URL` | Full Postgres connection string (overrides `DB_*` vars if set) | `postgres://user:pass@localhost:5432/VillageGameDB?sslmode=disable` |
| `DB_USER` | Database user (used by Docker Compose / Makefile) | `admin_tentellam` |
| `DB_PASS` | Database password | — |
| `DB_NAME` | Database name | `VillageGameDB` |
| `JWT_SECRET_KEY` | HMAC signing secret for access tokens — **required**, the app panics without it | — |
| `PORT` | HTTP server port | `8080` |

> ⚠️ The values shown in `.env.example`, `main.go`'s fallback connection string, and `Makefile` are placeholders for local development only — replace all of them (especially `JWT_SECRET_KEY` and database credentials) before deploying anywhere public.

## Database Migrations

Migrations live in `db/migrations` and are managed with `golang-migrate`. They run automatically when the server starts, but can also be controlled manually:

```bash
make help             # list all available migration commands
make migrate-current  # show the currently applied migration version
make migrate-up       # apply all pending migrations
make migrate-down     # roll back the last migration
make migrate-create name=add_something   # scaffold a new up/down migration pair
make db-force-clear version=X            # force the schema_migrations table to a specific version (for a "dirty" DB)
```

## API Reference

### REST Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/register` | Create an account (`username`, `email`, `password_text`). Returns an access + refresh token pair. |
| `POST` | `/login` | Log in with `username` **or** `email`, plus `password_text`. Returns an access + refresh token pair. |
| `POST` | `/refresh` | Exchange a valid `user_id` + refresh token for a new token pair. |
| `GET` | `/*` | Static file server for the `public/` frontend. |

### WebSocket Protocol

Connect to `/ws`. The first message must include a valid `access_token` to authenticate the connection; every subsequent message must include it too. Messages are shaped as:

```json
{ "action": "ACTION_NAME", "message": "<JSON-encoded payload as a string>", "access_token": "..." }
```

| Action | Payload | Description |
|---|---|---|
| `INITIAL_LOAD` | — | Returns the player's placed buildings, trained troops, in-progress construction, and user/profile data |
| `ALL_BUILDING_TROOP_DATA` | — | Returns the full static catalog of buildings/troops/defenses plus per-level stat lookups |
| `CREATE_BUILDING` | `building_id`, `x`, `y`, `use_gems` | Places a new building on the grid |
| `MOVE` | `placed_building_id`, `grid_x`, `grid_y` | Relocates an existing building |
| `UPGRADE_BUILDING` | `placed_building_id`, `use_gems` | Starts an upgrade on a building |
| `REPAIR_BUILDING` | `placed_building_id`, `use_gems` | Repairs a building damaged in battle |
| `CHECK_CONSTRUCTION_WORK` | — | Checks/completes any finished construction, upgrade, or repair tasks |
| `TRAIN_TROOP` | `barrack_placed_building_id`, `troop_id`, `level_from`, `count`, `use_gems` | Queues troop training at a barracks |
| `COLLECT_RESOURCE` | `placed_building_id` | Collects accumulated resources from a single building |
| `COLLECT_ALL` | — | Collects resources from every eligible building at once |
| `ATTACK` | — | Finds a power-matched opponent and starts a live battle |
| `REVENGE` | `<username as message>` | Starts a live battle against a specific named opponent |
| `DEFEND` | — | Placeholder handler for the defending side of a match |
| `BATTLE_HISTORY` | `fought_at`, `to_load` | Paginates the player's past battle history |
| `REPLAY` | `<battle id as message>` | Streams back a recorded battle for replay |
| `LOGOUT` | — | Revokes the refresh token and closes the connection |

## Testing

The project has unit tests alongside nearly every controller, model, and auth file, using `go-sqlmock` to test database logic without a real Postgres instance:

```bash
go test ./...
```

## How to Play

Most mechanics are meant to be discoverable: click an empty grid tile to open the building shop, or click an existing building to open its management menu. Buildings under construction show a countdown; defensive buildings show their attack radius when you're dragging something near them.

## Known Issues

- The construction countdown UI doesn't disappear once a building finishes construction.

## Roadmap

A few TODOs called out directly in the code:

- Username validation (uniqueness + disallowed characters) at registration time
- Email verification

## License

No license has been specified for this project yet. All rights are reserved by the author until a license is added.