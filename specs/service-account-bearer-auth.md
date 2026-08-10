# Feature Spec: Service Account Bearer Token Authentication

## Overview

Extend the existing service account authentication mechanism to support `Authorization: Bearer <token>` in addition to the existing `Authorization: Basic <base64>` (HTTP Basic Auth). This makes service accounts usable with clients and integrations that require or strongly prefer Bearer token authentication — most notably enterprise IdPs used with the SCIM server (see `specs/scim.md`), but also custom scripts and third-party tools that cannot easily construct Basic Auth headers.

## Goals

- Each service account user may optionally have a dedicated, long-lived API token.
- The token can be presented as `Authorization: Bearer <token>` on any endpoint where the service account's Basic Auth credentials would currently be accepted.
- The token is generated server-side with cryptographic randomness, returned once, and never retrievable in plaintext again.
- Admins can generate and revoke API tokens for service accounts via the user management interface.
- The existing HTTP Basic Auth mechanism is kept unchanged and continues to work alongside the new Bearer mechanism.
- No changes to the RO/RW permission model: read-only service accounts may only use Bearer tokens on `GET` requests; read-write service accounts may use Bearer tokens on any method.

## Non-Goals

- Bearer token authentication for non-service-account users (regular users, org admins, super admins). They continue to use JWT-based sessions.
- OAuth 2.0 client credential flows.
- Token expiry or rotation schedules (the token is long-lived until explicitly revoked).
- Multiple API tokens per service account in the first version.

## Existing System Constraints

### Service Account Authentication Today

`VerifyAuthMiddleware` in `server/router/routes.go` runs on every request. For non-whitelisted routes it calls two handlers in order:

```
success := handleTokenAuth(w, r) || handleServiceAccountAuth(w, r)
```

`handleTokenAuth` parses a JWT from `Authorization: Bearer <jwt>`. If the value is not a valid JWT it returns `false`.

`handleServiceAccountAuth` calls `r.BasicAuth()`. If the request carries `Authorization: Basic <base64>`, it decodes the credentials and looks up the service account. If the request carries a `Bearer` header (or no header), `r.BasicAuth()` returns `ok=false` and the function returns `false`.

For whitelisted routes, the middleware additionally attempts both handlers when a `Bearer` header is present:

```go
if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
    processedWithAuth = handleTokenAuth(w, r) || handleServiceAccountAuth(w, r)
}
```

### User Roles

Service account roles:

| Constant                   | Value | Allowed methods |
| -------------------------- | ----- | --------------- |
| `UserRoleServiceAccountRO` | 21    | GET only        |
| `UserRoleServiceAccountRW` | 22    | All methods     |

The existing `isServiceAccountRole(role int) bool` helper in `user-router.go` identifies both.

## Data Model Changes

### New Column: `users.api_token`

Add a nullable `VARCHAR` column to the `users` table to hold the SHA-256 hex digest of the raw API token:

```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS api_token VARCHAR NULL;
CREATE UNIQUE INDEX IF NOT EXISTS users_api_token ON users(api_token) WHERE api_token IS NOT NULL;
```

The unique index enables O(1) lookup by token hash and prevents accidental collisions.

This migration must be added to `server/repository/db-updates.go` under a new schema version number.

### Why SHA-256, Not bcrypt or EncryptString

**bcrypt** is ruled out for the same reason it is unsuitable for any lookup-keyed value: each invocation produces a different output (random salt), so the stored hash cannot be used as a database index key. Verifying a presented token would require decrypting or hashing every row in the table — O(n) work per request.

**`EncryptString`** (AES-256-GCM with a random nonce, used in this codebase for OAuth client secrets) is also ruled out. It has the same nonce-randomness problem as bcrypt: the same plaintext produces a different ciphertext on every call, so it cannot serve as an index key either. More importantly, `EncryptString` is reversible — it exists precisely so fields like client secrets can be recovered in plaintext when they must be forwarded to an external service. An API token is never recovered after initial issuance; the server only needs to answer "does the presented token match the stored value?" Storing a recoverable form of the token is a security regression: if both the database and the `CRYPT_KEY` config value are compromised, all tokens become usable by an attacker.

**SHA-256** is correct here because:

1. API tokens are server-generated 32-byte cryptographically random values (256 bits of entropy). Rainbow tables or brute-force are computationally infeasible regardless of hash algorithm.
2. SHA-256 is deterministic, so `sha256hex(presentedToken)` can be compared against a single indexed column in O(1).
3. A compromised database reveals only hashes. Without the original tokens (which only the token holders know), the hashes are useless to an attacker — no `CRYPT_KEY` is involved.

This is the same pattern used by GitHub personal access tokens, Gitea tokens, and similar systems. By contrast, passwords are stored with bcrypt because they are user-chosen with potentially low entropy, making a one-way-but-slow hash the correct defence.

### Updated `User` Struct

Add `ApiToken NullString` to `server/repository/user-repository.go`:

```go
type User struct {
    // ... existing fields ...
    ApiToken NullString  // SHA-256 hex digest; NULL when no token is configured
}
```

Include `api_token` in all relevant `SELECT`, `INSERT`, and `UPDATE` statements in `UserRepository`.

### New Repository Method

Add to `UserRepository`:

```go
// GetByApiToken looks up a service account by its hashed API token.
// Returns nil, nil when no matching user is found.
func (r *UserRepository) GetByApiToken(tokenHash string) (*User, error)
```

Implementation: `SELECT ... FROM users WHERE api_token = $1`.

## Authentication Flow Extension

### Extended `handleServiceAccountAuth`

`handleServiceAccountAuth` in `routes.go` is extended to try Bearer token auth after Basic Auth fails:

```
1. Try r.BasicAuth()
   → If ok: validate credentials as before (unchanged behavior)

2. If Basic Auth fails, inspect the Authorization header:
   → If it doesn't start with "Bearer ": return false
   → If the Bearer value starts with "eyJ": return false
     (JWTs have the eyJ prefix; handleTokenAuth handles those)
   → Compute SHA-256(bearerValue), call GetUserRepository().GetByApiToken(hash)
   → If no user found: return false
   → If user.Role not in {ServiceAccountRO, ServiceAccountRW}: return false
   → If user.Disabled: return false
   → If r.Method != "GET" and user.Role == UserRoleServiceAccountRO: return false
   → Set user ID in request context, call next.ServeHTTP, return true
```

**No other code changes are needed in `routes.go`** because `handleServiceAccountAuth` is already called on both whitelisted and non-whitelisted routes when a Bearer header is present.

### Interaction with `handleTokenAuth`

`handleTokenAuth` is always tried first. For a non-JWT Bearer value it will fail silently (the JWT parser returns an error). `handleServiceAccountAuth` is then called and handles the API token. The order of operations is unchanged.

### Interaction with Whitelisted Routes

For paths in `unauthorizedRoutes` the middleware tries auth when a `Bearer` header is present (but does not reject the request if auth fails). This means service account Bearer tokens work on whitelisted routes too — the user context is populated if the token is valid, and the handler can gate behavior accordingly. The SCIM router relies on this behavior (see `specs/scim.md`).

## API Endpoints

These endpoints are added to `server/router/user-router.go`. Existing route registration in `SetupRoutes` is extended.

### `POST /user/{id}/api-token`

Generate a new API token for a service account user.

**Authorization**: The calling user must be an org admin (`CanAdminOrg`) or super admin for the target user's organization. Service accounts cannot generate their own token.

**Request body**: none.

**Behavior**:

1. Load the target user. If not found or not in the caller's org → 404.
2. If the target user's role is not `UserRoleServiceAccountRO` or `UserRoleServiceAccountRW` → 400 (only service accounts may have API tokens).
3. Generate 32 cryptographically random bytes via `crypto/rand`. Encode as a 64-character lowercase hex string → `rawToken`.
4. Compute `sha256hex(rawToken)` → `tokenHash`.
5. Store `tokenHash` in `users.api_token` for the target user (overwriting any existing token).
6. Return HTTP 200 with body:

```json
{
  "token": "<rawToken>"
}
```

The raw token is returned exactly once. After this call completes, the plaintext is no longer accessible. The admin must copy it before navigating away.

**Note**: Regenerating the token immediately invalidates the previous one. There is no separate "rotate" endpoint.

### `DELETE /user/{id}/api-token`

Revoke the API token for a service account user.

**Authorization**: Same as POST above.

**Behavior**:

1. Load the target user. If not found or not in caller's org → 404.
2. If the target user is not a service account → 400.
3. Set `users.api_token = NULL` for the target user.
4. Return HTTP 204.

### `GET /user/{id}/api-token`

Check whether the service account has a token configured.

**Authorization**: Same as POST above.

**Response**: HTTP 200 with body:

```json
{
  "configured": true
}
```

or `{"configured": false}` when no token is set. The raw token is never returned.

## Admin UI Changes

Extend the service account user detail view in the admin interface with a new **API Token** section, shown only when the selected user has a service account role:

- `API Token` label with a status indicator: `Configured` (green) or `Not configured` (gray).
- `Generate token` button: calls `POST /user/{id}/api-token`, shows the returned raw token in a modal with a copy-to-clipboard button and a warning that the token is shown only once.
- `Revoke token` button (visible only when a token is configured): calls `DELETE /user/{id}/api-token` with a confirmation dialog.

All controls are hidden when `feature_scim` is `false` if the only use-case is SCIM (revisit if Bearer auth is used for other purposes in the future).

## Unit Tests

Test file: `server/router/test/user-router_test.go` (extend existing file).

### Helper Additions in testutil

```go
// CreateTestServiceAccountRW creates a RW service account in the given org.
func CreateTestServiceAccountRW(org *Organization) *User

// GenerateTestApiToken generates an API token for the given user and
// stores the hash in the DB. Returns the raw token.
func GenerateTestApiToken(userID string) string

// NewHTTPRequestBearer wraps NewHTTPRequest but uses the provided raw token
// as a Bearer token instead of a JWT.
func NewHTTPRequestBearer(method, url, rawToken string, body io.Reader) *http.Request
```

### Required Test Cases

#### Token Generation

- `TestApiTokenGenerate` — POST by org admin on a service account returns 200 with a `token` field; subsequent `GET api-token` returns `configured: true`.
- `TestApiTokenGenerateNonServiceAccount` — POST on a regular user → 400.
- `TestApiTokenGenerateForbidden` — POST by a non-admin user → 403.
- `TestApiTokenGenerateOtherOrg` — POST targeting a user in a different org → 404.
- `TestApiTokenRegenerate` — POST twice; old token no longer authenticates; new token authenticates.

#### Token Revocation

- `TestApiTokenRevoke` — DELETE clears token; `GET api-token` returns `configured: false`.
- `TestApiTokenRevokeNotFound` — DELETE on a user without a token → still 204 (idempotent).

#### Bearer Authentication

- `TestServiceAccountBearerAuth` — request with `Authorization: Bearer <validToken>` on a protected endpoint authenticates correctly and returns 200.
- `TestServiceAccountBearerAuthWrongToken` — wrong token value → 401.
- `TestServiceAccountBearerAuthDisabledUser` — service account is disabled → 401.
- `TestServiceAccountBearerAuthROReadRequest` — RO account with Bearer token on GET → 200.
- `TestServiceAccountBearerAuthROWriteRequest` — RO account with Bearer token on POST → 401.
- `TestServiceAccountBearerAuthRWWriteRequest` — RW account with Bearer token on POST → 200 (or 201/204 depending on endpoint).
- `TestServiceAccountBearerAuthJwtNotTreatedAsApiToken` — a valid JWT in the Bearer header is handled by `handleTokenAuth`, not misidentified as an API token.
- `TestServiceAccountBasicAuthStillWorks` — existing Basic Auth service account requests continue to authenticate after the extension.

## Migration Checklist

1. `server/repository/db-updates.go` — add `api_token` column and unique index migration.
2. `server/repository/user-repository.go` — add `ApiToken NullString` to `User` struct; add `GetByApiToken` method; include `api_token` in SELECT/INSERT/UPDATE queries.
3. `server/router/routes.go` — extend `handleServiceAccountAuth` to handle non-JWT Bearer tokens via SHA-256 token hash lookup.
4. `server/router/user-router.go` — add `POST`, `GET`, `DELETE /user/{id}/api-token` handlers and register routes in `SetupRoutes`.
5. `server/router/test/user-router_test.go` — add test cases listed above.
6. `server/testutil/testutil.go` — add `CreateTestServiceAccountRW`, `GenerateTestApiToken`, `NewHTTPRequestBearer` helpers.
