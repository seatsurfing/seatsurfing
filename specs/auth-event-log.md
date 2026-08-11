# Specification: Authentication Event Log

## 1. Overview

This specification describes an **authentication event log** for Seatsurfing. Every login attempt (successful or failed, via built-in password/TOTP/passkey login or via an external Auth Provider / IdP) is recorded in the database together with a stable error code and the raw error detail. Org admins get a new admin UI view to filter and browse these events, enabling them to diagnose login problems — especially IdP failures that today surface only as a generic "Login failed" message.

### 1.1 Goals

- Record every login attempt with: organization, user (if resolvable), attempted email, timestamp, outcome, auth method, auth provider (for IdP logins), a stable error code, the raw error detail (e.g. the IdP's OAuth error response body), and the client's device string (browser/OS).
- Capture failure causes in **all** branches of the IdP/OAuth flow (provider config, code exchange, userinfo fetch, attribute mapping, unknown user, provider mismatch, user limit) that today collapse into generic redirect reasons.
- Provide an org-admin-only REST endpoint to list and filter events, with pagination.
- Provide an admin UI page ("Auth Events") with filters (date range, outcome, method, error code, email) and a detail view.
- Bound table growth via a fixed, globally configured retention period.

### 1.2 Non-Goals

- No IP address storage (privacy decision — only the parsed user-agent/device string is stored).
- No recording of password resets, logouts, session evictions, or service-account/API-token authentication.
- No per-organization retention configuration.
- No change to what end users see on failed logins (responses/redirects stay as they are).
- No per-minute insert rate cap for unauthenticated failure events (growth is bounded by retention; may be revisited).

## 2. Product Decisions

- **Extend the existing `auth_attempts` table** (used today for brute-force ban protection) instead of adding a second table: every login attempt is recorded exactly once, and ban logic and audit log share one source of truth.
- **Login events only** (password, TOTP, passkey, passkey 2FA, OAuth/IdP, Confluence).
- **Device string, no IP**: the user agent is parsed to e.g. "Chrome on Windows" (existing `GetDeviceInfo` logic).
- **Fixed global retention** via env var `AUTH_EVENT_RETENTION_DAYS` (default `90`; `<= 0` disables purging). Purged in batches by the existing 1-minute cleanup ticker.
- Events whose organization cannot be determined (e.g. OAuth login request for a non-existing provider ID) are **not** recorded: they are unattributable noise and would be an unauthenticated write-amplification vector.
- Failed attempts for **unknown email addresses** are now recorded (they were not before), scoped to the organization ID from the login request. Unknown-email failures never count toward user bans (`user_id` is NULL).

## 3. Existing System Constraints

- `auth_attempts` table (`server/repository/auth-attempt-repository.go`): `id uuid, user_id uuid NULL, email VARCHAR, timestamp TIMESTAMP, successful BOOLEAN`; indexes on `user_id` and `email`. Written via `RecordLoginAttempt(user, success)`, which also runs `checkBanUser` (auto-disable after `LOGIN_PROTECTION_MAX_FAILS` failures within `LOGIN_PROTECTION_SLIDING_WINDOW_SECONDS`). No org scoping, no read API, no cleanup.
- Login flows in `server/router/auth-router.go` (`loginPassword`, OAuth `login`/`callback`/`verify`, `getUserInfo`, `createAndSendJWT`), `server/router/passkey-router.go` (discoverable login, passkey 2FA) and `server/router/confluence-router.go`. Successful logins funnel through `createAndSendJWT` except discoverable passkey login, which creates its session inline.
- OAuth failures are flattened by `getRedirectFailedUrl` into reason strings (`provider`, `config`, `userinfo`, `authState`, `login`) on `/ui/login/failed/`; the detailed causes (e.g. `*oauth2.RetrieveError` bodies) are logged at best and otherwise discarded.
- Schema migrations: `server/repository/db-updates.go`, current `targetVersion` 49; new changes go into `RunSchemaUpgrade` blocks guarded by `if curVersion < 50`.
- Org-admin authorization: `CanAdminOrg(GetRequestUser(r), orgID)` in `server/router/routes.go`.

## 4. Data Model Changes

Migration **v50** extends `auth_attempts` (additive, zero-downtime safe):

```sql
ALTER TABLE auth_attempts ADD COLUMN IF NOT EXISTS organization_id uuid NULL;
ALTER TABLE auth_attempts ADD COLUMN IF NOT EXISTS method VARCHAR NOT NULL DEFAULT '';
ALTER TABLE auth_attempts ADD COLUMN IF NOT EXISTS auth_provider_id uuid NULL;
ALTER TABLE auth_attempts ADD COLUMN IF NOT EXISTS error_code VARCHAR NOT NULL DEFAULT '';
ALTER TABLE auth_attempts ADD COLUMN IF NOT EXISTS error_detail VARCHAR NOT NULL DEFAULT '';
ALTER TABLE auth_attempts ADD COLUMN IF NOT EXISTS device VARCHAR NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_auth_attempts_org_ts ON auth_attempts(organization_id, timestamp DESC);

-- backfill org for existing rows
UPDATE auth_attempts a SET organization_id = u.organization_id
FROM users u WHERE a.user_id = u.id AND a.organization_id IS NULL;
```

- `method`: `password`, `totp`, `passkey`, `passkey_2fa`, `oauth`, `confluence`; `''` for legacy rows (rendered as "Unknown").
- `error_detail` is truncated to 2000 characters before insert.
- Rows whose user was deleted before the migration keep `organization_id = NULL`; they are invisible in the org-scoped list and age out via retention.
- Organization deletion deletes the org's auth events.

### 4.1 Error Codes

| Code                           | Meaning                                                                        |
| ------------------------------ | ------------------------------------------------------------------------------ |
| _(empty)_                      | Successful login                                                               |
| `user_not_found`               | No user with the attempted email in the organization                           |
| `password_pending`             | Password change was requested but not yet confirmed                            |
| `no_password_set`              | User has no password hash                                                      |
| `bound_to_auth_provider`       | User must log in via their IdP, not with a password                            |
| `user_disabled`                | User is disabled or banned                                                     |
| `service_account`              | Service accounts cannot use interactive login                                  |
| `wrong_password`               | Password mismatch                                                              |
| `password_update_required`     | Login blocked until password is updated                                        |
| `totp_missing`                 | TOTP code required but not provided                                            |
| `totp_replay`                  | TOTP code was already used                                                     |
| `totp_invalid`                 | TOTP code invalid                                                              |
| `passkey_state_invalid`        | Passkey ceremony state missing/invalid/expired                                 |
| `passkey_assertion_invalid`    | WebAuthn assertion failed validation                                           |
| `passkey_clone_detected`       | Authenticator sign counter regressed (possible cloned passkey)                 |
| `idp_provider_not_found`       | Auth provider does not exist (recorded only when org resolvable)               |
| `idp_config_invalid`           | Provider/org misconfigured (no primary domain, or client secret undecryptable) |
| `idp_state_invalid`            | OAuth state missing/mismatched                                                 |
| `idp_code_exchange_failed`     | Token endpoint rejected the authorization code (IdP response body in detail)   |
| `idp_userinfo_failed`          | Userinfo request failed or returned non-2xx (status + body in detail)          |
| `idp_attribute_mapping_failed` | Email attribute could not be extracted from the userinfo response              |
| `idp_provider_mismatch`        | User is bound to a different auth provider                                     |
| `user_limit_reached`           | Organization's user limit prevents auto-creating the user                      |
| `user_create_failed`           | Auto-creation of the IdP user failed                                           |
| `org_mismatch`                 | User's organization does not match the provider's organization                 |
| `internal_error`               | Unexpected server-side error (DB, auth-state creation, TOTP secret decryption) |
| `confluence_jwt_invalid`       | Atlassian Connect JWT verification failed                                      |

## 5. API Changes

New endpoint `GET /auth-attempt/` — **role: org admin** (or super admin); always scoped to the caller's own organization (no org parameter).

Query parameters:

| Param           | Type                               | Default                         |
| --------------- | ---------------------------------- | ------------------------------- |
| `start` / `end` | RFC3339Nano timestamps             | end = now, start = end − 7 days |
| `user`          | email substring (case-insensitive) | —                               |
| `method`        | method constant                    | —                               |
| `success`       | `true` / `false`                   | both                            |
| `errorCode`     | error code constant                | —                               |
| `limit`         | int, max 100                       | 50                              |
| `offset`        | int                                | 0                               |

Response `200`:

```json
{
  "total": 123,
  "items": [
    {
      "id": "…",
      "userId": "…",
      "email": "user@acme.com",
      "timestamp": "2026-08-11T09:00:00Z",
      "successful": false,
      "method": "oauth",
      "authProviderId": "…",
      "authProviderName": "Keycloak",
      "errorCode": "idp_code_exchange_failed",
      "errorDetail": "oauth2: \"invalid_grant\" …",
      "device": "Chrome on Windows"
    }
  ]
}
```

`403` for non-admins.

## 6. UI Changes

- New admin page **`/admin/auth-events/`** ("Auth Events"), linked in the sidebar for org admins only.
- Filters: date range (start/end pickers), outcome (all/success/failure), method, error code, email substring; filter state mirrored into the URL query.
- Paginated table (50 per page): timestamp, email, method, provider, outcome, translated error code. Row click opens a modal with the localized error explanation, raw `errorDetail`, device string, and user ID.
- Error codes are translated via i18n keys (`autherror.<code>`); raw detail is shown verbatim.

## 7. Integration Points

- **Ban protection**: `checkBanUser` continues to run for exactly the attempt types that trigger it today (password, TOTP, passkey with a resolved user). New event types (IdP failures, unknown email) never affect bans.
- **Cleanup ticker**: `App.onTimerTick` purges events older than `AUTH_EVENT_RETENTION_DAYS` in batches of 100. Retention is always far larger than the ban sliding window, so ban counting is unaffected.
- **Organization deletion**: `OrganizationStore.Delete` deletes the org's auth events.
- **Multi-tenancy**: all list queries filter by `organization_id`; the endpoint derives the org from the authenticated user.
