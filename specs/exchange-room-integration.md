# Specification: Microsoft Exchange Integration - One-Way Sync

## 1. Overview

This specification details the implementation of "One-Way Sync (Seatsurfing ➔ Exchange)":

- Bookings created, updated, or deleted in Seatsurfing are pushed to the corresponding Exchange Room Calendar.
- Exchange is a read-only destination.
- Bookings made natively in Exchange are **not** synced back to Seatsurfing.

## 2. Multi-Organization Architecture & Authentication

Seatsurfing's multi-organization architecture requires Exchange configurations to be scoped per organization.

**Authentication Method:** App-Only (Client Credentials Flow).

- Each organization's admin will manually provision an Azure AD App Registration in their own M365 tenant with `Calendars.ReadWrite` application permissions restricted via an Application Access Policy (see Section 7).
- Credentials are stored per organization as key-value pairs in the standard `settings` table (see Section 3.1).

## 3. Database Schema

Schema changes are applied via the `RunSchemaUpgrade(curVersion, targetVersion int)` mechanism in each repository's file. The global `targetVersion` constant in `server/repository/db-updates.go` is currently **44**. All repositories are registered in the `repositories` slice inside `RunDBSchemaUpdates()`.

### 3.1 Exchange Settings in the `settings` Key-Value Table

Exchange credentials are **not** stored in a dedicated table. They use the existing `settings` key-value store (one row per `(organization_id, name)`) through the standard `SettingName` constant pattern in `settings-repository.go`:

| Constant                      | Key name                 | Type               |
| ----------------------------- | ------------------------ | ------------------ |
| `SettingExchangeEnabled`      | `exchange_enabled`       | bool               |
| `SettingExchangeTenantID`     | `exchange_tenant_id`     | string             |
| `SettingExchangeClientID`     | `exchange_client_id`     | string             |
| `SettingExchangeClientSecret` | `exchange_client_secret` | string (encrypted) |

**Go struct (`ExchangeSettings`):**

```go
type ExchangeSettings struct {
    Enabled      bool
    TenantID     string
    ClientID     string
    ClientSecret string  // plaintext in memory, encrypted in DB
}
```

**Repository helpers in `SettingsRepository`:**

- `GetExchangeSettings(orgID string) (*ExchangeSettings, error)` — reads all four keys and decrypts the secret.
- `SetExchangeSettings(orgID string, s *ExchangeSettings) error` — writes all four keys, encrypts the secret; if `ClientSecret` is empty the existing encrypted value is preserved.

On read, `DecryptString` is called for `exchange_client_secret`. On write, `EncryptString` is called. Both are in `server/util/encryption.go` and use AES-256-GCM with the `CRYPT_KEY` environment variable — the same pattern used for `auth_providers.client_secret`.

The standard `settings-router.go` handles all read/write access to these keys via `GET /setting/{name}` and `PUT /setting/{name}`. The `exchange_client_secret` key is **masked**: reads return `"1"` when a secret is stored and `""` when none is set; the plaintext is never returned. A write with an empty value is a no-op (preserves the stored secret). Any write to an exchange credential key additionally invalidates the in-memory OAuth2 token cache for that organization.

### 3.2 `room_email` Column on `spaces`

The Exchange Room Mailbox email is stored directly as a column on the `spaces` table rather than in a separate mapping table. It was added in schema version 44:

```sql
ALTER TABLE spaces ADD COLUMN IF NOT EXISTS room_email VARCHAR NOT NULL DEFAULT '';
-- Data migration from the old exchange_space_mapping table (if present):
UPDATE spaces SET room_email = esm.room_email
FROM exchange_space_mapping esm WHERE spaces.id = esm.space_id;
```

The `Space` Go struct has a `RoomEmail string` field. All space repository queries (Create, GetOne, GetAll, GetAllInTime, GetByKeyword, Update) include this column. The `room_email` is exposed via the standard space CRUD REST endpoints — it is part of `CreateSpaceRequest` / `GetSpaceResponse` as `roomEmail`.

### 3.3 `exchange_sync_queue` Table

The persistent job queue for asynchronous sync operations. Polling this table replaces an in-memory queue, which is lost on restart.

```sql
CREATE TABLE IF NOT EXISTS exchange_sync_queue (
    id            UUID NOT NULL DEFAULT uuid_generate_v4(),
    booking_id    UUID NOT NULL,          -- logical reference; not a FK so deleted bookings can still be processed for DELETE operations
    operation     VARCHAR NOT NULL,       -- 'CREATE', 'UPDATE', 'DELETE'
    status        VARCHAR NOT NULL DEFAULT 'pending',  -- 'pending', 'processing', 'failed'
    retry_count   INTEGER NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_error    VARCHAR NOT NULL DEFAULT '',
    payload       TEXT NOT NULL DEFAULT '',  -- JSON snapshot of the data needed to execute the operation (see Section 5.2)
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_exchange_sync_queue_status_retry
    ON exchange_sync_queue (status, next_retry_at)
    WHERE status = 'pending';
```

**Why `payload` as JSON snapshot?** A `DELETE` job must still execute even after the booking row has been deleted. A `UPDATE` job for a booking that is subsequently deleted must be skipped. Storing a JSON snapshot of the required fields (enter, leave, user display name, space name, room email, org credentials reference) at enqueue time decouples the job from the live booking row. See Section 4.5 for the payload schema.

### 3.4 `exchange_booking_mapping` Table

Tracks the Exchange Graph API Event ID for each synced booking. Stored in a separate table to avoid modifying the `bookings` table.

```sql
CREATE TABLE IF NOT EXISTS exchange_booking_mapping (
    booking_id        UUID NOT NULL,   -- not a FK; booking may be deleted
    exchange_event_id VARCHAR NOT NULL,
    room_email        VARCHAR NOT NULL, -- denormalized for DELETE operations after space mapping is removed
    PRIMARY KEY (booking_id)
);
```

**Repository operations required:**

- `Create(bookingID, exchangeEventID, roomEmail string) error`
- `GetByBookingID(bookingID string) (*ExchangeBookingMapping, error)`
- `Delete(bookingID string) error`

## 4. Sync Execution Strategy

The sync runs **asynchronously**. The Seatsurfing HTTP handler creates a booking in the primary DB, enqueues a sync job, and returns immediately. The Exchange API call happens in a background goroutine driven by the existing timer infrastructure.

### 4.1 Integration with the Existing Timer

`app.go` already has a `CleanupTicker` that fires every 60 seconds and calls `onTimerTick()`. The Exchange sync worker is added as a new call within `onTimerTick()`:

```go
// Inside App.onTimerTick():
go GetExchangeSyncWorker().ProcessPendingJobs()
```

`ProcessPendingJobs` must be safe to call concurrently (use a mutex or `sync.Once`-per-tick guard) to prevent overlapping runs if a tick fires while a previous run is still in progress.

### 4.2 Enqueueing Jobs

Three places in `booking-router.go` trigger Exchange sync enqueuing, mirroring the existing CalDAV pattern (`createCalDavEvent`, `updateCalDavEvent`):

| Booking Event                    | Operation enqueued | Trigger location      |
| -------------------------------- | ------------------ | --------------------- |
| Booking created and approved     | `CREATE`           | `onBookingCreated(e)` |
| Booking approved (after pending) | `CREATE`           | `approveBooking(…)`   |
| Booking updated                  | `UPDATE`           | `update(…)` handler   |
| Booking deleted                  | `DELETE`           | `delete(…)` handler   |

If `space.RoomEmail` is empty (no room email configured), **no job is enqueued**. If the organization's `exchange_enabled` setting is `false`, **no job is enqueued**.

For `DELETE`: before the booking row is deleted, the `exchange_booking_mapping` must be read to obtain the `exchange_event_id`. The job payload must include this ID.

### 4.3 Job Processing Logic

`ProcessPendingJobs` runs the following loop:

1. **Claim a batch** of up to N (e.g. 20) jobs where `status = 'pending' AND next_retry_at <= NOW()`. Set `status = 'processing'` atomically using `UPDATE … WHERE id = $1 AND status = 'pending' RETURNING id` to avoid double-processing.
2. For each job, decode the payload JSON.
3. Acquire a valid OAuth2 token for the organization (see Section 6.2).
4. Call the appropriate Graph API endpoint (see Section 6).
5. On **success**:
   - For `CREATE`: insert a row into `exchange_booking_mapping` with the returned `id` from Graph.
   - Delete the queue row.
6. On **failure** (network error, 429, 5xx): apply the retry schedule (see Section 4.4) and set `status = 'pending'` again.
7. On **permanent failure** (4xx other than 429, or `retry_count >= MAX_RETRIES`): set `status = 'failed'` and record the error in `last_error`.

For `UPDATE` jobs: check whether a `exchange_booking_mapping` row exists. If not (meaning the prior `CREATE` job also failed), skip the `UPDATE` — when `CREATE` eventually succeeds or fails permanently, the state will be consistent.

For `DELETE` jobs: if the `exchange_event_id` in the payload is empty (CREATE never succeeded), delete the queue row without calling Graph. If the Graph API returns 404, treat it as success (already deleted).

### 4.4 Retry Schedule (Exponential Backoff)

| `retry_count` before this attempt | Delay before next retry                           |
| --------------------------------- | ------------------------------------------------- |
| 0 (first attempt)                 | immediate (enqueued with `next_retry_at = NOW()`) |
| 1                                 | 1 minute                                          |
| 2                                 | 2 minutes                                         |
| 3                                 | 4 minutes                                         |
| 4                                 | 8 minutes                                         |
| 5 (MAX_RETRIES)                   | mark `status = 'failed'`                          |

Formula: `next_retry_at = NOW() + interval '1 minute' * pow(2, retry_count - 1)` (applied after incrementing `retry_count`).

`MAX_RETRIES = 5` results in a total window of ~15 minutes before a job is considered permanently failed.

**Retriable errors:** any network/transport error, HTTP 429, HTTP 500–599.
**Non-retriable errors (fail immediately):** HTTP 400, 401, 403, 404 (for CREATE/UPDATE), any unrecoverable parse error.

### 4.5 Job Payload JSON Schema

The payload column stores a JSON object to make jobs self-contained:

```json
{
  "orgID": "uuid",
  "bookingID": "uuid",
  "operation": "CREATE|UPDATE|DELETE",
  "roomEmail": "room@contoso.com",
  "exchangeEventID": "AAMk...", // populated for UPDATE and DELETE; empty string for CREATE
  "enter": "2026-09-10T12:00:00Z", // UTC
  "leave": "2026-09-10T14:00:00Z", // UTC
  "locationTZ": "Europe/Berlin", // IANA timezone; see Section 6.3
  "userFirstname": "John",
  "userLastname": "Doe",
  "spaceName": "Room A",
  "locationName": "HQ Floor 3"
}
```

Credentials (`tenant_id`, `client_id`, `client_secret`) are **not** stored in the payload. The worker looks them up fresh via `GetSettingsRepository().GetExchangeSettings(orgID)` at processing time using `orgID`.

## 5. Microsoft Graph API Integration

**Base URL:** `https://graph.microsoft.com/v1.0`
**Auth:** OAuth2 Client Credentials (Bearer token). Token endpoint: `https://login.microsoftonline.com/{tenant_id}/oauth2/v2.0/token`. Required scope: `https://graph.microsoft.com/.default`.

### 5.1 OAuth2 Token Acquisition & Caching

- Use Go's `golang.org/x/oauth2/clientcredentials` package (already a common Go dependency; add to `go.mod` if not present).
- Cache the token in memory per `organization_id` until 60 seconds before its `expires_in`. Use a mutex-protected map `map[string]*oauth2.Token`.
- On token fetch failure (network or 401): mark the job as failed immediately (non-retriable) — a bad credential will not recover on retry.

### 5.2 Event Creation (`CREATE`)

**Endpoint:** `POST /users/{roomEmail}/calendar/events`

**Request body:**

```json
{
  "subject": "[Seatsurfing] Booking: John Doe",
  "body": {
    "contentType": "HTML",
    "content": "<p>Booked by: <b>John Doe</b><br>Space: Room A, HQ Floor 3<br>Time: 10:00–12:00 (Europe/Berlin)</p>"
  },
  "start": {
    "dateTime": "2026-09-10T10:00:00",
    "timeZone": "Europe/Berlin"
  },
  "end": {
    "dateTime": "2026-09-10T12:00:00",
    "timeZone": "Europe/Berlin"
  },
  "showAs": "busy",
  "isReminderOn": false,
  "responseRequested": false
}
```

**On HTTP 201:** parse the `id` field from the response body and insert it into `exchange_booking_mapping`.

### 5.3 Event Update (`UPDATE`)

**Endpoint:** `PATCH /users/{roomEmail}/calendar/events/{exchangeEventID}`

Send the same fields as CREATE (subject, body, start, end, showAs). HTTP 200 = success.

### 5.4 Event Deletion (`DELETE`)

**Endpoint:** `DELETE /users/{roomEmail}/calendar/events/{exchangeEventID}`

HTTP 204 = success. HTTP 404 = treat as success (already deleted). Remove the `exchange_booking_mapping` row on success.

## 6. Timezone Handling

### 6.1 How Booking Times Are Stored

Booking `enter` and `leave` times are stored as UTC in the `bookings` table. The conversion is performed in `booking-router.go`'s `copyFromRestModel` via `GetLocationRepository().AttachTimezoneInformation(m.Enter, location)`.

### 6.2 Resolving the Location Timezone

Each Location has a `Timezone` field (IANA string, e.g. `"Europe/Berlin"`) stored in the `locations.tz` column. If the field is empty, the organization-level default is used. The existing helper `GetLocationRepository().GetTimezone(location *Location) string` encapsulates this fallback logic and must be used.

The timezone string is resolved at **enqueue time** (when the job is created), stored in the payload's `locationTZ` field, and used by the worker without further DB lookups. This avoids issues where the location's timezone changes after a job is queued.

### 6.3 Converting UTC to Local Wall-Clock Time for Graph API

The MS Graph API `dateTimeTimeZone` resource requires a **local wall-clock datetime string** (no `Z` suffix, no UTC offset) paired with an **IANA timezone identifier**.

The worker must:

1. Load the IANA location: `tz, err := time.LoadLocation(payload.LocationTZ)`
2. Convert the UTC time to local: `localEnter := payload.Enter.In(tz)`
3. Format as a naive datetime string: `localEnter.Format("2006-01-02T15:04:05")`
4. Pass the IANA string directly as `timeZone` — the Graph API accepts IANA timezone identifiers.

**Do not** use `util.AttachTimezoneInformationTz` for this purpose. That function modifies the underlying Unix timestamp and is designed for the Seatsurfing iCal/CalDAV round-trip, not for producing a plain local time string.

**Edge case:** If `payload.LocationTZ` is empty (location and org both have no timezone configured), fall back to `"UTC"` and format accordingly.

## 7. User Interface (UI/UX) Updates

### 7.1 Admin Settings (Next.js Frontend)

1. **Integration Settings Page:**
   - A new section under Organization Settings (admin area) called `Microsoft Exchange`.
   - Fields: Enabled toggle, Tenant ID (text), Client ID (text), Client Secret (see below).
   - **"Test Connection" button:** calls `POST /setting/exchange/test` which acquires an OAuth2 token and makes a lightweight Graph API call. Returns 200 on success or a descriptive error. The button is disabled when (a) the Exchange Integration feature flag is off, (b) Exchange is not enabled, or (c) there are **unsaved changes** to any exchange field (`exchangeUnsavedChanges: true`). This last condition ensures the button always tests the persisted settings.
   - **"Save" button:** disabled only when the Exchange Integration feature flag is off. It remains enabled even when the enabled checkbox is unchecked, so that disabling Exchange can be saved.

   **Client Secret field behavior** — mirrors the pattern used for the `clientSecret` field in `admin/settings/auth-providers/[id].tsx`:
   - The component tracks a `clientSecretEditing: boolean` state flag. It is `true` for new records (no saved secret exists yet) and `false` after loading an existing configuration.
   - **When `clientSecretEditing` is `false`** (existing secret stored): render a read-only `<input type="text">` whose `value` is `RendererUtils.SECRET_PLACEHOLDER` (`"••••••••••••••••"`), paired with an edit `<Button>` containing `<IconEdit>`. Clicking the button sets `clientSecretEditing: true`, clears the local `clientSecret` state to `""`, and sets `exchangeUnsavedChanges: true`, switching the field into edit mode with `autoFocus`.
   - **When `clientSecretEditing` is `true`**: render an editable `<Form.Control type="text">` with a `pattern="[^\s]+"` constraint (no whitespace).
   - **On submit**: only include `exchange_client_secret` in the payload when `clientSecretEditing` is `true`. When `clientSecretEditing` is `false`, omit it so the backend retains the existing encrypted value.
   - **Backend contract**: `GET /setting/exchange_client_secret` returns `"1"` when a secret is stored and `""` when none is set — the plaintext is never returned. The UI infers a secret is stored when the value is `"1"`, and sets `exchangeClientSecretEditing: false` accordingly. After a successful save, the frontend resets `clientSecretEditing` to `false` and `exchangeUnsavedChanges` to `false`.
   - **Backend upsert logic**: a write of an empty string to `exchange_client_secret` is a no-op — the existing encrypted value is preserved.

   **Unsaved-changes tracking:** every onChange handler for any exchange field (enabled checkbox, tenant ID, client ID, client secret, clicking the secret edit button) sets `exchangeUnsavedChanges: true`. A successful save resets it to `false`.

2. **Space Configuration:**
   - In the Space setup form, a text field for `Exchange Room Email Address` is included as part of the standard space form. It is saved as the `roomEmail` field via the regular space PUT endpoint — no separate API call is needed.
3. **Sync Error Surface:**
   - A read-only table in the admin area listing `exchange_sync_queue` rows with `status = 'failed'`, showing booking ID, operation, last error, and a "Retry" button that resets `status = 'pending'` and `retry_count = 0`.

### 7.2 Backend API Endpoints

| Method | Path                                  | Description                                                      |
| ------ | ------------------------------------- | ---------------------------------------------------------------- |
| `GET`  | `/setting/{name}`                     | Read a single setting by name (includes all `exchange_*` keys).  |
| `PUT`  | `/setting/{name}`                     | Write a single setting by name (includes all `exchange_*` keys). |
| `POST` | `/setting/exchange/test`              | Tests connectivity to Graph API with stored credentials.         |
| `GET`  | `/setting/exchange/errors/`           | Lists failed sync queue jobs for the org.                        |
| `POST` | `/setting/exchange/errors/{id}/retry` | Resets a failed job to pending.                                  |

All endpoints require the `OrgAdmin` role. The `exchange_*` setting keys are admin-only (not accessible to regular users). The `exchange_client_secret` key is additionally masked on read (returns `"1"` or `""`).

### 7.3 End User UI

No changes required. The booking workflow remains identical.

## 8. Security Considerations

### 8.1 Credential Encryption

`exchange_client_secret` is encrypted at rest using `util.EncryptString()` / `util.DecryptString()` from `server/util/encryption.go`. These functions use AES-256-GCM with the server's `CRYPT_KEY` environment variable — the same mechanism used for `auth_providers.client_secret` and TOTP secrets. The encrypted value is base64-encoded before storage.

The REST API **never** returns the plaintext secret. `GET /setting/exchange_client_secret` returns the string `"1"` when a secret is stored and `""` when none is set. The UI uses this indicator to determine whether to show the placeholder or the edit field.

### 8.2 Least Privilege (Azure AD)

The documentation for organization administrators must instruct them to:

1. Create an Azure AD App Registration with **application permission** `Calendars.ReadWrite` (not delegated).
2. Restrict the permission to specific room mailboxes using an **Application Access Policy** in Exchange Online PowerShell:
   ```powershell
   New-ApplicationAccessPolicy -AppId <ClientID> -PolicyScopeGroupId <RoomMailboxGroup> -AccessRight RestrictAccess
   ```
   This prevents the Seatsurfing service principal from accessing personal user mailboxes.

### 8.3 Data in the Sync Queue

The `exchange_sync_queue.payload` column contains user display names, space names, and room email addresses. It does **not** contain OAuth2 credentials. Access to this table must be limited to the Seatsurfing DB user.
