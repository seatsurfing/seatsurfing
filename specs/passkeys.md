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

| Term                        | Definition                                                                                                         |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| **Passkey**                 | A FIDO2/WebAuthn discoverable credential stored on the user's device or in a cloud-synced keychain.                |
| **Relying Party (RP)**      | The Seatsurfing server, identified by its RP ID (domain) and RP name.                                              |
| **Discoverable credential** | A WebAuthn credential that can initiate authentication without the user first providing a username (resident key). |
| **User verification**       | Biometric/PIN check performed locally by the authenticator during WebAuthn ceremonies.                             |
| **Ceremony**                | A WebAuthn registration or authentication handshake between browser and server.                                    |

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

| Column               | Type           | Constraints                                 | Description                                                                                                                  |
| -------------------- | -------------- | ------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `id`                 | `uuid`         | PK, DEFAULT `uuid_generate_v4()`            | Internal row ID.                                                                                                             |
| `user_id`            | `uuid`         | NOT NULL, FK → `users.id` ON DELETE CASCADE | Owning user.                                                                                                                 |
| `credential_id`      | `varchar`      | NOT NULL                                    | **Encrypted** (AES-GCM via `EncryptString`). Base64-encoded binary credential ID, then encrypted.                            |
| `credential_id_hash` | `varchar(64)`  | NOT NULL, UNIQUE                            | SHA-256 hex hash of the raw credential ID (pre-encryption). Used for lookups since the encrypted value is non-deterministic. |
| `public_key`         | `varchar`      | NOT NULL                                    | **Encrypted** (AES-GCM via `EncryptString`). Base64-encoded CBOR public key, then encrypted.                                 |
| `attestation_type`   | `varchar`      | NOT NULL                                    | Attestation type (e.g., `"none"`). Not encrypted (non-sensitive metadata).                                                   |
| `aaguid`             | `varchar`      | NOT NULL                                    | **Encrypted** (AES-GCM via `EncryptString`). Base64-encoded AAGUID, then encrypted.                                          |
| `sign_count`         | `bigint`       | NOT NULL DEFAULT 0                          | Signature counter for clone detection. Not encrypted (non-sensitive integer).                                                |
| `name`               | `varchar(255)` | NOT NULL                                    | User-given display name (e.g., "MacBook Touch ID"). Not encrypted.                                                           |
| `transports`         | `varchar[]`    |                                             | Authenticator transports (e.g., `{"internal","hybrid"}`). Not encrypted (non-sensitive metadata).                            |
| `created_at`         | `timestamp`    | NOT NULL DEFAULT NOW()                      | Registration timestamp.                                                                                                      |
| `last_used_at`       | `timestamp`    | NULL                                        | Last successful authentication timestamp.                                                                                    |

**Encryption scheme:** Binary fields (`credential_id`, `public_key`, `aaguid`) are first base64-encoded to produce a string, then encrypted via `EncryptString()` (AES-256-GCM using the `CRYPT_KEY` environment variable) before being stored. On read, values are decrypted via `DecryptString()` and base64-decoded back to binary. This matches the pattern used for TOTP secrets in `user-repository.go`.

**Lookup strategy:** Because `EncryptString()` uses a random nonce (non-deterministic), the same plaintext produces different ciphertexts. The `credential_id_hash` column stores a SHA-256 hex digest of the **raw** (unencrypted) credential ID bytes, enabling `WHERE credential_id_hash = $1` lookups without decryption.

**Indexes:**

- `UNIQUE INDEX passkeys_credential_id_hash ON passkeys(credential_id_hash)`
- `INDEX passkeys_user_id ON passkeys(user_id)`

### 4.2 Database Migration

In `db-updates.go`, increment `targetVersion` to **37**. Register `GetPasskeyRepository()` in the repositories list so its `RunSchemaUpgrade()` is called.

Following the existing pattern (see `session-repository.go` for reference), the **table and index creation** happens inside the `GetPasskeyRepository()` singleton initializer, not in `RunSchemaUpgrade()`. The `CREATE TABLE IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS` statements are idempotent and run on every startup:

```go
func GetPasskeyRepository() *PasskeyRepository {
    passkeyRepositoryOnce.Do(func() {
        passkeyRepository = &PasskeyRepository{}
        _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS passkeys (" +
            "id uuid DEFAULT uuid_generate_v4(), " +
            "user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, " +
            "credential_id varchar NOT NULL, " +
            "credential_id_hash varchar(64) NOT NULL, " +
            "public_key varchar NOT NULL, " +
            "attestation_type varchar NOT NULL DEFAULT 'none', " +
            "aaguid varchar NOT NULL, " +
            "sign_count bigint NOT NULL DEFAULT 0, " +
            "name varchar(255) NOT NULL, " +
            "transports varchar[] DEFAULT '{}', " +
            "created_at timestamp NOT NULL DEFAULT NOW(), " +
            "last_used_at timestamp NULL, " +
            "PRIMARY KEY (id))")
        if err != nil {
            panic(err)
        }
        if _, err = GetDatabase().DB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS " +
            "passkeys_credential_id_hash ON passkeys(credential_id_hash)"); err != nil {
            panic(err)
        }
        if _, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS " +
            "idx_passkeys_user_id ON passkeys(user_id)"); err != nil {
            panic(err)
        }
    })
    return passkeyRepository
}
```

The `RunSchemaUpgrade()` method starts empty (no schema changes yet) and will be used for future column additions or alterations:

```go
func (r *PasskeyRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
    // no schema changes yet
}
```

### 4.3 New Repository: `passkey-repository.go`

**Entity (in-memory, decrypted):**

```go
type Passkey struct {
    ID              string
    UserID          string
    CredentialID    []byte     // raw binary; encrypted at rest
    PublicKey       []byte     // raw binary; encrypted at rest
    AttestationType string
    AAGUID          []byte     // raw binary; encrypted at rest
    SignCount       uint32
    Name            string
    Transports      []string
    CreatedAt       time.Time
    LastUsedAt      *time.Time
}
```

The `Passkey` struct always holds **decrypted, decoded** binary values in memory. Encryption/decryption is handled transparently inside the repository methods, never leaking to callers.

**Encryption helpers** (internal to the repository):

```go
// encryptBytes: base64-encode raw bytes, then EncryptString()
func encryptBytes(data []byte) (string, error) {
    b64 := base64.StdEncoding.EncodeToString(data)
    return EncryptString(b64)
}

// decryptBytes: DecryptString(), then base64-decode to raw bytes
func decryptBytes(ciphertext string) ([]byte, error) {
    b64, err := DecryptString(ciphertext)
    if err != nil {
        return nil, err
    }
    return base64.StdEncoding.DecodeString(b64)
}

// hashCredentialID: SHA-256 hex digest for deterministic lookup
func hashCredentialID(credentialID []byte) string {
    h := sha256.Sum256(credentialID)
    return hex.EncodeToString(h[:])
}
```

**Methods:**

| Method                                                                                              | Description                                                                                                                                                                                       |
| --------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Create(p *Passkey) error`                                                                          | Encrypts `CredentialID`, `PublicKey`, `AAGUID` via `encryptBytes()`; computes `credential_id_hash` via `hashCredentialID()`; inserts the row.                                                     |
| `GetOne(id string) (*Passkey, error)`                                                               | Reads row by primary key, decrypts encrypted fields, returns populated struct.                                                                                                                    |
| `GetByCredentialIDRaw(credentialID []byte) (*Passkey, error)`                                       | Computes `hashCredentialID(credentialID)`, queries `WHERE credential_id_hash = $1`, decrypts fields. Used by the discoverable login handler where only the raw credential ID bytes are available. |
| `GetByCredentialIDHash(hash string) (*Passkey, error)`                                              | Queries `WHERE credential_id_hash = $1` directly when the caller already has the hex hash.                                                                                                        |
| `GetAllByUserID(userID string) ([]*Passkey, error)`                                                 | Lists all rows for user, decrypts each.                                                                                                                                                           |
| `GetCountByUserID(userID string) int`                                                               | `SELECT COUNT(*)` — no decryption needed. Returns `int` (0 on error).                                                                                                                             |
| `UpdateSignCount(p *Passkey) error`                                                                 | Updates the unencrypted `sign_count` column using `p.ID` and `p.SignCount`.                                                                                                                       |
| `UpdateLastUsedAt(p *Passkey) error`                                                                | Sets `last_used_at = NOW()` for the given passkey.                                                                                                                                                |
| `UpdateName(p *Passkey) error`                                                                      | Updates the unencrypted `name` column using `p.ID` and `p.Name`.                                                                                                                                  |
| `Delete(p *Passkey) error`                                                                          | Removes a single passkey row.                                                                                                                                                                     |
| `DeleteAllByUserID(userID string) error`                                                            | Removes all passkey rows for a user.                                                                                                                                                              |
| `ToWebAuthnCredential() (*webauthn.Credential, error)`                                              | Helper method on `*Passkey` — decrypts fields and reconstructs a `webauthn.Credential` for use in ceremonies.                                                                                     |
| `NewPasskeyFromCredential(userID string, cred *webauthn.Credential, name string) (*Passkey, error)` | Package-level constructor — wraps a freshly minted `webauthn.Credential` in a `Passkey` ready for `Create()`.                                                                                     |

### 4.4 New AuthState Types

Add three new `AuthStateType` constants for the WebAuthn challenge/response lifecycle:

| Constant                  | Value | Purpose                                        | Payload                                                                                                  | Expiry |
| ------------------------- | ----- | ---------------------------------------------- | -------------------------------------------------------------------------------------------------------- | ------ |
| `AuthPasskeyRegistration` | 10    | Passkey registration ceremony                  | JSON: `webauthn.SessionData`                                                                             | 5 min  |
| `AuthPasskeyLogin`        | 11    | Passkey authentication ceremony (passwordless) | JSON: `{ orgId: string, sessionData: webauthn.SessionData }` — encrypted with AES-256-GCM before storage | 5 min  |
| `AuthPasskey2FA`          | 12    | Passkey 2FA challenge after password login     | JSON: `{ orgId: string, sessionData: webauthn.SessionData }` — encrypted with AES-256-GCM before storage | 5 min  |

Using **distinct** state types for the passwordless and 2FA flows prevents cross-flow state reuse: a state created by `beginPasskeyLogin` (type 11) cannot be consumed by the 2FA path (which requires type 12), and vice versa.

These follow the existing `AuthState` pattern used by TOTP setup and OAuth flows.

---

## 5. Backend API

### 5.1 New Go Dependency

Add the [`github.com/go-webauthn/webauthn`](https://github.com/go-webauthn/webauthn) library to `go.mod`. This is the standard Go library for WebAuthn server-side operations.

### 5.2 WebAuthn Configuration

Add to `config.go`:

| Env Variable               | Default         | Description                                                                                                  |
| -------------------------- | --------------- | ------------------------------------------------------------------------------------------------------------ |
| `WEBAUTHN_RP_DISPLAY_NAME` | `"Seatsurfing"` | Human-readable RP display name shown in passkey prompts.                                                     |
| `MAX_PASSKEYS_PER_USER`    | `10`            | Maximum number of passkeys a single user may register. Must be ≥ 1; values below 1 are reset to the default. |

The RP ID and allowed origin are **always derived from the organisation's primary domain** (looked up via `GetOrganizationRepository().GetPrimaryDomain(org)`) and are **not configurable via environment variables**. This is the correct approach for a multi-tenant service: each organisation's passkeys are bound to its own domain, preventing credential reuse across organisations.

`getWebAuthnInstance(org *Organization)` is the internal factory that builds a `webauthn.WebAuthn` instance for a specific organisation. It returns an error if the organisation has no primary domain configured.

```go
func getWebAuthnInstance(org *Organization) (*webauthn.WebAuthn, error) {
    domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
    // ... strip port from domain for bare RPID ...
    return webauthn.New(&webauthn.Config{
        RPID:          rpID,               // bare domain, port stripped
        RPDisplayName: rpDisplayName,
        RPOrigins:     []string{FormatURL(domain.DomainName)}, // scheme://domain[:port]
    })
}
```

### 5.3 WebAuthn User Interface Implementation

The `go-webauthn` library requires a `webauthn.User` interface. Implement an adapter:

```go
type WebAuthnUser struct {
    user        *User
    credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte                         { return []byte(u.user.ID) }
func (u *WebAuthnUser) WebAuthnName() string                       { return u.user.Email }
func (u *WebAuthnUser) WebAuthnDisplayName() string                { return u.user.Firstname + " " + u.user.Lastname }
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }
```

Note: `WebAuthnIcon()` was removed in `go-webauthn` v0.11 and is not required.

A helper `loadWebAuthnUser(user *User) (*WebAuthnUser, error)` fetches all passkeys for the user, calls `pk.ToWebAuthnCredential()` for each, and returns the adapter. Passkeys whose decryption fails are skipped with a warning (so a corrupt credential does not lock the user out).

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
  "challenge": {
    /* PublicKeyCredentialCreationOptions */
  }
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

**Server-side:** Creates an `AuthState` (type `AuthPasskeyRegistration`) containing the `webauthn.SessionData` serialized as JSON. The payload is encrypted with AES-256-GCM (`EncryptString()`) before being stored. Expiry: 5 minutes.

**Error responses:**

- `403 Forbidden` — user is IdP user, has no password, or has already reached the `MAX_PASSKEYS_PER_USER` limit.

---

#### `POST /user/passkey/registration/finish`

Complete the WebAuthn registration ceremony.

**Request body:**

```json
{
  "stateId": "uuid",
  "name": "MacBook Touch ID",
  "credential": {
    /* AuthenticatorAttestationResponse */
  }
}
```

**Validation:**

- `name` is required, max 255 characters, and must not contain the characters `<`, `>`, `&`, `"`, `'`, or null bytes (returns `400 Bad Request`).
- `stateId` must reference a valid, non-expired `AuthPasskeyRegistration` state owned by the user.

**Server-side:**

1. Load the `AuthState` and deserialize `webauthn.SessionData`.
2. Call `webauthn.FinishRegistration()` with the credential response.
3. On success: create a `Passkey` record in the database.
4. Delete the `AuthState`.
5. Return `200 OK` with the new passkey record as JSON.

**Response** `200 OK`:

```json
{
  "id": "uuid",
  "name": "MacBook Touch ID",
  "createdAt": "2026-02-19T10:30:00Z"
}
```

**Error responses:**

- `400 Bad Request` — invalid credential or invalid name.
- `404 Not Found` — state ID not found or expired.

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
- `name` is required, max 255 characters, and must not contain `<`, `>`, `&`, `"`, `'`, or null bytes.

**Response** `204 No Content` (via `SendUpdated()`).

**Error responses:**

- `400 Bad Request` — missing or invalid name.
- `403 Forbidden` — passkey belongs to a different user.
- `404 Not Found` — passkey not found.

---

#### `DELETE /user/passkey/{id}`

Delete a passkey.

**Validation:**

- Passkey must belong to the authenticated user.
- If `enforce_totp` is enabled and this is the user's last passkey and TOTP is not configured → `403 Forbidden` (deleting would leave the user without a second factor, violating the enforcement policy).

**Response** `204 No Content` (via `SendUpdated()`).

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

`organizationId` is **required**. It is used to look up the organisation's primary domain, which defines the WebAuthn Relying Party ID and allowed origin for the ceremony.

**Response** `200 OK`:

```json
{
  "stateId": "uuid",
  "challenge": {
    /* PublicKeyCredentialRequestOptions */
  }
}
```

The `options` object contains:

- `challenge`: random challenge
- `rpId`: the RP ID
- `userVerification`: `"required"`
- `timeout`: 300000 (5 minutes)
- `allowCredentials`: **empty** (discoverable credential flow — the authenticator selects the credential)

**Server-side:** Creates an `AuthState` (type `AuthPasskeyLogin`). The payload JSON is `{ orgId: string, sessionData: webauthn.SessionData }` — the organisation ID is persisted so `finishPasskeyLogin` can reconstruct the same `webauthn.WebAuthn` instance (bound to the same primary domain) when verifying the assertion. The payload is encrypted with AES-256-GCM (`EncryptString()`) before being stored. Expiry: 5 minutes.

No user identification is needed at this stage (discoverable credentials).

**Prerequisites and protection:**

- `CanCrypt()` must return `true` (i.e., `CRYPT_KEY` must be configured). If not, the endpoint returns `500 Internal Server Error` immediately.
- The endpoint is **rate-limited** to **10 requests per minute per client IP** (using `ulule/limiter` with an in-memory store). Clients that exceed the limit receive `429 Too Many Requests`. The client IP is extracted with `X-Forwarded-For` support for deployments behind a reverse proxy.

---

#### `POST /auth/passkey/login/finish`

Complete a passwordless passkey authentication ceremony.

**Request body:**

```json
{
  "stateId": "uuid",
  "credential": {
    /* AuthenticatorAssertionResponse */
  }
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
- `429 Too Many Requests` — rate limit exceeded on `beginPasskeyLogin` (the ceremony cannot begin).

**Timing protection:**
To prevent a timing side-channel that could confirm whether a credential exists in the database, `finishPasskeyLogin` enforces a **minimum response time of 100 ms** via a deferred timer. This equalises the response time for the not-found path (single DB query) and the found path (multiple DB queries).

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
         "passkeyChallenge": {
           /* PublicKeyCredentialRequestOptions */
         },
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

> **State type note:** The passkey challenge `AuthState` created here uses type `AuthPasskey2FA` (not `AuthPasskeyLogin`). The 2FA finishing path validates incoming state IDs against `AuthPasskey2FA`, preventing a state generated by the passwordless flow from being consumed here.

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

| Method   | Path                                | Auth | Purpose                       |
| -------- | ----------------------------------- | ---- | ----------------------------- |
| `GET`    | `/user/passkey/`                    | JWT  | List user's passkeys          |
| `POST`   | `/user/passkey/registration/begin`  | JWT  | Start passkey registration    |
| `POST`   | `/user/passkey/registration/finish` | JWT  | Complete passkey registration |
| `PUT`    | `/user/passkey/{id}`                | JWT  | Rename a passkey              |
| `DELETE` | `/user/passkey/{id}`                | JWT  | Delete a passkey              |
| `POST`   | `/auth/passkey/login/begin`         | None | Start passwordless login      |
| `POST`   | `/auth/passkey/login/finish`        | None | Complete passwordless login   |

Unauthenticated routes (`/auth/passkey/*`) must be added to the whitelist in `unauthorized-routes.go`.

---

## 6. Frontend Changes

### 6.1 Login Page (`/login`)

#### Passwordless Passkey Login

Add a **"Sign in with passkey"** button on the login page. This button is:

- Displayed when the browser supports WebAuthn (`window.PublicKeyCredential` is available) as a fast synchronous pre-check, refined by an asynchronous call to `PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable()` on component mount. Both the login page and passkey settings component store the result in component state and use it to gate all passkey UI (`Passkey.isPlatformAuthAvailable()`).
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
  hasPasskeys: boolean; // NEW
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

| Key                         | Value                                                                                                                               |
| --------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `passkey`                   | Passkey                                                                                                                             |
| `passkeys`                  | Passkeys                                                                                                                            |
| `passkeyDescription`        | Sign in securely without a password using biometrics or a security key.                                                             |
| `addPasskey`                | Add passkey                                                                                                                         |
| `namePasskey`               | Name this passkey                                                                                                                   |
| `passkeyNamePlaceholder`    | e.g., MacBook Touch ID                                                                                                              |
| `renamePasskey`             | Rename passkey                                                                                                                      |
| `deletePasskey`             | Delete passkey                                                                                                                      |
| `deletePasskeyConfirm`      | Are you sure you want to delete the passkey "{name}"?                                                                               |
| `passkeyRegistered`         | Passkey registered successfully.                                                                                                    |
| `passkeyDeleted`            | Passkey deleted.                                                                                                                    |
| `passkeyRenamed`            | Passkey renamed.                                                                                                                    |
| `passkeyLoginFailed`        | Passkey login failed. Please try again or use another login method.                                                                 |
| `passkeyRequired`           | Please verify your identity with a passkey.                                                                                         |
| `passkeyFallbackToTotp`     | Use authenticator code instead                                                                                                      |
| `signInWithPasskey`         | Sign in with passkey                                                                                                                |
| `noPasskeys`                | No passkeys registered yet.                                                                                                         |
| `lastUsed`                  | Last used                                                                                                                           |
| `neverUsed`                 | Never used                                                                                                                          |
| `passkeyNotSupported`       | Your browser does not support passkeys.                                                                                             |
| `passkeyEnforcementMessage` | Your organization requires a second factor. Set up an authenticator app or register a passkey.                                      |
| `passkeyCannotDeleteLast`   | Cannot delete the last passkey while second-factor enforcement is active. Set up an authenticator app first or add another passkey. |

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

The `signCount` is checked during authentication in **both** the passwordless path (`finishPasskeyLogin`) and the 2FA path (`handlePasskey2FA`). If the returned sign count is not greater than the stored value (and the stored value is > 0), the login is rejected and a warning is logged. This detects cloned authenticators. (Note: Many modern platform authenticators always return `signCount = 0`, in which case clone detection is not applicable and should be skipped.)

### 7.6 Brute-Force Protection

Failed passkey authentication attempts are recorded through the existing `AuthAttemptRepository`, applying the same sliding-window ban logic as password logins.

### 7.7 Origin Validation

The `go-webauthn` library validates the origin in the client data against the configured RP origins. In multi-tenant deployments, the RP origin must match the organization's domain.

### 7.8 Credential Storage & Encryption at Rest

All credential material (`credential_id`, `public_key`, `aaguid`) is **encrypted at rest** using the existing `EncryptString()`/`DecryptString()` functions from `util/encryption.go`. These use AES-256-GCM with the `CRYPT_KEY` environment variable (the same key used for TOTP secret encryption).

**Encryption flow (write):** raw binary → base64-encode → `EncryptString()` → stored as `varchar`.
**Decryption flow (read):** stored `varchar` → `DecryptString()` → base64-decode → raw binary.

Because `EncryptString()` uses a random nonce (non-deterministic), a separate `credential_id_hash` column stores the SHA-256 hex digest of the raw credential ID for indexed lookups. This hash is not reversible and does not leak the credential ID (SHA-256 is preimage-resistant; credential IDs are high-entropy random values).

While public keys are not secrets in the cryptographic sense (the private key never leaves the authenticator), encrypting them at rest provides defense-in-depth: a database breach does not reveal any credential material that could be used for targeted phishing, credential correlation across services, or authenticator fingerprinting via AAGUIDs.

**WebAuthn challenge payloads** (`AuthState.Payload` for types `AuthPasskeyRegistration`, `AuthPasskeyLogin`, and `AuthPasskey2FA`) are **also encrypted** with AES-256-GCM before being written to the `auth_states` table. Decryption includes a plaintext fallback for backward compatibility with any unencrypted states written before this hardening. This protects the WebAuthn session data (challenge, allowed credentials) from direct database reads.

**Prerequisite:** `CRYPT_KEY` must be configured (32-byte key). The `CanCrypt()` function is checked at startup; if it returns `false`, passkey login (`beginPasskeyLogin`) returns `500 Internal Server Error`. This is the same prerequisite as TOTP.

### 7.9 Account Deletion

When a user is deleted, all associated passkeys are automatically removed via the `ON DELETE CASCADE` foreign key constraint.

### 7.10 Rate Limiting on Unauthenticated Endpoints

The `POST /auth/passkey/login/begin` endpoint is unauthenticated and could otherwise be used to trigger unlimited database writes (one `AuthState` per request). To prevent this:

- A **per-IP rate limiter** caps requests at **10 per minute**. The client IP is resolved using `X-Forwarded-For` headers (first trusted hop), falling back to `RemoteAddr` for direct connections.
- Requests exceeding the limit receive `429 Too Many Requests`.
- The rate limiter uses an **in-memory store** (no Redis dependency), suitable for single-instance deployments. Horizontal scaling would require a shared store.

### 7.11 Input Validation on Passkey Names

Passkey names are user-visible metadata displayed in the management UI. To prevent stored content that could be interpreted as markup in non-React contexts, names are validated on both `POST /user/passkey/registration/finish` and `PUT /user/passkey/{id}` to reject strings containing `<`, `>`, `&`, `"`, `'`, or null bytes. Invalid names return `400 Bad Request`.

### 7.12 Per-User Credential Limit

To prevent a compromised session from exhausting database storage or degrading credential-lookup performance, the number of passkeys a user may register is capped by `MAX_PASSKEYS_PER_USER` (default: 20, configurable via environment variable). `POST /user/passkey/registration/begin` returns `403 Forbidden` when the current count meets or exceeds the limit.

---

## 8. Multi-Domain Considerations

Seatsurfing supports multiple organizations, each accessible under one or more custom domains, with one domain flagged as the "primary domain". The WebAuthn Relying Party ID (RPID) is always derived from the organization's primary domain (see `getWebAuthnInstance` in section 5.2). When a user accesses the application from a non-primary domain, the browser rejects the WebAuthn ceremony because the RPID does not match the current origin. The user sees an error: _"The relying party ID is not a registrable domain suffix of, nor equal to the current domain."_

This section evaluates options for handling this mismatch and documents the chosen approach.

### 8.1 Options Considered

#### Option A: Domain-Scoped Passkeys

Change `getWebAuthnInstance` to use the current request's `Host` header as the RPID instead of the primary domain. Store the RPID alongside each passkey credential in the database.

**Pros:**

- Passkey registration works on every domain.
- Conceptually simple: each passkey is tied to the domain where it was created.
- No external infrastructure needed.

**Cons:**

- A passkey registered on domain A is **not usable** on domain B — users switching domains must register separate passkeys or fall back to TOTP.
- Potentially confusing if users don't understand why their passkey "disappeared" on another domain.
- More complex login logic (must match RPID per credential).

#### Option B: WebAuthn Related Origin Requests (`.well-known/webauthn`)

Keep the primary domain as RPID, but serve a `/.well-known/webauthn` JSON resource on the primary domain that lists all of the organization's non-primary domains as related origins. This is the [WebAuthn Level 3 Related Origin Requests](https://w3c.github.io/webauthn/#sctn-related-origins) mechanism.

**Pros:**

- Single RPID — passkeys work seamlessly across all org domains.
- Standards-based; the ideal long-term solution.

**Cons:**

- **Limited browser support** — Chrome 128+ and Edge support it; Safari and Firefox do not yet (as of early 2026).
- **Infrastructure requirement**: the `.well-known/webauthn` file must be served from the primary domain's origin, meaning the primary domain must be reachable even when the user is on a secondary domain.
- Partial solution until all major browsers adopt it.

#### Option C: Disable Passkey Registration on Non-Primary Domains

Detect whether the user is accessing the application from the primary domain. If not, hide or disable the "Add passkey" button and show an informational message directing the user to the primary domain.

**Pros:**

- Simplest change — minimal code, no schema changes, no new infrastructure.
- Eliminates the confusing browser error entirely.
- Clear, actionable user guidance.

**Cons:**

- Doesn't solve the underlying limitation — passkeys are still primary-domain-only.
- Users on non-primary domains cannot register passkeys (they must visit the primary domain).

#### Option D: Disable Registration on Non-Primary Domains + Domain-Scoped Login

A hybrid of Options A and C: passkey _registration_ remains restricted to the primary domain, but passkey _login_ uses the current request domain as origin, enabling passkeys to be used (not created) on non-primary domains — but only when the RPID (primary domain) is a registrable domain suffix of the current domain (e.g., primary = `example.com`, secondary = `app.example.com`).

**Pros:**

- Registration is clean and predictable (primary domain only).
- Login can work cross-subdomain when domains share a suffix.

**Cons:**

- Does **not** help when domains are entirely unrelated (e.g., `company-a.com` vs `company-b.com`) — the common multi-domain scenario.
- More complex than Option C with a narrow applicability window.

### 8.2 Decision: Option C

Option C (disable passkey registration on non-primary domains with a clear message) is the chosen approach.

**Rationale:**

1. **Lowest risk and complexity** — a small frontend change and a single backend check. No database migrations, no new WebAuthn flows, no risk of breaking existing passkeys.
2. **Eliminates user confusion** — replaces a cryptic browser error with a clear, actionable message pointing to the correct domain.
3. **Future-proof** — when browser support for Related Origin Requests (Option B) matures, it can be layered on top without undoing any work. Option C acts as a clean stepping stone.
4. Option A (domain-scoped passkeys) solves registration but introduces a subtler UX problem (passkeys vanishing across domains) that is arguably worse than a clear "go to the primary domain" message. Option D only helps in the narrow subdomain case.

### 8.3 Implementation Details

#### Backend

- Extend the response of `GET /user/me` (or a suitable existing endpoint) with a boolean field `isPrimaryDomain` that indicates whether the current request's `Host` matches the organization's primary domain.
- The `POST /user/passkey/registration/begin` endpoint should also reject requests from non-primary domains with `403 Forbidden` and a descriptive error message, as a server-side safeguard.

#### Frontend

- In the `PasskeySettings` component, check `isPrimaryDomain`. When `false`:
  - Disable the "Add passkey" button.
  - Display a message: _"Passkey registration is only available on {primaryDomain}. Please visit {link} to set up a passkey."_ where `{primaryDomain}` is a clickable link to the primary domain's preferences page.
- Existing passkeys (registered on the primary domain) continue to be listed in read-only mode (rename/delete remain functional regardless of domain).

#### New i18n Keys

| Key                             | Value                                                                                        |
| ------------------------------- | -------------------------------------------------------------------------------------------- |
| `passkeyRegOnlyOnPrimaryDomain` | Passkey registration is only available on {domain}. Please visit {link} to set up a passkey. |

---

## 9. Admin Considerations

### 9.1 Admin User Management

When an admin views a user's details (admin panel → Users → user detail):

- Show a read-only count of registered passkeys (e.g., "Passkeys: 3").
- Provide an admin action to **remove all passkeys** for a user (useful for account recovery scenarios).
- This calls a new admin endpoint: `DELETE /user/{userId}/passkey/` (requires OrgAdmin role).

### 9.2 Organization Settings

No new organization settings are required. The existing `enforce_totp` setting implicitly covers passkeys (see section 3.3).

---

## 10. Testing Strategy

### 10.1 Backend Unit Tests

Add tests in `server/repository/test/` for the passkey repository (CRUD operations).

Add tests in the router test suite for:

| Test Case                                   | Description                                                  |
| ------------------------------------------- | ------------------------------------------------------------ |
| Registration begin – valid                  | Authenticated user gets valid creation options.              |
| Registration begin – IdP user               | Returns 403.                                                 |
| Registration finish – valid                 | Credential is stored, state is deleted.                      |
| Registration finish – expired state         | Returns 404.                                                 |
| Registration finish – duplicate name        | Returns 400.                                                 |
| Registration finish – rate limit            | Returns 429 after 5 attempts.                                |
| Passwordless login – valid                  | Full ceremony succeeds, tokens returned.                     |
| Passwordless login – disabled user          | Returns 401.                                                 |
| Passwordless login – banned user            | Returns 401.                                                 |
| Passwordless login – clone detection        | Returns error when sign count is suspicious.                 |
| Password + passkey 2FA – valid              | Password login with passkey assertion succeeds.              |
| Password + passkey 2FA – challenge returned | Password login returns 401 with passkey options.             |
| Password + TOTP fallback                    | User with both passkeys and TOTP can fall back to TOTP code. |
| List passkeys                               | Returns passkey metadata (no secrets).                       |
| Delete passkey                              | Removes passkey.                                             |
| Delete last passkey + enforce_totp          | Returns 403.                                                 |
| Rename passkey                              | Updates name.                                                |
| Rename passkey – duplicate name             | Returns 400.                                                 |

### 10.2 Frontend Tests

Add Vitest unit tests for:

- `PasskeySettings` component rendering and interactions.
- Login page passkey button visibility (based on WebAuthn availability).
- TOTP enforcement modal condition update (show when no TOTP and no passkeys).

### 10.3 E2E Tests

Add Playwright tests in `e2e/tests/` covering:

- Passkey registration flow (requires WebAuthn virtual authenticator — Playwright supports this via `cdpSession`).
- Passwordless passkey login.
- Password + passkey second factor login.
- Passkey deletion and enforcement interaction.

---

## 11. Rollout Considerations

### 11.1 Feature Flag

No feature flag is required. Passkeys are opt-in per user (users must explicitly register a passkey). The "Sign in with a passkey" button appears only when the browser supports WebAuthn.

### 11.2 Migration

The database migration (schema version 37) adds the `passkeys` table. No data migration is needed — all users start with zero passkeys.

### 11.3 Backward Compatibility

- Users without passkeys experience no change to their login flow.
- The `POST /auth/login` response is extended but remains backward-compatible (new fields are additive).
- The `GET /user/me` response gains `hasPasskeys: false` by default — existing clients ignore unknown fields.

---

## 12. Summary of Changes by File

### New Files

| File                                      | Description                                 |
| ----------------------------------------- | ------------------------------------------- |
| `server/repository/passkey-repository.go` | Passkey data access layer.                  |
| `server/router/passkey-router.go`         | Authenticated passkey management endpoints. |
| `ui/src/components/PasskeySettings.tsx`   | Passkey management UI component.            |
| `ui/src/types/Passkey.ts`                 | Passkey TypeScript type and API methods.    |

### Modified Files

| File                                         | Change                                                                                                                                                                                                              |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `server/go.mod`                              | Add `github.com/go-webauthn/webauthn` dependency.                                                                                                                                                                   |
| `server/config/config.go`                    | Add `WEBAUTHN_RP_DISPLAY_NAME` and `MAX_PASSKEYS_PER_USER` configuration fields. (`WEBAUTHN_RP_ID` and `WEBAUTHN_RP_ORIGINS` are not used — the RP ID and origin are always derived from the org's primary domain.) |
| `server/repository/db-updates.go`            | Increment `targetVersion` to 37; register `GetPasskeyRepository()`.                                                                                                                                                 |
| `server/repository/auth-state-repository.go` | Add `AuthPasskeyRegistration` (10), `AuthPasskeyLogin` (11), and `AuthPasskey2FA` (12) state types.                                                                                                                 |
| `server/router/auth-router.go`               | Add `/auth/passkey/login/begin` and `/auth/passkey/login/finish` routes; modify `loginPassword` to support passkey as second factor.                                                                                |
| `server/router/user-router.go`               | Add `HasPasskeys` and `IsPrimaryDomain` to `GetUserResponse`; set `IsPrimaryDomain` in `getSelf` by comparing `r.Host` against the org's primary domain; register passkey management routes.                        |
| `server/router/passkey-router.go`            | Add `isRequestFromPrimaryDomain` helper; guard `POST /user/passkey/registration/begin` to return `403 Forbidden` when not on primary domain (spec §8.3).                                                            |
| `server/router/unauthorized-routes.go`       | Whitelist `/auth/passkey/` routes.                                                                                                                                                                                  |
| `ui/src/components/RuntimeConfig.ts`         | Add `hasPasskeys` and `isPrimaryDomain` to `RuntimeUserInfos`; set `isPrimaryDomain` from `GET /user/me` response.                                                                                                  |
| `ui/src/types/User.ts`                       | Add `hasPasskeys` and `isPrimaryDomain` fields; add passkey API methods.                                                                                                                                            |
| `ui/src/pages/login/index.tsx`               | Add "Sign in with a passkey" button; handle passkey 2FA challenge.                                                                                                                                                  |
| `ui/src/pages/preferences.tsx`               | Add `PasskeySettings` component to Security tab.                                                                                                                                                                    |
| `ui/src/components/PasskeySettings.tsx`      | Gate the "Add passkey" UI behind `isPrimaryDomain`; show informational message with link to primary domain when not on primary domain (spec §8.3).                                                                  |
| `ui/src/pages/_app.tsx`                      | Update enforcement check to include `hasPasskeys`.                                                                                                                                                                  |
| `ui/i18n/*.json`                             | Add `passkeyRegOnlyOnPrimaryDomain` translation key (all 13 locales) in addition to all other passkey-related translation strings.                                                                                  |
