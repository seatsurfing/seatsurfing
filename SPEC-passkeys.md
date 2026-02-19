# Specification: Passkeys Support for Seatsurfing

## 1. Overview

This specification describes the addition of **WebAuthn/Passkeys** as a passwordless login method for Seatsurfing users who authenticate via username/password. Passkeys provide a phishing-resistant, biometric-backed authentication mechanism that is both more secure and more convenient than traditional passwords.

### 1.1 Goals

- Allow username/password users to register one or more passkeys on their account.
- Allow those users to log in using only a passkey (passwordless), without entering username or password.
- When a user with passkeys falls back to username/password login, use the passkey as the second factor (replacing TOTP).
- Do **not** offer passkeys to users who authenticate via an external Auth Provider (IdP).
- Each passkey is identified by a unique, user-given name (e.g., "MacBook Touch ID", "YubiKey").

### 1.2 Non-Goals

- Passkeys are not offered to external Auth Provider (IdP) users.
- Passkeys do not replace admin-facing authentication flows (service accounts use HTTP Basic Auth, unchanged).
- This specification does not cover passkey-based account recovery.

---

## 2. Terminology

| Term | Definition |
|------|-----------|
| **Passkey** | A FIDO2/WebAuthn discoverable credential stored on the user's device or in a cloud-synced keychain. |
| **Relying Party (RP)** | The Seatsurfing server, identified by its RP ID (domain) and RP name. |
| **Discoverable credential** | A WebAuthn credential that can initiate authentication without the user first providing a username (resident key). |
| **User verification** | Biometric/PIN check performed locally by the authenticator during WebAuthn ceremonies. |
| **Ceremony** | A WebAuthn registration or authentication handshake between browser and server. |

---

## 3. Interaction with Existing Auth Mechanisms

### 3.1 Passkeys vs. Username/Password

Passkeys offer an **alternative** login path. Users can choose either:

- **Passwordless flow**: Click "Sign in with passkey" → browser shows passkey selection → authenticate with biometrics/PIN → logged in. No username, password, or TOTP required.
- **Password flow** (unchanged for users without passkeys): Enter username + password → (optional TOTP) → logged in.

### 3.2 Passkeys vs. TOTP (Second-Factor Interaction)

When a user has **both** passkeys and TOTP configured and logs in via **username/password** (not via the passwordless flow):

1. After successful password verification, the server detects the user has registered passkeys.
2. The server responds with a **passkey challenge** instead of the TOTP prompt.
3. The client initiates a WebAuthn `get()` ceremony.
4. If the passkey assertion succeeds → login complete.
5. If the user cannot complete the passkey challenge (e.g., device not available), a **fallback link** is offered to use a TOTP code instead (only if TOTP is also configured).

When a user has **only passkeys** (no TOTP) and logs in via username/password:

1. After successful password verification, the server responds with a passkey challenge.
2. The user must complete the passkey challenge to log in.
3. No TOTP fallback is available (TOTP is not configured).

When a user has **only TOTP** (no passkeys) and logs in via username/password:

1. Existing behavior is unchanged — TOTP code is prompted.

### 3.3 `enforce_totp` Setting Interaction

The existing `enforce_totp` organization setting requires users to set up a second factor. With passkeys:

- Having **at least one passkey** registered satisfies the `enforce_totp` requirement, even if TOTP is not separately configured.
- The enforcement modal (shown on app load) should offer both options: set up TOTP or register a passkey.
- The check becomes: `enforce_totp && !totpEnabled && !hasPasskeys && !idpLogin`.

### 3.4 Auth Provider (IdP) Users

- Passkey registration and login are **not available** for IdP users.
- The passkey management UI is hidden for IdP users (same pattern as TOTP).
- The "Sign in with passkey" button on the login page is **always visible** (an IdP user who also had a local password could have passkeys; the server validates eligibility on the backend).

### 3.5 Password Reset / Invitation Flow

- After a password reset, existing passkeys remain registered. The user can still log in with passkeys.
- After accepting an invitation (setting initial password), the user has no passkeys. They can register passkeys from the preferences page.

---

## 4. Data Model

### 4.1 New Table: `passkeys`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `uuid` | PK, DEFAULT `uuid_generate_v4()` | Internal row ID. |
| `user_id` | `uuid` | NOT NULL, FK → `users.id` ON DELETE CASCADE | Owning user. |
| `credential_id` | `bytea` | NOT NULL, UNIQUE | WebAuthn credential ID (binary). |
| `public_key` | `bytea` | NOT NULL | CBOR-encoded public key. |
| `attestation_type` | `varchar` | NOT NULL | Attestation type (e.g., `"none"`). |
| `aaguid` | `bytea` | NOT NULL | Authenticator Attestation GUID. |
| `sign_count` | `bigint` | NOT NULL DEFAULT 0 | Signature counter for clone detection. |
| `name` | `varchar(255)` | NOT NULL | User-given display name (e.g., "MacBook Touch ID"). |
| `transports` | `varchar[]` | | Authenticator transports (e.g., `{"internal","hybrid"}`). |
| `created_at` | `timestamp` | NOT NULL DEFAULT NOW() | Registration timestamp. |
| `last_used_at` | `timestamp` | NULL | Last successful authentication timestamp. |

**Indexes:**
- `UNIQUE INDEX passkeys_credential_id ON passkeys(credential_id)`
- `INDEX passkeys_user_id ON passkeys(user_id)`

### 4.2 Database Migration

In `db-updates.go`, increment `targetVersion` to **37** and add the `passkeys` table creation:

```go
if curVersion < 37 {
    _, err := GetDatabase().DB().Exec(`
        CREATE TABLE IF NOT EXISTS passkeys (
            id uuid DEFAULT uuid_generate_v4(),
            user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            credential_id bytea NOT NULL,
            public_key bytea NOT NULL,
            attestation_type varchar NOT NULL DEFAULT 'none',
            aaguid bytea NOT NULL,
            sign_count bigint NOT NULL DEFAULT 0,
            name varchar(255) NOT NULL,
            transports varchar[] DEFAULT '{}',
            created_at timestamp NOT NULL DEFAULT NOW(),
            last_used_at timestamp NULL,
            PRIMARY KEY (id)
        )`)
    // + create indexes
}
```

### 4.3 New Repository: `passkey-repository.go`

Register `GetPasskeyRepository()` in the repositories list in `db-updates.go`.

**Entity:**

```go
type Passkey struct {
    ID              string
    UserID          string
    CredentialID    []byte
    PublicKey       []byte
    AttestationType string
    AAGUID          []byte
    SignCount       uint32
    Name            string
    Transports      []string
    CreatedAt       time.Time
    LastUsedAt      *time.Time
}
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Create(p *Passkey) error` | Insert a new passkey. |
| `GetByID(id string) (*Passkey, error)` | Get a passkey by internal ID. |
| `GetByCredentialID(credentialID []byte) (*Passkey, error)` | Look up credential by WebAuthn credential ID. |
| `GetAllByUserID(userID string) ([]*Passkey, error)` | List all passkeys for a user. |
| `GetCountByUserID(userID string) (int, error)` | Count passkeys for a user. |
| `UpdateSignCount(id string, signCount uint32) error` | Update the sign counter after authentication. |
| `UpdateLastUsedAt(id string, t time.Time) error` | Record last authentication time. |
| `UpdateName(id string, name string) error` | Rename a passkey. |
| `Delete(id string) error` | Remove a single passkey. |
| `DeleteAllByUserID(userID string) error` | Remove all passkeys for a user. |

### 4.4 New AuthState Types

Add two new `AuthStateType` constants for the WebAuthn challenge/response lifecycle:

| Constant | Value | Purpose | Payload | Expiry |
|----------|-------|---------|---------|--------|
| `AuthPasskeyRegistration` | 10 | Passkey registration ceremony | JSON: `webauthn.SessionData` | 5 min |
| `AuthPasskeyLogin` | 11 | Passkey authentication ceremony | JSON: `webauthn.SessionData` + login context | 5 min |

These follow the existing `AuthState` pattern used by TOTP setup and OAuth flows.

---

## 5. Backend API

### 5.1 New Go Dependency

Add the [`github.com/go-webauthn/webauthn`](https://github.com/go-webauthn/webauthn) library to `go.mod`. This is the standard Go library for WebAuthn server-side operations.

### 5.2 WebAuthn Configuration

Add to `config.go`:

| Env Variable | Default | Description |
|-------|---------|-------------|
| `WEBAUTHN_RP_DISPLAY_NAME` | `"Seatsurfing"` | Human-readable RP display name shown in passkey prompts. |
| `WEBAUTHN_RP_ID` | (derived from request `Host` header) | The RP ID for WebAuthn. Should be the domain (e.g., `seatsurfing.example.com`). If unset, derived automatically per-request. |
| `WEBAUTHN_RP_ORIGINS` | (derived from request) | Comma-separated list of allowed origins. If unset, derived automatically. |

The `webauthn.WebAuthn` instance is initialized at startup using these values. If per-organization RP IDs are needed (multi-tenant), the RP ID is derived from the organization's primary domain.

### 5.3 WebAuthn User Interface Implementation

The `go-webauthn` library requires a `webauthn.User` interface. Implement an adapter:

```go
type WebAuthnUser struct {
    user    *User
    passkeys []*Passkey
}

func (u *WebAuthnUser) WebAuthnID() []byte {
    return []byte(u.user.ID) // UUID as bytes
}

func (u *WebAuthnUser) WebAuthnName() string {
    return u.user.Email
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
    return u.user.GetDisplayName()
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
    // Convert []Passkey to []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnIcon() string {
    return ""
}
```

### 5.4 Passkey Management Endpoints (Authenticated)

These endpoints are added to the `UserRouter` (same router as TOTP endpoints). All require a valid JWT.

#### `GET /user/passkey/`

List all passkeys for the authenticated user.

**Response** `200 OK`:
```json
[
  {
    "id": "uuid",
    "name": "MacBook Touch ID",
    "createdAt": "2026-01-15T10:30:00Z",
    "lastUsedAt": "2026-02-18T14:22:00Z",
    "transports": ["internal"]
  }
]
```

**Notes:** Does not return `credentialId`, `publicKey`, or other sensitive fields.

---

#### `POST /user/passkey/registration/begin`

Start the WebAuthn registration ceremony.

**Preconditions:**
- User must have a password set (`HashedPassword != ""`).
- User must not be an IdP user (`AuthProviderID` is null/empty).

**Response** `200 OK`:
```json
{
  "stateId": "uuid",
  "options": { /* PublicKeyCredentialCreationOptions */ }
}
```

The `options` object contains the standard WebAuthn `PublicKeyCredentialCreationOptions`, including:
- `rp`: `{ id, name }`
- `user`: `{ id, name, displayName }`
- `challenge`: random challenge
- `pubKeyCredParams`: `[{type: "public-key", alg: -7}, {type: "public-key", alg: -257}]` (ES256> + RS256)
- `authenticatorSelection`: `{ residentKey: "required", userVerification: "required" }`
- `attestation`: `"none"` (no attestation needed for passkeys)
- `excludeCredentials`: list of the user's already-registered credential IDs (prevents re-registration).
- `timeout`: 300000 (5 minutes)

**Server-side:** Creates an `AuthState` (type `AuthPasskeyRegistration`) containing the `webauthn.SessionData` serialized as JSON. Expiry: 5 minutes.

**Error responses:**
- `403 Forbidden` — user is IdP user or has no password.

---

#### `POST /user/passkey/registration/finish`

Complete the WebAuthn registration ceremony.

**Request body:**
```json
{
  "stateId": "uuid",
  "name": "MacBook Touch ID",
  "credential": { /* AuthenticatorAttestationResponse */ }
}
```

**Validation:**
- `name` is required, max 255 characters, trimmed. Must not be empty after trimming.
- `name` must be unique among the user's existing passkeys.
- `stateId` must reference a valid, non-expired `AuthPasskeyRegistration` state owned by the user.
- Rate limit: max 5 completion attempts per `stateId` (same pattern as TOTP validation).

**Server-side:**
1. Load the `AuthState` and deserialize `webauthn.SessionData`.
2. Call `webauthn.FinishRegistration()` with the credential response.
3. On success: create a `Passkey` record in the database.
4. Delete the `AuthState`.
5. Return `201 Created`.

**Response** `201 Created`:
```json
{
  "id": "uuid",
  "name": "MacBook Touch ID",
  "createdAt": "2026-02-19T10:30:00Z"
}
```

**Error responses:**
- `400 Bad Request` — invalid credential, invalid name, or duplicate name.
- `404 Not Found` — state ID not found or expired.
- `429 Too Many Requests` — exceeded attempt limit.

---

#### `PUT /user/passkey/{id}`

Rename an existing passkey.

**Request body:**
```json
{
  "name": "Office YubiKey"
}
```

**Validation:**
- Passkey must belong to the authenticated user.
- `name` must be unique among the user's passkeys, max 255 characters.

**Response** `200 OK`.

**Error responses:**
- `400 Bad Request` — invalid or duplicate name.
- `404 Not Found` — passkey not found or not owned by user.

---

#### `DELETE /user/passkey/{id}`

Delete a passkey.

**Validation:**
- Passkey must belong to the authenticated user.
- If `enforce_totp` is enabled and this is the user's last passkey and TOTP is not configured → `403 Forbidden` (deleting would leave the user without a second factor, violating the enforcement policy).

**Response** `204 No Content`.

---

### 5.5 Passkey Authentication Endpoints (Unauthenticated)

These endpoints are added to the `AuthRouter` and whitelisted in the auth middleware (same as `/auth/login`).

#### `POST /auth/passkey/login/begin`

Start a passwordless passkey authentication ceremony.

**Request body:**
```json
{
  "organizationId": "uuid"
}
```

**Response** `200 OK`:
```json
{
  "stateId": "uuid",
  "options": { /* PublicKeyCredentialRequestOptions */ }
}
```

The `options` object contains:
- `challenge`: random challenge
- `rpId`: the RP ID
- `userVerification`: `"required"`
- `timeout`: 300000 (5 minutes)
- `allowCredentials`: **empty** (discoverable credential flow — the authenticator selects the credential)

**Server-side:** Creates an `AuthState` (type `AuthPasskeyLogin`) with the `webauthn.SessionData`. Expiry: 5 minutes.

**Notes:**
- `organizationId` is required to scope the RP ID correctly in multi-tenant deployments.
- No user identification is needed at this stage (discoverable credentials).

---

#### `POST /auth/passkey/login/finish`

Complete a passwordless passkey authentication ceremony.

**Request body:**
```json
{
  "stateId": "uuid",
  "credential": { /* AuthenticatorAssertionResponse */ }
}
```

**Server-side:**
1. Load the `AuthState` and deserialize `webauthn.SessionData`.
2. The credential response contains the `userHandle` (= `WebAuthnID()` = user UUID).
3. Look up the user by the `userHandle`.
4. Load the user's passkey credentials.
5. Call `webauthn.FinishLogin()` to verify the assertion.
6. **Validate preconditions:**
   - User is not disabled.
   - User is not banned (check `BanExpiry`).
   - User has a password set (not an IdP-only user).
   - User is not `PasswordPending`.
7. **Clone detection:** If the returned `signCount` is ≤ the stored `signCount` and the stored `signCount` is > 0, reject the login (credential may have been cloned). Log a warning.
8. Update `signCount` and `lastUsedAt` on the passkey record.
9. Record a successful login attempt via `AuthAttemptRepository.RecordLoginAttempt()`.
10. Update user `LastActivityAtUTC`.
11. Create session (same logic as password login: enforce `MAX_SESSIONS_PER_USER`).
12. Generate JWT access token + refresh token.
13. Delete the `AuthState`.
14. Return `200 OK` with `{ accessToken, refreshToken }`.

**Error responses:**
- `400 Bad Request` — invalid assertion.
- `401 Unauthorized` — user is disabled/banned.
- `404 Not Found` — state not found, expired, or user not found.

**Brute-force protection:**
- Failed passkey login attempts are recorded via `AuthAttemptRepository.RecordLoginAttempt(user, false)`.
- The existing brute-force rate-limiting/banning logic applies.

---

### 5.6 Passkey as Second Factor (Password Login Modification)

The existing `POST /auth/login` (`loginPassword`) handler is modified:

**Current behavior (unchanged for users without passkeys):**
1. Verify username + password.
2. If TOTP is configured and no `code` provided → respond `401`.
3. If TOTP `code` provided → verify it.

**New behavior (for users with passkeys):**

After successful password verification:

1. Check if user has registered passkeys (`GetPasskeyRepository().GetCountByUserID()`).
2. **If user has passkeys:**
   - If request contains a passkey assertion (`passkeyStateId` + `passkeyCredential` fields) → verify the passkey assertion (same logic as `POST /auth/passkey/login/finish` steps 5-8). If valid → login succeeds.
   - If request does **not** contain a passkey assertion:
     - Create a passkey login `AuthState` (`AuthPasskeyLogin`).
     - Return `HTTP 401` with a JSON body:
       ```json
       {
         "requirePasskey": true,
         "stateId": "uuid",
         "options": { /* PublicKeyCredentialRequestOptions */ },
         "allowTotpFallback": true
       }
       ```
     - The `options.allowCredentials` lists only this user's credential IDs (non-discoverable assertion, since we already know the user).
     - `allowTotpFallback` is `true` only if the user also has TOTP configured.
3. **If user does NOT have passkeys, but has TOTP:** existing behavior (return `401` for TOTP code).
4. **TOTP fallback:** If the request contains a `code` field (TOTP code) and the user has both passkeys and TOTP configured, TOTP verification is still accepted. This allows users to fall back to TOTP when their passkey device is unavailable.

**Modified request body** (`AuthPasswordRequest`):
```go
type AuthPasswordRequest struct {
    Email              string `json:"email"`
    Password           string `json:"password"`
    OrganizationID     string `json:"organizationId"`
    Code               string `json:"code"`               // existing: TOTP code
    PasskeyStateID     string `json:"passkeyStateId"`      // new: passkey state ID
    PasskeyCredential  any    `json:"passkeyCredential"`   // new: AuthenticatorAssertionResponse
}
```

**Decision logic after password verification:**

```
Has passkeys?
├── YES → Passkey assertion provided?
│   ├── YES → Verify passkey → Success / Fail
│   └── NO → TOTP code provided AND TOTP configured?
│       ├── YES → Verify TOTP → Success / Fail (fallback)
│       └── NO → Return 401 with passkey challenge + allowTotpFallback flag
└── NO → Has TOTP?
    ├── YES → TOTP code provided?
    │   ├── YES → Verify TOTP → Success / Fail
    │   └── NO → Return 401 (require TOTP)
    └── NO → Login succeeds (no 2FA configured)
```

### 5.7 User Info Response Modification

The `GetUserResponse` struct (returned by `GET /user/me` and user endpoints) gains a new field:

```go
type GetUserResponse struct {
    // ... existing fields ...
    TotpEnabled     bool   `json:"totpEnabled"`
    HasPasskeys     bool   `json:"hasPasskeys"`     // NEW
}
```

Set `HasPasskeys = GetPasskeyRepository().GetCountByUserID(user.ID) > 0`.

### 5.8 Route Summary

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET` | `/user/passkey/` | JWT | List user's passkeys |
| `POST` | `/user/passkey/registration/begin` | JWT | Start passkey registration |
| `POST` | `/user/passkey/registration/finish` | JWT | Complete passkey registration |
| `PUT` | `/user/passkey/{id}` | JWT | Rename a passkey |
| `DELETE` | `/user/passkey/{id}` | JWT | Delete a passkey |
| `POST` | `/auth/passkey/login/begin` | None | Start passwordless login |
| `POST` | `/auth/passkey/login/finish` | None | Complete passwordless login |

Unauthenticated routes (`/auth/passkey/*`) must be added to the whitelist in `unauthorized-routes.go`.

---

## 6. Frontend Changes

### 6.1 Login Page (`/login`)

#### Passwordless Passkey Login

Add a **"Sign in with passkey"** button on the login page. This button is:

- Displayed when the browser supports WebAuthn (`window.PublicKeyCredential` is available) **and** `PublicKeyCredential.isConditionalMediationAvailable()` or `isUserVerifyingPlatformAuthenticatorAvailable()` resolves to `true`.
- Displayed **alongside** the existing login form (both the Auth Provider buttons and the username/password form).
- Visually separated from other login methods (e.g., a divider with "or").
- **Not** gated by `disablePasswordLogin` — passkey login is independent (though passkeys can only be registered by password users).

**Flow:**
1. User clicks "Sign in with passkey".
2. Frontend calls `POST /auth/passkey/login/begin` with `{ organizationId }`.
3. Frontend calls `navigator.credentials.get()` with the received options.
4. User selects a passkey and authenticates via biometrics/PIN.
5. Frontend sends the assertion to `POST /auth/passkey/login/finish`.
6. On success: store tokens, load user info, redirect to `/search`.
7. On failure: show error message ("Passkey login failed. Please try again or use another login method.").

#### Passkey as Second Factor (After Password)

When the existing password login returns `401` with `{ requirePasskey: true }`:

1. Hide the password form.
2. Automatically trigger `navigator.credentials.get()` with the provided options.
3. On success: send the assertion back in a second `POST /auth/login` request (with the original `email`, `password`, `passkeyStateId`, and `passkeyCredential`).
4. On failure or user cancellation:
   - If `allowTotpFallback` is `true`: show the existing TOTP input form with a message ("Passkey verification failed. Enter your TOTP code instead.").
   - If `allowTotpFallback` is `false`: show an error ("Passkey verification required. Please try again.").

### 6.2 Preferences Page (`/preferences` → Security Tab)

#### Passkey Management Component (`PasskeySettings`)

Add a new `PasskeySettings` component in the Security tab, below the existing `TotpSettings`. Hidden when `RuntimeConfig.INFOS.idpLogin` is `true` (same logic as TOTP).

**Content:**

1. **Header:** "Passkeys" with a brief description ("Sign in securely without a password using biometrics or a security key.").
2. **Passkey list** (loaded via `GET /user/passkey/`):
   - Each entry shows: name, creation date, last used date (or "Never used").
   - Each entry has:
     - **Rename** action (inline edit or modal → `PUT /user/passkey/{id}`).
     - **Delete** action (confirmation dialog → `DELETE /user/passkey/{id}`).
3. **"Add passkey" button:**
   - Disabled if `window.PublicKeyCredential` is not available (with a tooltip explaining browser support is required).
   - On click:
     1. Prompt user for a name (modal with text input, e.g., "Name this passkey").
     2. Call `POST /user/passkey/registration/begin`.
     3. Call `navigator.credentials.create()` with the received options.
     4. On success: call `POST /user/passkey/registration/finish` with the credential and name.
     5. Show success message and refresh the passkey list.
     6. On failure: show error ("Could not register passkey. Please try again.").
4. **Empty state:** When no passkeys are registered, show a message encouraging the user to add one.

### 6.3 TOTP Enforcement Modal Update

The existing TOTP enforcement modal (shown in `_app.tsx` when `enforce_totp` is enabled and user has no 2FA) should be updated:

- **Condition change:** Show the enforcement modal when `enforceTOTP && !totpEnabled && !hasPasskeys && !idpLogin` (previously: `enforceTOTP && !totpEnabled && !idpLogin`).
- **Modal content:** Offer both options:
  - "Set up authenticator app (TOTP)" — existing TOTP setup flow.
  - "Register a passkey" — triggers passkey registration flow.
- Once either is completed, the enforcement is satisfied and the modal closes.

### 6.4 RuntimeConfig Update

Add to `RuntimeUserInfos`:

```typescript
interface RuntimeUserInfos {
  // ... existing fields ...
  totpEnabled: boolean;
  hasPasskeys: boolean;  // NEW
  enforceTOTP: boolean;
}
```

Set from the `GET /user/me` response: `RuntimeConfig.INFOS.hasPasskeys = user.hasPasskeys`.

### 6.5 User Type Update

Add to the `User` type (`ui/src/types/User.ts`):

```typescript
// Existing fields
totpEnabled: boolean;
hasPasskeys: boolean;  // NEW

// New static methods
static listPasskeys(): Promise<PasskeyInfo[]>;
static beginPasskeyRegistration(): Promise<{ stateId: string; options: PublicKeyCredentialCreationOptions }>;
static finishPasskeyRegistration(stateId: string, name: string, credential: Credential): Promise<PasskeyInfo>;
static renamePasskey(id: string, name: string): Promise<void>;
static deletePasskey(id: string): Promise<void>;
static beginPasskeyLogin(organizationId: string): Promise<{ stateId: string; options: PublicKeyCredentialRequestOptions }>;
static finishPasskeyLogin(stateId: string, credential: Credential): Promise<{ accessToken: string; refreshToken: string }>;
```

### 6.6 New Type: `PasskeyInfo`

```typescript
interface PasskeyInfo {
  id: string;
  name: string;
  createdAt: string;
  lastUsedAt: string | null;
  transports: string[];
}
```

### 6.7 Internationalization (i18n)

Add translation keys for all supported languages. English keys include:

| Key | Value |
|-----|-------|
| `passkey` | Passkey |
| `passkeys` | Passkeys |
| `passkeyDescription` | Sign in securely without a password using biometrics or a security key. |
| `addPasskey` | Add passkey |
| `namePasskey` | Name this passkey |
| `passkeyNamePlaceholder` | e.g., MacBook Touch ID |
| `renamePasskey` | Rename passkey |
| `deletePasskey` | Delete passkey |
| `deletePasskeyConfirm` | Are you sure you want to delete the passkey "{name}"? |
| `passkeyRegistered` | Passkey registered successfully. |
| `passkeyDeleted` | Passkey deleted. |
| `passkeyRenamed` | Passkey renamed. |
| `passkeyLoginFailed` | Passkey login failed. Please try again or use another login method. |
| `passkeyRequired` | Please verify your identity with a passkey. |
| `passkeyFallbackToTotp` | Use authenticator code instead |
| `signInWithPasskey` | Sign in with passkey |
| `noPasskeys` | No passkeys registered yet. |
| `lastUsed` | Last used |
| `neverUsed` | Never used |
| `passkeyNotSupported` | Your browser does not support passkeys. |
| `passkeyEnforcementMessage` | Your organization requires a second factor. Set up an authenticator app or register a passkey. |
| `passkeyCannotDeleteLast` | Cannot delete the last passkey while second-factor enforcement is active. Set up an authenticator app first or add another passkey. |

---

## 7. Security Considerations

### 7.1 User Verification

All WebAuthn ceremonies **require** user verification (`userVerification: "required"`). This ensures the authenticator confirms the user's identity via biometrics or PIN, providing multi-factor assurance (possession + inherence/knowledge) in a single step.

### 7.2 Discoverable Credentials

Passkey registration requests `residentKey: "required"` to ensure credentials are discoverable. This enables the passwordless login flow where the user does not need to provide a username.

### 7.3 Attestation

Attestation is set to `"none"`. Seatsurfing does not need to verify the specific make/model of the authenticator. This maximizes compatibility and avoids privacy concerns.

### 7.4 Challenge Expiry

WebAuthn challenges (stored as `AuthState`) expire after **5 minutes**, consistent with existing TOTP setup and OAuth state expiry.

### 7.5 Clone Detection

The `signCount` is checked during authentication. If the returned sign count is not greater than the stored value (and the stored value is > 0), the login is rejected and a warning is logged. This detects cloned authenticators. (Note: Many modern platform authenticators always return `signCount = 0`, in which case clone detection is not applicable and should be skipped.)

### 7.6 Brute-Force Protection

Failed passkey authentication attempts are recorded through the existing `AuthAttemptRepository`, applying the same sliding-window ban logic as password logins.

### 7.7 Origin Validation

The `go-webauthn` library validates the origin in the client data against the configured RP origins. In multi-tenant deployments, the RP origin must match the organization's domain.

### 7.8 Credential Storage

Credential public keys and IDs are stored as raw bytes (`bytea`) in PostgreSQL. No encryption is needed — public keys are not secrets. The security of the passkey system relies on the private key never leaving the authenticator.

### 7.9 Account Deletion

When a user is deleted, all associated passkeys are automatically removed via the `ON DELETE CASCADE` foreign key constraint.

---

## 8. Admin Considerations

### 8.1 Admin User Management

When an admin views a user's details (admin panel → Users → user detail):

- Show a read-only count of registered passkeys (e.g., "Passkeys: 3").
- Provide an admin action to **remove all passkeys** for a user (useful for account recovery scenarios).
- This calls a new admin endpoint: `DELETE /user/{userId}/passkey/` (requires OrgAdmin role).

### 8.2 Organization Settings

No new organization settings are required. The existing `enforce_totp` setting implicitly covers passkeys (see section 3.3).

---

## 9. Testing Strategy

### 9.1 Backend Unit Tests

Add tests in `server/repository/test/` for the passkey repository (CRUD operations).

Add tests in the router test suite for:

| Test Case | Description |
|-----------|-------------|
| Registration begin – valid | Authenticated user gets valid creation options. |
| Registration begin – IdP user | Returns 403. |
| Registration finish – valid | Credential is stored, state is deleted. |
| Registration finish – expired state | Returns 404. |
| Registration finish – duplicate name | Returns 400. |
| Registration finish – rate limit | Returns 429 after 5 attempts. |
| Passwordless login – valid | Full ceremony succeeds, tokens returned. |
| Passwordless login – disabled user | Returns 401. |
| Passwordless login – banned user | Returns 401. |
| Passwordless login – clone detection | Returns error when sign count is suspicious. |
| Password + passkey 2FA – valid | Password login with passkey assertion succeeds. |
| Password + passkey 2FA – challenge returned | Password login returns 401 with passkey options. |
| Password + TOTP fallback | User with both passkeys and TOTP can fall back to TOTP code. |
| List passkeys | Returns passkey metadata (no secrets). |
| Delete passkey | Removes passkey. |
| Delete last passkey + enforce_totp | Returns 403. |
| Rename passkey | Updates name. |
| Rename passkey – duplicate name | Returns 400. |

### 9.2 Frontend Tests

Add Vitest unit tests for:

- `PasskeySettings` component rendering and interactions.
- Login page passkey button visibility (based on WebAuthn availability).
- TOTP enforcement modal condition update (show when no TOTP and no passkeys).

### 9.3 E2E Tests

Add Playwright tests in `e2e/tests/` covering:

- Passkey registration flow (requires WebAuthn virtual authenticator — Playwright supports this via `cdpSession`).
- Passwordless passkey login.
- Password + passkey second factor login.
- Passkey deletion and enforcement interaction.

---

## 10. Rollout Considerations

### 10.1 Feature Flag

No feature flag is required. Passkeys are opt-in per user (users must explicitly register a passkey). The "Sign in with passkey" button appears only when the browser supports WebAuthn.

### 10.2 Migration

The database migration (schema version 37) adds the `passkeys` table. No data migration is needed — all users start with zero passkeys.

### 10.3 Backward Compatibility

- Users without passkeys experience no change to their login flow.
- The `POST /auth/login` response is extended but remains backward-compatible (new fields are additive).
- The `GET /user/me` response gains `hasPasskeys: false` by default — existing clients ignore unknown fields.

---

## 11. Summary of Changes by File

### New Files

| File | Description |
|------|-------------|
| `server/repository/passkey-repository.go` | Passkey data access layer. |
| `server/router/passkey-router.go` | Authenticated passkey management endpoints. |
| `ui/src/components/PasskeySettings.tsx` | Passkey management UI component. |
| `ui/src/types/Passkey.ts` | Passkey TypeScript type and API methods. |

### Modified Files

| File | Change |
|------|--------|
| `server/go.mod` | Add `github.com/go-webauthn/webauthn` dependency. |
| `server/config/config.go` | Add `WEBAUTHN_RP_*` configuration fields. |
| `server/repository/db-updates.go` | Increment `targetVersion` to 37; register `GetPasskeyRepository()`. |
| `server/repository/auth-state-repository.go` | Add `AuthPasskeyRegistration` (10) and `AuthPasskeyLogin` (11) state types. |
| `server/router/auth-router.go` | Add `/auth/passkey/login/begin` and `/auth/passkey/login/finish` routes; modify `loginPassword` to support passkey as second factor. |
| `server/router/user-router.go` | Add `HasPasskeys` to `GetUserResponse`; register passkey management routes. |
| `server/router/unauthorized-routes.go` | Whitelist `/auth/passkey/` routes. |
| `ui/src/components/RuntimeConfig.ts` | Add `hasPasskeys` to `RuntimeUserInfos`. |
| `ui/src/types/User.ts` | Add `hasPasskeys` field; add passkey API methods. |
| `ui/src/pages/login/index.tsx` | Add "Sign in with passkey" button; handle passkey 2FA challenge. |
| `ui/src/pages/preferences.tsx` | Add `PasskeySettings` component to Security tab. |
| `ui/src/pages/_app.tsx` | Update enforcement check to include `hasPasskeys`. |
| `ui/i18n/*.json` | Add passkey-related translation strings. |
