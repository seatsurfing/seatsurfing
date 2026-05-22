# AGENTS.md — AI Agent Guardrails for Seatsurfing

This document defines coding patterns, conventions, and guardrails for AI agents generating specs and code in this repository. All generated code **must** follow these rules. Deviations will break consistency with the existing codebase.

---

## Repository Overview

Seatsurfing is a desk booking / hot-desking web application. The repo is a monorepo with:

| Directory | Language | Purpose |
|-----------|----------|---------|
| `server/` | Go 1.25+ | Backend API server (PostgreSQL, Gorilla Mux) |
| `ui/` | TypeScript / React 19 / Next.js 16 | Frontend SPA (static export, served under `/ui`) |
| `e2e/` | TypeScript / Playwright | End-to-end browser tests |
| `specs/` | Markdown | Feature specifications |
| `healthcheck/` | Go | Container health check binary |

---

## General Rules

- **Conventional commits** are mandatory. Format: `type(scope): description` (max 150 chars). Required scopes: `booking-ui`, `admin-ui`, `server`, `deps`, `deps-dev`, `i18n`, `main`. Scope is optional only for `ci`, `refactor`, `test` types.
- **Tests are required.** New backend logic needs Go unit tests (both positive and negative). New frontend features need E2E tests. Modified functionality requires updated tests.
- **No third-party libraries** without justification. The backend uses a deliberately small dependency set.
- **Multi-tenant by design.** All data is scoped to an organization. Every database query must filter by `organization_id`. Never leak data across organizations.

---

## Go Backend (`server/`)

### Project Structure

```
server/
  main.go                  # Entry point — sequential initialization
  config/config.go         # Singleton config from env vars (sync.Once)
  repository/              # Database layer — one file per entity
  router/                  # HTTP handlers — one file per entity
  api/                     # Plugin interface definitions
  plugin/                  # Plugin loader (.so files)
  util/                    # Validation, encryption, email, formatting
  testutil/testutil.go     # Shared test helpers
  res/                     # Email templates (JSON, per language)
```

### Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| Repository file | `<entity>-repository.go` | `booking-repository.go` |
| Router file | `<entity>-router.go` | `booking-router.go` |
| Repository test | `<entity>-repository_test.go` in `repository/test/` | `booking-repository_test.go` |
| Router test | `<entity>-router_test.go` in `router/test/` | `booking-router_test.go` |
| Struct type | PascalCase, singular | `Booking`, `User`, `Space` |
| Repository singleton | `Get<Entity>Repository()` | `GetBookingRepository()` |
| Test function | `Test<Entity><Scenario>` | `TestBookingsCRUD` |
| Test helper | `CreateTest<Entity>(...)` | `CreateTestOrg("test.com")` |
| Assertion helper | `CheckTest<Type>(t, expected, actual)` | `CheckTestString(t, "foo", val)` |

### Singleton Pattern

All repositories and the config use `sync.Once` for thread-safe lazy initialization. Follow this pattern exactly:

```go
var bookingRepository *BookingRepository
var bookingRepositoryOnce sync.Once

func GetBookingRepository() *BookingRepository {
    bookingRepositoryOnce.Do(func() {
        bookingRepository = &BookingRepository{}
        _, err := GetDatabase().DB().Exec(`
            CREATE TABLE IF NOT EXISTS bookings (...)
        `)
        if err != nil {
            log.Println(err)
        }
    })
    return bookingRepository
}
```

### Repository Pattern

- One file per entity in `server/repository/`.
- The repository struct is a stateless receiver for database methods.
- Standard CRUD methods: `Create(e *Entity) error`, `GetOne(id string) (*Entity, error)`, `Update(e *Entity) error`, `Delete(e *Entity) error`.
- All SQL queries use **parameterized placeholders** (`$1`, `$2`, ...). Never concatenate user input into SQL.
- All `WHERE` clauses must include `organization_id` filtering (multi-tenancy).
- Emails are always stored and compared in lowercase: `LOWER(email) = $1` and `strings.ToLower(email)`.
- Use `uuid_generate_v4()` as column default for IDs. The ID is returned via `RETURNING id` and assigned after `Create`.
- Multi-row queries: iterate with `rows.Next()`, always `defer rows.Close()`.
- Nullable types use custom wrappers: `NullString`, `NullTime`, `NullUUID` with `CheckNullString()` converters.

### Schema Migrations

- All schema changes go in the entity's `RunSchemaUpgrade(curVersion, targetVersion int)` method.
- The orchestrator is `repository/db-updates.go`.
- Use sequential version checks: `if curVersion < N { ... }`. Never reuse version numbers.
- Use `ADD COLUMN IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS` for idempotency.
- Always add `ALTER TABLE` statements guarded by the version check; never modify existing version blocks.

### Router / HTTP Handler Pattern

- One file per entity in `server/router/`.
- Router struct implements the `api.Route` interface with `SetupRoutes(s *mux.Router)`.
- Route registration is done centrally in `app/app.go` via `PathPrefix` subrouters.
- **Handler signature**: `func (router *EntityRouter) handlerName(w http.ResponseWriter, r *http.Request)`.

#### Request Handling

```go
func (router *BookingRouter) create(w http.ResponseWriter, r *http.Request) {
    // 1. Extract authenticated user from JWT context
    requestUser := GetRequestUser(r)

    // 2. Authorization check — return early on failure
    if !CanSpaceAdminOrg(requestUser, requestUser.OrganizationID) {
        SendForbidden(w)
        return
    }

    // 3. Parse and validate request body
    m := &CreateBookingRequest{}
    if UnmarshalValidateBody(r, m) != nil {
        SendBadRequest(w)
        return
    }

    // 4. Business logic + repository calls

    // 5. Send response
    SendCreated(w, id)
}
```

#### Request DTOs

- Use `validate` struct tags for input validation: `validate:"required"`, `validate:"omitempty,max=256"`.
- JSON field names are camelCase: `json:"spaceId"`.
- Use `UnmarshalValidateBody(r, &m)` for automatic JSON parsing + validation.
- Use `UnmarshalBody(r, &m)` when validation is not needed.

#### Response Functions

Always use the existing response helpers. Never write raw status codes:

| Function | HTTP Status | Use When |
|----------|------------|----------|
| `SendJSON(w, &obj)` | 200 | Returning data |
| `SendCreated(w, id)` | 201 | Entity created (sets `X-Object-ID` header) |
| `SendUpdated(w)` | 204 | Entity updated |
| `SendBadRequest(w)` | 400 | Invalid input |
| `SendBadRequestCode(w, code)` | 400 | Invalid input with error code in `X-Error-Code` |
| `SendUnauthorized(w)` | 401 | Auth failure |
| `SendForbidden(w)` | 403 | Permission denied |
| `SendNotFound(w)` | 404 | Entity not found |
| `SendAlreadyExists(w)` | 409 | Duplicate/conflict |
| `SendTooManyRequests(w)` | 429 | Rate limit hit |
| `SendInternalServerError(w)` | 500 | Unexpected error |

#### Custom Error Codes

Application-specific error codes are returned via the `X-Error-Code` response header. Define new codes as package-level `var` in `routes.go`, grouped by domain with a numeric prefix:

```
1xxx = Booking errors
2xxx = Report errors
3xxx = User errors
4xxx = Group errors
5xxx = Password errors
6xxx = Auth provider errors
```

### Authentication & Authorization

- JWT authentication uses RS512 signing. Claims include `UserID`, `SessionID`, `Email`, `Role`.
- User roles are integer constants: `UserRoleUser (0)`, `UserRoleSpaceAdmin (10)`, `UserRoleOrgAdmin (20)`, `UserRoleServiceAccountRO (21)`, `UserRoleServiceAccountRW (22)`, `UserRoleSuperAdmin (90)`.
- Use `GetRequestUser(r)` to get the authenticated user from context.
- Authorization checks use `Can*` helper functions in `permissions.go`.
- Service accounts support both Basic Auth and Bearer token auth.
- Public/unauthenticated routes are whitelisted in `unauthorized-routes.go`.

### Middleware Stack

The middleware order is: `SecurityHeaderMiddleware` → `VerifyAuthMiddleware` → `RateLimiterMiddleware`. Do not change this order. New middleware should be added with care.

### Configuration

All configuration is via environment variables, read once at startup in `config/config.go`. The `Config` struct uses `sync.Once`. Helper methods:
- `getEnv(key, defaultValue string) string`
- `getEnvInt(key string, defaultValue int) int`
- `getEnvBool(key string, defaultValue bool) bool`

### Encryption & Security

- AES-256-GCM for symmetric encryption (`util/encryption.go`). Requires 32-byte `CRYPT_KEY`.
- SHA-256 hex for irreversible token hashing (API tokens).
- bcrypt for password hashing.
- Never store secrets in plaintext. Use `EncryptString()` / `DecryptString()` for reversible storage, SHA-256 for lookup hashes.

### Caching

Dual-mode caching via `repository/cache.go`: in-memory (freecache, 10 MB) or Valkey (Redis-compatible). Settings use a 5-minute TTL. Check `config.GetConfig().CacheType` before choosing path.

### Plugin System

Plugins are `.so` shared libraries loaded from `plugins/` directory. They implement the `api.SeatsurfingPlugin` interface with lifecycle hooks (`OnInit`, `OnTimer`, `OnUserCreated`, `OnBookingCreated`, etc.). Do not modify the plugin interface without considering backward compatibility.

### Dot Imports

The codebase uses dot imports (`. "package"`) extensively for internal packages (`repository`, `router`, `util`, `config`, `api`). Follow this convention in new files within the same packages.

---

## TypeScript Frontend (`ui/`)

### Project Setup

- **Framework**: Next.js 16 with static export (`output: "export"`, `basePath: "/ui"`).
- **Styling**: Bootstrap 5 + React Bootstrap. CSS files in `src/styles/`, one per component/feature. No CSS-in-JS or Tailwind.
- **State management**: Local component state only. No Redux, no Context API for global state.
- **Path alias**: `@/*` maps to `./src/*`.

### Component Patterns

- **Class components** are the dominant pattern. New components should match the existing style.
- Props and state use TypeScript interfaces named `Props` and `State`.
- Data is fetched in `componentDidMount` (class) or `useEffect` (functional).
- Instance variables (not state) cache fetched data: `this.data: User[] = []`.
- Event handlers use arrow function class properties for auto-binding.

```tsx
interface Props {
  router: NextRouter;
  t: TranslationFunc;
}

interface State {
  loading: boolean;
}

class MyPage extends React.Component<Props, State> {
  data: Entity[] = [];

  componentDidMount = () => {
    if (!Ajax.hasAccessToken()) {
      RedirectUtil.toLogin(this.props.router);
      return;
    }
    this.loadData();
  };

  loadData = () => {
    Entity.list().then((list) => {
      this.data = list;
      this.setState({ loading: false });
    });
  };
}

export default withTranslation(MyPage as any);
```

### Internationalization (i18n)

- Library: `next-export-i18n`.
- Master language file: `i18n/translations.en-GB.json`. All new keys go here first.
- Run `./ui/add-missing-translations.sh` to propagate new keys to other language files.
- Class components: wrap with `withTranslation()` HOC, access `this.props.t("key")`.
- Functional components: `const { t } = useTranslation()`.
- Utility code: `Formatting.t("key")`.
- All user-visible strings must be translation keys, never hardcoded text.

### API Communication

- All HTTP calls go through the static `Ajax` class in `util/Ajax.ts`.
- Entity types have static API methods: `Entity.list()`, `Entity.getSelf()`, `Entity.create()`, etc.
- Entities extend a base `Entity` class with `serialize()` / `deserialize()` methods.
- Promises use `.then()` / `.catch()` chains (not async/await in class components).
- Token refresh is handled automatically with a mutex lock in `Ajax`.

### Routing & Auth Protection

- File-based routing under `src/pages/`.
- Admin pages live under `pages/admin/<feature>/`.
- Route protection is checked in `componentDidMount` via `Ajax.hasAccessToken()` → `RedirectUtil.toLogin()`.
- All routes use trailing slashes (`trailingSlash: true` in next.config.js).

### File Naming

| Element | Convention | Example |
|---------|-----------|---------|
| Page component | feature name, lowercase | `bookings.tsx`, `search.tsx` |
| Admin page | `pages/admin/<feature>/index.tsx` or `[id].tsx` | `pages/admin/users/[id].tsx` |
| Reusable component | PascalCase | `NavBar.tsx`, `SaveButton.tsx` |
| CSS file | PascalCase or feature name in `src/styles/` | `NavBar.css`, `Booking.css` |
| Entity/type | PascalCase in `src/types/` | `User.ts`, `Booking.ts` |
| Utility | PascalCase in `src/util/` | `Ajax.ts`, `DateUtil.ts` |

---

## Testing

### Go Backend Tests

#### Structure

- Repository tests: `server/repository/test/<entity>-repository_test.go`
- Router tests: `server/router/test/<entity>-router_test.go`
- Both share a `test_test.go` with `TestMain` that calls `testutil.TestRunner(m)`.

#### Test Setup

- `TestRunner(m)` initializes the database, drops all tables, re-creates schema, and runs tests.
- Every test function starts with `ClearTestDB()` (truncates all tables).
- Environment: `POSTGRES_URL` defaults to `postgres://postgres:root@localhost/seatsurfing_test?sslmode=disable`.
- `CRYPT_KEY` must be set to a 32-byte value in `TestMain`.
- `MOCK_SENDMAIL=1` prevents real email sending.

#### Test Helpers (in `testutil/`)

| Helper | Purpose |
|--------|---------|
| `CreateTestOrg(domain)` | Creates an org with domain |
| `CreateTestUserInOrg(org)` | Creates a regular user in org |
| `CreateTestUserOrgAdmin(org)` | Creates an org admin |
| `CreateTestUserSuperAdmin()` | Creates a super admin |
| `CreateTestLocationAndSpace(org)` | Creates a location + space |
| `CreateTestBooking9To5(user, space, dayOffset)` | Creates a 9-5 booking |
| `CreateTestGroup(org, user)` | Creates a group with member |
| `NewHTTPRequest(method, url, userID, body)` | Creates authenticated HTTP request |
| `ExecuteTestRequest(req)` | Executes request against the router |
| `ClearTestDB()` | Truncates all tables |
| `LoginTestUser(userID)` | Creates test login response |

#### Assertion Helpers

Do **not** use third-party assertion libraries. Use the custom helpers:

```go
CheckTestResponseCode(t, http.StatusOK, res.Code)
CheckTestString(t, "expected", actual)
CheckTestInt(t, 42, actual)
CheckTestBool(t, true, actual)
CheckTestIsNil(t, obj)
CheckStringNotEmpty(t, s)
```

#### Router Test Pattern

```go
func TestEntityCRUD(t *testing.T) {
    ClearTestDB()
    org := CreateTestOrg("test.com")
    user := CreateTestUserOrgAdmin(org)
    login := LoginTestUser(user.ID)

    // Create
    payload := `{"name": "Test"}`
    req := NewHTTPRequest("POST", "/entity/", login.UserID, bytes.NewBufferString(payload))
    res := ExecuteTestRequest(req)
    CheckTestResponseCode(t, http.StatusCreated, res.Code)
    id := res.Header().Get("X-Object-Id")

    // Read
    req = NewHTTPRequest("GET", "/entity/"+id, login.UserID, nil)
    res = ExecuteTestRequest(req)
    CheckTestResponseCode(t, http.StatusOK, res.Code)

    // Update
    payload = `{"name": "Updated"}`
    req = NewHTTPRequest("PUT", "/entity/"+id, login.UserID, bytes.NewBufferString(payload))
    res = ExecuteTestRequest(req)
    CheckTestResponseCode(t, http.StatusNoContent, res.Code)

    // Delete
    req = NewHTTPRequest("DELETE", "/entity/"+id, login.UserID, nil)
    res = ExecuteTestRequest(req)
    CheckTestResponseCode(t, http.StatusNoContent, res.Code)
}
```

### Frontend Tests (Vitest)

- Test files: `src/components/__tests__/*.test.tsx` or `src/util/*.test.ts`.
- Framework: Vitest with jsdom environment.
- Mock external dependencies with `vi.mock()` (i18n, react-bootstrap, Ajax).
- Use `describe` / `it` / `expect` from vitest.
- Use `vi.fn()` for mock functions.

### E2E Tests (Playwright)

- Test files: `e2e/tests/*.spec.ts`.
- Config: sequential execution (`workers: 1`), no parallelism, 2 retries on CI.
- Always suppress the MFA encouragement modal in `beforeEach`:
  ```ts
  await page.addInitScript(() => {
    window.localStorage.setItem("mfaEncouragementDismissed", "1");
  });
  ```
- Use `login(page, email, password)` helper from `e2e/util/helper.ts`.
- **Selector preference**: `getByRole()` > `getByText()` > `getByLabel()`. Avoid CSS class selectors.
- Default test credentials: `admin@seatsurfing.local` / `Sea!surf1ng`.

### Running Tests

```bash
# Backend unit tests
cd server && ./test.sh

# Frontend unit tests
cd ui && npx vitest run

# E2E tests (requires built UI + running server)
cd e2e && npx playwright test
```

---

## Feature Specifications (`specs/`)

### Spec Format

Feature specs follow this structure:

1. **Overview** — One-paragraph summary of the feature.
2. **Goals** — Bulleted list of what the feature achieves.
3. **Non-Goals** — Explicit list of what is out of scope.
4. **Product Decisions** — Key design choices made upfront.
5. **Existing System Constraints** — Relevant existing code, APIs, and data models that the feature must work with. Include code snippets and table schemas.
6. **Data Model Changes** — New tables, columns, or schema migrations.
7. **API Changes** — New or modified endpoints with request/response formats.
8. **UI Changes** — Frontend modifications.
9. **Integration Points** — How the feature connects with existing systems (auth, plugins, etc.).

### Spec Rules

- Always document existing system constraints before proposing changes.
- Reference actual code paths and function names from the codebase.
- Include SQL schema snippets for data model changes.
- Specify which role(s) can access new endpoints.
- Note multi-tenancy implications (organization scoping).
- Specs live in `specs/` as Markdown files named after the feature: `<feature-name>.md`.

---

## Security Rules

- **SQL injection**: Always use parameterized queries. Never string-concatenate user input into SQL.
- **Multi-tenancy**: Every query must scope to `organization_id`. Verify org isolation in both repository and router layers.
- **Auth checks**: Every non-public endpoint must verify the user's role before processing. Use `GetRequestUser(r)` and `Can*` helpers.
- **Input validation**: Validate at the boundary using `validate` struct tags and `UnmarshalValidateBody`. Use the regex validators in `util/validation.go` for emails, domains, passwords, UUIDs, colors.
- **Secrets**: Use `EncryptString()` for reversible storage, SHA-256 hex for token lookup hashes, bcrypt for passwords. Never log or return secrets in responses.
- **Rate limiting**: All authenticated endpoints are automatically rate-limited. Public endpoints must consider abuse.
- **CORS & headers**: Security headers are set by `SecurityHeaderMiddleware`. Do not bypass.

---

## Build & Deployment

- **Docker**: Multi-stage build (UI → Server → Healthcheck → distroless base). Non-root user (UID 65532).
- **Build**: `./build.sh` builds the UI with version from `version.txt`.
- **Dev**: `./dev.sh` uses tmux to run UI, server, and database in parallel.
- **Release**: Automated via release-please from conventional commits.

---

## Do Not

- Do not introduce global state outside the established singleton pattern.
- Do not add new ORM or query builder dependencies. Raw SQL with `database/sql` is the standard.
- Do not use `context.Context` for passing business data. It is used only for request-scoped auth data (`UserID`, `SessionID`).
- Do not use async/await in class components. Use `.then()` / `.catch()` chains.
- Do not use CSS-in-JS, Tailwind, or styled-components. Use Bootstrap + CSS files.
- Do not introduce Redux, Zustand, or any global state library in the frontend.
- Do not modify existing migration version blocks. Only append new blocks with higher version numbers.
- Do not add routes without corresponding authorization checks.
- Do not skip `ClearTestDB()` at the start of test functions.
- Do not use third-party assertion libraries in Go tests. Use the `CheckTest*` helpers.
- Do not hardcode user-visible strings. Always use i18n translation keys.
