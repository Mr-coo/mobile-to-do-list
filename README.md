# Task Manager

A task management app: a Go (Gin + GORM) REST API backed by MySQL 8, a Redis cache for the list endpoint, and an Expo (React Native, runs on web) client.

No starter repo was provided with the brief. I first built a small baseline app (the four oldest commits in `git log`) that deliberately contained the brief's bugs: duplicate titles returned 500, and soft-deleted tasks showed up in the list. The rest of the history fixes those and adds the requested features one commit at a time.

## Prerequisites

- Docker Desktop (Compose v2)
- Node 20+
- Go 1.26+, only to run the backend tests locally. `go.mod` says `go 1.26.0` because the latest Gin needs it, and `toolchain go1.26.11` because earlier 1.26 releases have standard-library security fixes missing (`govulncheck ./...` reports none at 1.26.11). An older Go downloads that toolchain automatically.

## Run

Start MySQL, Redis and the API (migrations and seed data run on startup):

```bash
docker compose up --build
```

API: http://localhost:8080/api. MySQL is on host port 3307 and Redis on 6379, both bound to localhost only.

Configuration is optional. Compose reads a `.env` file next to `docker-compose.yml` if one exists, and otherwise uses the defaults in [`.env.example`](.env.example) (database passwords, `CORS_ORIGINS`). To change them, copy `.env.example` to `.env` and edit it. The API connects as the `tasks` user, which only has access to the `tasks` database. Changing the database credentials after the first start needs `docker compose down -v`, because MySQL only creates users on an empty volume.

Start the mobile app (web):

```bash
cd mobile
npm install
npx expo start --web
```

Open http://localhost:8081.

The app reads the API address from `EXPO_PUBLIC_API_URL` (default `http://localhost:8080/api`). To change it, copy `mobile/.env.example` to `mobile/.env`; Expo loads it on start.

- Android emulator: `EXPO_PUBLIC_API_URL=http://10.0.2.2:8080/api`
- Physical device: your machine's LAN IP, e.g. `http://192.168.1.20:8080/api`. The API's CORS allow-list (`CORS_ORIGINS`) only matters for the web build.

## Tests

Backend unit tests (integration tests skip without a database):

```bash
cd backend
go test ./...
```

Backend integration tests (real MySQL):

```bash
docker compose up -d mysql
cd backend
TEST_MYSQL_DSN='root:root@tcp(localhost:3307)/tasks_test?parseTime=true&multiStatements=true&loc=UTC' go test -p 1 ./...
```

PowerShell:

```powershell
$env:TEST_MYSQL_DSN='root:root@tcp(localhost:3307)/tasks_test?parseTime=true&multiStatements=true&loc=UTC'; go test -p 1 ./...
```

The tests use the MySQL root account because they need the separate `tasks_test` database; replace `root:root` if you changed `MYSQL_ROOT_PASSWORD`. `-p 1` is needed because all packages share one test database. Without `TEST_MYSQL_DSN` the repository tests report SKIP.

Mobile tests and type check:

```bash
cd mobile
npm test
npx tsc --noEmit
```

## API reference

Swagger UI: http://localhost:8080/swagger/index.html

Regenerate the docs after changing annotations (do not run `swag fmt`):

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
cd backend
swag init -g cmd/api/main.go -o docs --parseInternal
```

| Method | Path | Description | Success |
| --- | --- | --- | --- |
| POST | `/api/tasks` | Create a task (`title` required) | 201 |
| GET | `/api/tasks` | List with filters, sort, pagination | 200 |
| GET | `/api/tasks/:id` | Get one task | 200 |
| PUT | `/api/tasks/:id` | Update a task | 200 |
| DELETE | `/api/tasks/:id` | Soft delete | 204 |

Task fields: `id`, `title` (max 255), `description`, `status` (`todo`, `in_progress`, `done`; default `todo`), `assignee` (max 100), `created_at`, `updated_at`.

List parameters for `GET /api/tasks`:

| Param | Default | Rule |
| --- | --- | --- |
| `status` | none | `todo`, `in_progress` or `done` |
| `keyword` | none | matches title or description; max 100 characters |
| `assignee` | none | exact match, case-insensitive; max 100 characters |
| `page` | `1` | integer >= 1 |
| `limit` | `10` | 1 to 100 |
| `sort` | `-created_at` | `created_at`, `updated_at`, `title` or `status`; prefix `-` for descending |

List response: `{"data": [...], "meta": {"page", "limit", "total", "total_pages"}}`.

Errors always use one shape:

```json
{"error": {"code": "VALIDATION_ERROR", "message": "Invalid request", "details": {"limit": "must be between 1 and 100"}}}
```

| Code | HTTP | When |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | bad body, id or query parameter; `details` names each field |
| `NOT_FOUND` | 404 | unknown or deleted task, unknown route |
| `METHOD_NOT_ALLOWED` | 405 | wrong method on a known route |
| `DUPLICATE_TITLE` | 409 | another active task has this title |
| `PAYLOAD_TOO_LARGE` | 413 | request body over 1 MB |
| `INTERNAL_ERROR` | 500 | unexpected failure; the message never contains internal details |

Examples:

```bash
# create
curl -s -X POST localhost:8080/api/tasks -H 'Content-Type: application/json' \
  -d '{"title":"Buy milk","assignee":"Ana"}'

# filtered list, with the cache header
curl -si 'localhost:8080/api/tasks?status=todo&keyword=milk&sort=title&limit=5' | grep -i -e '^HTTP' -e '^X-Cache'

# update (PUT replaces the task: status is required, omitted description/assignee are cleared)
curl -s -X PUT localhost:8080/api/tasks/1 -H 'Content-Type: application/json' \
  -d '{"title":"Buy oat milk","status":"done"}'

# soft delete
curl -si -X DELETE localhost:8080/api/tasks/1 | head -1
```

Calling the list endpoint twice with the same parameters returns `X-Cache: MISS`, then `X-Cache: HIT`.

## Brief to where it's done

| Brief item | Commits | Tests |
| --- | --- | --- |
| Backend: filtering, sort, pagination | `feat: add task list filtering, sorting and pagination`, `fix: escape backslash in keyword search` | `backend/internal/task/query_test.go`, `repository_test.go` (`TestList_Search`), `handler_test.go` (`TestListTasks_Meta`) |
| Backend: PUT update | `feat: add task update endpoint` | `service_test.go` (`TestUpdate_*`), `handler_test.go` (`TestUpdateTask_*`), `repository_test.go` |
| Backend: soft delete | `feat: add soft delete endpoint` | `service_test.go` (`TestDelete_*`), `handler_test.go` (`TestDeleteTask_*`), `repository_test.go` (`TestDelete_*`) |
| Backend: consistent errors | `feat: return consistent error responses`, `refactor: use json tag names in validation details`, `fix: validate title and assignee length on create` | `backend/internal/httpx/errors_test.go`, `backend/internal/apperr/binding_test.go`, `handler_test.go` |
| Redis cache (60s, invalidate on writes, params in key) | `feat: cache task list in redis`, `fix: fail fast when redis is unavailable` | `backend/internal/task/cache_test.go` |
| API docs | `feat: add swagger api docs` | n/a (generated) |
| Frontend: search, status filter, pagination | `feat: add search, status filter and pagination to task list` | `mobile/src/components/__tests__/SearchBar.test.tsx`, `Pagination.test.tsx`, `TaskList.test.tsx` |
| Frontend: edit modal, loading state | `feat: add task edit modal with delete` | `mobile/src/screens/__tests__/TasksScreen.test.tsx`, `TaskList.test.tsx` (spinner) |
| Bug: duplicate title returns 500, should be 409 | `fix: return 409 for duplicate task titles` | `handler_test.go` (`TestCreateTask_DuplicateTitle_Returns409`), `repository_test.go` (`TestCreate_DuplicateTitle_ReturnsErrDuplicateTitle`) |
| Bug: list not refreshed after update | Prevented by design: no edit feature existed in the baseline, and `useUpdateTask` invalidates the `['tasks']` query | `TasksScreen.test.tsx` ("refreshes the list after saving an edit"), `backend/internal/task/cache_test.go` (`TestUpdate_InvalidatesListCache`) |
| Bug: soft-deleted tasks visible | `fix: hide soft-deleted tasks`, `refactor: skip soft-deleted tasks in fake repo` | `repository_test.go` (`TestList_ExcludesDeleted`, `TestGetByID_Deleted_ReturnsErrNotFound`) |
| DB migration | `feat: add project setup, compose and tasks migrations` | `backend/internal/database/database_test.go` |
| Testing: update, search, cache invalidation, a frontend component | Each feature commit includes its tests | Update: `service_test.go` (`TestUpdate_*`), `handler_test.go` (`TestUpdateTask_*`). Search: `query_test.go`, `repository_test.go` (`TestList_Search`). Cache invalidation: `cache_test.go` (`TestUpdate_InvalidatesListCache`, `TestDelete_InvalidatesListCache`). Components: `SearchBar.test.tsx`, `TaskList.test.tsx`, `Pagination.test.tsx` |

Baseline (existing app) commits: `feat: add project setup, compose and tasks migrations`, `refactor: move test db helper and trim cors origins`, `feat: add task create, list and get endpoints`, `feat: add mobile task list and create form`.

## Assumptions

- No starter repo was provided, so the "existing app" is the baseline described at the top, built first with the brief's bugs in it.
- PUT replaces the whole task: `title` and `status` are required, and an omitted `description` or `assignee` is cleared.
- Titles are unique among active tasks only. Deleting a task frees its title for reuse.
- A soft-deleted task behaves as if it does not exist: GET, PUT and DELETE return 404.
- New tasks default to `todo`.
- Keyword search is case-insensitive and matches the title or description. The assignee filter is an exact, case-insensitive match.
- `sort=status` follows the workflow order `todo`, `in_progress`, `done`, because the column is an ENUM, not alphabetical order.
- Asking for a page past the end returns an empty `data` with the real `meta`. The app then moves back to the last page.

## Project layout

```text
backend/
  cmd/api/              entry point: config, MySQL, migrations, Redis, router
  internal/task/        handler -> service -> repository, list query, Redis cache
  internal/apperr/      domain errors and validation messages
  internal/httpx/       the shared error response
  internal/server/      router, middleware, HTTP server settings
  internal/database/    connection and migrations (dbtest: integration test helper)
  migrations/           SQL migrations, embedded in the binary
  docs/                 generated Swagger files
mobile/
  src/api/              fetch client and task endpoints
  src/hooks/            React Query hooks
  src/components/       list, item, search, status filter, pagination, form modal
  src/screens/          TasksScreen, which wires them together
docker/mysql/init.sql   creates the test database
```

## Design decisions

- Layered backend: `handler` (HTTP, binding) -> `service` (rules, cache) -> `repository` (GORM). The service depends on `Repository` and `ListCache` interfaces, so unit tests use a fake repository and miniredis (in-process Redis), and only `repository_test.go` needs MySQL.
- GORM for queries, plain SQL files for schema. No `AutoMigrate`; `backend/migrations/*.sql` run through golang-migrate on API start. Migration 000002 seeds sample data, including four soft-deleted rows.
- Unique titles among active tasks only, case- and accent-insensitive (column collation `utf8mb4_0900_ai_ci`). MySQL has no partial unique index, so migration 000003 adds a stored generated column `active_title` (the title when `deleted_at IS NULL`, else NULL) with a unique index. NULLs do not collide, so deleting a task frees its title. GORM's `TranslateError` turns the duplicate-key error into `DUPLICATE_TITLE`.
- Cache key: `tasks:list:` plus the normalised query (defaults applied, keys sorted, empty values omitted), so `/tasks` and `/tasks?page=1&limit=10` share an entry. TTL 60s.
- Invalidation: after every successful create, update or delete, `SCAN tasks:list:*` then `DEL`. SCAN rather than KEYS so Redis is never blocked. Known trade-off: a list read that races a write can cache a stale page for up to 60s. For a larger system a version-counter key (`tasks:list:v<N>:...`, bump N on write) avoids the scan and the race.
- Redis failure: errors are logged and the request falls back to MySQL. The client uses 200ms timeouts and no retries, so an outage costs about 0.4s per request (one failed read plus one failed write) instead of seconds.
- Search: `LIKE` on title and description with `\`, `%` and `_` escaped so user input matches literally.
- Sort: the column comes from a Go whitelist, never from the request, and `id` is added as a tie-breaker so pages do not shift between requests.
- Frontend: React Query for server state; search is debounced 300ms; changing search or status resets to page 1; saving or deleting invalidates `['tasks']`.

## Example generated SQL

Captured with GORM's debug logger for `GET /api/tasks?status=todo&keyword=milk&page=2&limit=5&sort=title` (the logging change was temporary and is not committed):

```sql
SELECT count(*) FROM `tasks`
WHERE status = 'todo' AND ((title LIKE '%milk%' OR description LIKE '%milk%')) AND `tasks`.`deleted_at` IS NULL;

SELECT * FROM `tasks`
WHERE status = 'todo' AND ((title LIKE '%milk%' OR description LIKE '%milk%')) AND `tasks`.`deleted_at` IS NULL
ORDER BY title ASC,id ASC LIMIT 5 OFFSET 5;
```

(Line breaks added for reading; GORM logs each statement on one line.) The `deleted_at IS NULL` condition comes from GORM's soft-delete support on the `gorm.DeletedAt` field.

## Security notes

- SQL: every value is a bound parameter; the sort column comes from a whitelist and `LIKE` wildcards are escaped.
- Limits: field lengths are validated, request bodies are capped at 1 MB, search filters at 100 characters, and the HTTP server has read, write and idle timeouts. Redis is capped at 64 MB with LRU eviction, so distinct searches cannot fill it.
- Errors: unexpected failures return a generic 500; details only go to the server log.
- Containers: the API runs as a non-root user; MySQL and Redis listen on localhost only. The API port stays on all interfaces so a phone on the same network can reach it.
- Not done, because the brief left them out: authentication, rate limiting and HTTPS. Swagger UI is also served in release mode.
- Migrations need `multiStatements=true`, so the API's connection allows multi-statement queries. In production, migrations would run as a separate step with their own user, and the API user would only get `SELECT`, `INSERT` and `UPDATE`.
- `npm audit` reports advisories in `picomatch`, `minimatch` and `glob`. These come from Expo CLI, Metro and Jest, which run on the developer machine and are not part of the app bundle. `npm audit fix --force` would downgrade Expo, so it is not applied.

## With more time

- Authentication and per-user tasks.
- Rate limiting.
- FULLTEXT index for keyword search; `LIKE '%x%'` cannot use an index.
- Cursor (keyset) pagination instead of offset.
- Version-counter cache invalidation.
- End-to-end tests against the running stack (API plus the web client).
- CI that runs the unit, integration and mobile tests.
