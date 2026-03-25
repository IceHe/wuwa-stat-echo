# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview
- Monorepo with FastAPI backend (`backend/`) and Vue 3 + Vite frontend (`frontend/`).
- Backend stores Wuthering Waves “echo” tuning logs in PostgreSQL via SQLModel; serves REST + WebSocket; depends on external auth service for token validation/proxying.
- Frontend consumes the REST/WebSocket endpoints directly; backend/Frontend assume same host with different ports (8888 API, 3000 UI by default).

## Key commands
### Backend (run inside `backend/`)
- Start dev server: `cp .env.example .env && uvicorn main:app --reload --host 0.0.0.0 --port 8888`
- Install backend deps with `pip install -r requirements.txt`.
- Env vars: store `DATABASE_URL`, `AUTH_SERVICE_URL`, and `AUTH_SERVICE_TIMEOUT_SECONDS` in `backend/.env` or inject them from the service manager.
- Database schema/init data: `backend/db.sql` (schema), `backend/db_init_data/*.sql` (seed).
- Tests: none present.

### Frontend (run inside `frontend/`)
- Install deps: `npm install`
- Dev server: `npm run dev` (Vite, serves at 0.0.0.0:3000)
- Build: `npm run build`
- Lint: `npm run lint` (runs eslint + oxlint); format: `npm run format`
- Type check: `npm run type-check`

## Architecture snapshot
### Backend
- Entry: `backend/main.py` — FastAPI app with CORS `*`, lifespan initializes `shared.init_tune_stats()`, routers for echo, substat, analysis, db_data, predict, counts, auth.
- Auth: `backend/auth.py` proxies to external service; permissions: view/edit/manage; extracts Bearer/X-Token/query token; caches per-request. Keep `AUTH_SERVICE_URL` reachable.
- DB: `backend/db.py` builds SQLModel engine from `DATABASE_URL` (adds `+psycopg` automatically). Session dependency `get_session()`.
- Models & logic: `backend/model.py` (EchoLog/SubstatLog), `backend/consts.py` (substat bitmasks, templates, EXP tables), `backend/shared.py` (cached tune stats), `backend/util.py` (bit ops), routers under `backend/api/` in repo history but current code mounts from `api.*` modules located alongside `main.py`.
- APIs & docs: see `backend/docs/ARCHITECTURE.md`, `backend/docs/API.md`, `backend/docs/REQUIREMENTS.md`; Swagger at `/docs`; WebSocket `/ws`.

### Frontend
- SPA with routes in `frontend/src/router/index.ts`; views under `src/views/` (Echo, Substat, Analysis, EchoBoard, EchoDcritCount, etc.).
- Shared constants & API host derivation in `frontend/src/stores/constants.ts`: host = current browser hostname, port 8888; HTTP `http://${API_SERV}` and WS `ws://${API_SERV}/ws` must match backend address.
- Builds via Vite; TS configs `tsconfig*.json`; lint/format via ESLint + Oxlint + Prettier.

## Notable coupling / gotchas
- Backend auth is mandatory for most routes; ensure tokens from external auth service (`~/wuwa/auth`) or requests will 401/403.
- Backend dependency list lives in `backend/requirements.txt`.
- DB must exist and be populated; core tables referenced: `wuwa_tune_log`, `wuwa_echo_log`, `wuwa_substat`, `wuwa_substat_value` (see `db_init_data/`).
- Bitmask substat encoding (13 type bits + value tier bits) is pervasive in backend and consumed directly by frontend; do not change without coordinating both sides.
- Frontend assumes backend CORS `*` and same-host port mapping; adjust `constants.ts` if deploying differently.
