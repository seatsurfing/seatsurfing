# Feature Spec: Space Kiosk Mode

## Overview

Add a per-space kiosk mode that exposes a lightweight, read-only web page for a single space. The page is intended for wall-mounted tablets, small 7" screens, and e-ink displays placed near a desk or room.

The kiosk page must be available in two presentation variants:

- A monochrome variant for low-power e-ink displays
- A more attractive color variant for LCD displays

The kiosk page must show:

- Space title
- Current availability with strong visual status indication
- Current booking details, if a booking is active
- Next upcoming booking details, if one exists

The kiosk page is addressed by the space UUID and must require authentication before any booking information is returned.

## Goals

- Let admins enable kiosk mode individually per space.
- Provide a stable kiosk URL based on the existing space UUID.
- Show the minimum information needed for a glanceable occupancy display.
- Reuse existing availability semantics where practical.
- Support both color LCD screens and monochrome e-ink screens.
- Allow the kiosk presentation variant to be selected from the URL.

## Non-Goals

- No booking creation, update, approval, or deletion from the kiosk page.
- No location-wide kiosk dashboard in this feature.
- No anonymous/public booking data access.
- No QR code generation requirement in the first version.

## Product Decisions

- The kiosk page must respect the existing organization-level `show_names` privacy setting.
- "Next booking" means the next upcoming booking regardless of date.
- The page is reachable by the existing space UUID, not by a separate kiosk UUID.
- The first version uses a query parameter to select the kiosk presentation variant.
- Kiosk authentication is global per organization, not stored per space.

## Existing System Constraints

### Existing Availability APIs

Seatsurfing already exposes location-scoped availability endpoints:

- `GET /location/{locationId}/space/availability`
- `GET /location/{locationId}/space/{id}/availability`

These endpoints already return:

- Space name
- Current availability
- Booking subject
- Booking owner email, but only when name visibility is allowed
- Booking enter and leave timestamps

However, these endpoints are not suitable as the kiosk contract because:

- They require `locationId` in the route, while kiosk access must be possible from the space UUID alone.
- They are authenticated through normal user or service-account auth, not a kiosk-specific mechanism.
- They return broader availability payloads than the kiosk page needs.

### Existing Service Accounts

Seatsurfing already supports service accounts via HTTP Basic Auth using:

- Username: `{orgId}_{email}`
- Password: service account password

Read-only service accounts may access `GET` endpoints.

This is technically usable for kiosk API retrieval, but it is not a good primary design for kiosk mode because:

- Service accounts are general-purpose accounts, not dedicated kiosk credentials.
- A credential reused across kiosks grants broader read access than required.
- Rotating or revoking kiosk access would be coupled to a user-like account lifecycle.
- Browser-based kiosk provisioning with Basic Auth is awkward and difficult to manage cleanly.

## Authentication Recommendation

### Recommendation

Do not use existing service accounts as the primary kiosk authentication mechanism.

Instead, add a dedicated organization-wide kiosk access credential managed from the admin Settings page.

### Kiosk Credential Model

Each organization gains one kiosk credential with these properties:

- Shared across all kiosk-enabled spaces in that organization
- Set or rotated by admins from the Settings page
- Stored hashed server-side, never retrievable in plaintext after creation
- Used only for kiosk access, not for general API access
- Valid only when the requested space has kiosk mode enabled

### Provisioning Flow

1. An admin enables kiosk mode organization-wide (`kiosk_mode_enabled`) from the Settings page.
2. An admin configures the organization-wide kiosk secret in the Settings page.
3. An admin enables kiosk mode for one or more spaces.
4. For each enabled space, the admin is shown kiosk URLs for the available presentation variants.
5. The admin provisions kiosk devices by opening the URL directly: passing `?secret=<kiosk-secret>` as a query parameter on the first visit stores the secret in the device's `localStorage` and removes it from the URL. Subsequent visits use the stored secret.
6. Subsequent kiosk API calls use the stored secret in an `Authorization: Bearer` header.

### Why This Design

- The visible page address still uses only the space UUID plus an optional variant selector.
- Authentication is separate from addressing.
- Kiosk access is managed centrally from one admin surface.
- A dedicated kiosk secret is still cleaner than reusing service accounts.

### Security Notes

- The kiosk secret is sensitive and must never be returned from normal read APIs after it is saved.
- Rotating the kiosk credential invalidates all existing kiosk sessions for the organization.
- If kiosk mode is disabled, kiosk API access must stop immediately.
- Invalid UUID plus invalid credential combinations must not reveal whether the space exists.
- Because the credential is organization-wide, compromise affects all kiosk-enabled spaces in that organization; this is an accepted tradeoff for centralized administration.

## Data Model Changes

Add the following field to the space entity:

| Field          | Type   | Description                                 |
| -------------- | ------ | ------------------------------------------- |
| `kioskEnabled` | `bool` | Whether kiosk mode is enabled for the space |

Add organization-level settings for kiosk authentication and feature control:

| Setting               | Type     | Visibility    | Description                                                   |
| --------------------- | -------- | ------------- | ------------------------------------------------------------- |
| `kiosk_access_secret` | `string` | admin only    | Bcrypt hash of the organization-wide kiosk secret             |
| `kiosk_mode_enabled`  | `bool`   | admin only    | Whether kiosk mode is active for this organization            |
| `feature_kiosk_mode`  | `bool`   | public (read) | License/feature flag controlling kiosk availability in the UI |

Notes on storage:

- `kiosk_access_secret` stores the bcrypt hash directly under this key — no `_hash` suffix. There is no separate `_updated_at` field.
- `kiosk_mode_enabled` defaults to `false` for all organizations; an admin must explicitly enable it.
- `feature_kiosk_mode` is a deployment-level flag; it is publicly readable so the UI can gate kiosk controls without an authenticated request.

The kiosk secret itself is never returned from normal `GET setting`, `GET space`, or similar read APIs after it has been saved. Reading `kiosk_access_secret` returns `"1"` when a hash is stored, or `""` when none is configured.

These values are stored in the existing settings repository, organization-wide for the credential and org-level toggle, and per-space for kiosk enablement.

## Admin UI Changes

### Space Settings Dialog

Extend the existing per-space settings modal in the location admin page with a new section:

- `Enable kiosk mode` checkbox
- Read-only display of the kiosk page URL for the color variant (with `CopyToClipboard` button)
- Read-only display of the kiosk page URL for the monochrome variant (with `CopyToClipboard` button)
- Hint that kiosk authentication is configured globally in Settings

The kiosk section is hidden entirely when `feature_kiosk_mode` is `false` or when `kiosk_mode_enabled` is `false` for the organization. It belongs in the existing space details dialog, alongside fields like name, enabled, and require subject.

### Settings Page

Extend the admin Settings page with a new kiosk access section:

- `Enable kiosk mode` toggle (`kiosk_mode_enabled`) to activate or deactivate kiosk access organization-wide
- `Kiosk access secret` input; displays `(configured)` when a hash is already stored, empty otherwise
- `Generate random secret` button to produce a cryptographically random alphanumeric secret
- `Save secret` action to set or replace the organization-wide kiosk secret
- Warning that replacing the secret invalidates all existing kiosk sessions in the organization

All kiosk controls (generate, save, copy) are disabled when `feature_kiosk_mode` is `false`.

This section belongs in the existing organization settings flow, not in the per-space dialog.

### Admin Behavior

- Enabling kiosk mode for a space does not create a per-space credential.
- Kiosk mode for a space is effective only when an organization-wide kiosk secret is configured.
- Disabling kiosk mode for a space immediately blocks kiosk access for that space.
- Replacing the global kiosk secret invalidates all existing kiosk sessions across all kiosk-enabled spaces in the organization.

### Permissions

The same admin permissions that can edit a space can manage kiosk mode for that space.

Managing the global kiosk secret belongs to the same permission level that can edit organization settings.

## Public Routes and API Contract

### Web Page Route

Add a new UI route:

- `GET /ui/kiosk/{spaceUuid}`

Query parameters:

- `?variant=color` selects the color LCD-oriented presentation
- `?variant=mono` selects the monochrome e-ink-oriented presentation
- `?lang=<locale>` overrides the display language (e.g. `lang=en`, `lang=de`). The value is persisted in `localStorage` so subsequent visits without the parameter retain the chosen language.
- `?secret=<plaintext-secret>` provisions the kiosk device on first visit: the value is stored in `localStorage` under a space-scoped key and immediately stripped from the URL so it does not appear in browser history. All subsequent API calls use the stored secret.
- If `variant` is omitted, `color` is the default.

Behavior:

- The HTML shell may be delivered without booking data.
- The page must not embed booking data server-side before authentication succeeds.
- The page loads minimal JS/CSS and then retrieves kiosk data through a dedicated API.
- The variant changes presentation only; it does not change the data contract.

### Kiosk API Route

Add a new dedicated endpoint:

- `GET /kiosk/{spaceUuid}/status`

This route is registered outside the normal session-authenticated middleware (unauthorized routes). Authentication is performed by the handler itself.

Authentication:

- Uses `Authorization: Bearer <plaintext-secret>` where the secret is the organization-wide kiosk secret in plaintext
- The handler bcrypt-compares the presented value against the stored hash
- Does not use normal session auth
- Does not require `locationId`

Response shape:

```json
{
  "spaceId": "uuid",
  "spaceName": "Desk A-12",
  "locationId": "uuid",
  "locationName": "First Floor",
  "timezone": "Europe/Berlin",
  "status": "available",
  "currentBooking": null,
  "nextBooking": {
    "id": "uuid",
    "subject": "Sprint Planning",
    "owner": "alice@example.com",
    "ownerVisible": true,
    "enter": "2026-04-16T09:00:00+02:00",
    "leave": "2026-04-16T10:00:00+02:00"
  },
  "refreshedAt": "2026-04-16T08:31:00+02:00"
}
```

Field requirements:

| Field            | Required | Notes                                       |
| ---------------- | -------- | ------------------------------------------- |
| `spaceId`        | yes      | Existing space UUID                         |
| `spaceName`      | yes      | Space title shown on page                   |
| `locationId`     | yes      | For internal reference and future expansion |
| `locationName`   | no       | Optional in UI, useful in payload           |
| `timezone`       | yes      | Use location timezone for formatting        |
| `status`         | yes      | `available` or `occupied`                   |
| `currentBooking` | yes      | `null` when no active booking               |
| `nextBooking`    | yes      | `null` when no upcoming booking exists      |
| `refreshedAt`    | yes      | Timestamp of server evaluation              |

Booking object requirements:

| Field          | Required | Notes                                                                                                   |
| -------------- | -------- | ------------------------------------------------------------------------------------------------------- |
| `id`           | yes      | Booking UUID                                                                                            |
| `subject`      | yes      | Empty string if no subject is set                                                                       |
| `owner`        | yes      | First and last name of the booking owner, or email as fallback; empty string if hidden by privacy rules |
| `ownerVisible` | yes      | Explicit flag so UI can render correct fallback                                                         |
| `enter`        | yes      | Localized to space location timezone                                                                    |
| `leave`        | yes      | Localized to space location timezone                                                                    |

### Booking Selection Rules

- `currentBooking` is the booking active at the current time, if any.
- `nextBooking` is the nearest booking whose start time is greater than now.
- If the current booking is also the only remaining booking today, `nextBooking` is still `null` until a later booking exists.
- Disabled spaces or spaces with kiosk mode disabled must not return kiosk data.

### Privacy Rules

Kiosk mode must follow the existing organization-level `show_names` rule.

If `show_names = false`:

- `ownerVisible = false`
- `owner = ""`
- UI shows a generic label such as `Booked` instead of owner identity

If `show_names = true`:

- `ownerVisible = true`
- `owner` contains the same owner identifier already exposed by existing availability responses

This keeps kiosk mode aligned with current privacy behavior instead of introducing a new disclosure rule.

## UI Specification

### Variants

The kiosk page has two presentation variants with the same underlying data:

#### Color Variant

- Intended for LCD and other color displays
- Uses a more attractive visual treatment with stronger color surfaces and richer spacing
- Must still not rely on color alone to communicate occupancy state

#### Monochrome Variant

- Intended for low-power e-ink and other monochrome displays
- Uses a very minimalistic layout with high contrast and minimal decorative styling
- Must minimize visual noise, gradients, heavy fills, and non-essential UI elements
- Must remain legible during slower-refresh display updates

### Layout

The kiosk page must use a single-screen, glanceable layout with three vertical zones:

1. Header
2. Current status block
3. Current and next booking blocks

On 7" screens, the default layout is a single column.

On larger screens, the booking blocks may align in two columns if readability improves, but the status block remains dominant.

### Content

#### Header

- Space title in large type
- Optional smaller location name

#### Status Block

- Large status label: `Available` or `Occupied`
- Strong background or border treatment
- Last refresh time in smaller text

#### Current Booking Block

- Heading: `Now`
- Subject, if present
- Owner, when visible by privacy rules
- Time range
- Fallback text when there is no active booking

#### Next Booking Block

- Heading: `Next`
- Subject, if present
- Owner, when visible by privacy rules
- Time range
- Fallback text when there is no upcoming booking

### Visual Design Constraints

The page must work well on both color LCD and monochrome e-ink devices.

Requirements:

- Do not rely on color alone to communicate state.
- `Occupied` must use both red styling and explicit text in the color variant.
- `Available` must use both green styling and explicit text in the color variant.
- Monochrome readability must remain clear through contrast, border weight, iconography, and text labels.
- Avoid animations, skeletons, pulsing indicators, or continuously moving UI.
- Avoid low-contrast gray text.
- Prefer large typography and simple blocks over dense tables.
- The monochrome variant should use fewer decorative elements and a smaller CSS/asset footprint than the color variant.

### Responsive Target

The page must remain usable at:

- 800x480
- 1024x600
- Larger tablet and desktop displays

No horizontal scrolling is allowed in the default kiosk presentation.

## Refresh Behavior

The kiosk page should auto-refresh data on a fixed interval.

Recommended initial behavior:

- Poll every 60 seconds
- Refresh immediately when the browser tab regains focus
- Show the last successful refresh timestamp

Rationale:

- Frequent enough for occupancy signage
- Conservative enough for e-ink devices and low-power tablets

If later needed, the interval can be made configurable, but the first version does not require a per-space refresh setting.

## Error States

The kiosk page must have explicit states for:

- Invalid or missing kiosk credential
- Kiosk mode disabled for this space
- Space not found
- Temporary network failure
- Server error

UI requirements:

- Do not leak booking details on auth failure.
- Use a terse fullscreen error message suitable for unattended devices.
- Retry automatically only for transient network/server errors, not for auth errors.

Suggested HTTP behavior:

- `401` for missing or invalid kiosk credential
- `404` when the space UUID does not exist or kiosk mode is disabled

Returning `404` for disabled kiosk mode avoids distinguishing between "space exists but kiosk disabled" and "space does not exist".

## Backend Implementation Notes

### Space Lookup

The kiosk API must resolve the space directly from the provided space UUID, then derive location and organization context from the space record.

### Booking Data Source

The kiosk API should reuse the same repository logic used by existing space availability retrieval where practical, but it must return a kiosk-specific payload instead of reusing the existing location-scoped REST contract verbatim.

### Authorization Checks

The kiosk API validates in this order to avoid leaking information about space existence:

1. `Authorization: Bearer` header is present — returns `401` immediately if missing or malformed
2. Space exists — returns `404` if not found
3. `space.kioskEnabled` is `true` — returns `404` if not
4. Location associated with the space exists — returns `404` if not
5. `kiosk_access_secret` is configured and bcrypt matches the presented secret — returns `401` on mismatch or if no secret is configured
6. `feature_kiosk_mode` is enabled — returns `404` if not
7. `kiosk_mode_enabled` is `true` for the organization — returns `404` if not

Returning `404` for disabled kiosk mode (steps 3, 6, 7) avoids distinguishing between "space exists but kiosk disabled" and "space does not exist" to any caller that has not yet authenticated.

The kiosk credential alone is the authorization scope; it must not grant access to any non-kiosk APIs. It may be reused across kiosk-enabled spaces in the same organization by design.

## OpenAPI Changes

Update [specs/openapi.yaml](specs/openapi.yaml) to add:

- The new kiosk endpoint
- A kiosk auth security scheme, if represented separately from normal bearer auth
- The kiosk response schema
- Error responses for auth failure and disabled/not-found cases

The UI page route itself does not need to appear in OpenAPI.

## Testing

### Backend Tests

Add tests for:

1. Kiosk endpoint returns `401` without credential
2. Kiosk endpoint returns `401` with invalid credential
3. Kiosk endpoint returns `404` when space does not exist
4. Kiosk endpoint returns `404` when kiosk mode is disabled
5. Kiosk endpoint returns `200` with valid credential
6. Response shows `available` when no current booking exists
7. Response shows `occupied` when a current booking exists
8. `currentBooking` is selected correctly
9. `nextBooking` is selected correctly across day boundaries
10. Owner is hidden when `show_names = false`
11. Owner is returned when `show_names = true`
12. Replacing the global kiosk credential invalidates the old credential immediately for all kiosk-enabled spaces in the organization

### Frontend Tests

Add tests for:

1. Kiosk page renders space title and status
2. Kiosk page renders current booking fallback when none exists
3. Kiosk page renders next booking fallback when none exists
4. Owner label is hidden or replaced when `ownerVisible = false`
5. Status presentation is readable without relying on color only in the color variant
6. Monochrome variant renders a reduced, high-contrast layout
7. `?variant=mono` selects the monochrome presentation
8. `?variant=color` or no variant selects the color presentation
9. Polling refresh updates displayed content

### E2E Tests

Add at least one end-to-end scenario that:

1. Sets the global kiosk secret in Settings
2. Enables kiosk mode for a space
3. Opens the color kiosk URL and verifies booking data loads
4. Opens the monochrome kiosk URL and verifies the alternate presentation loads
5. Replaces the global kiosk secret
6. Verifies the old kiosk session stops working after reload

## Files Changed During Implementation

| File                                         | Change                                                                           |
| -------------------------------------------- | -------------------------------------------------------------------------------- |
| `server/repository/space-repository.go`      | Added `KioskEnabled` field; DB migration adds `kiosk_enabled` column             |
| `server/repository/settings-repository.go`   | Added `SettingKioskSecret`, `SettingKioskModeEnabled`, `SettingFeatureKioskMode` |
| `server/repository/booking-repository.go`    | Added `GetCurrentAndNextBySpaceID`; `KioskBookingEntry` with name fields         |
| `server/repository/db-updates.go`            | DB migration version 41: adds `kiosk_enabled` column to spaces table             |
| `server/router/kiosk-router.go`              | New router; handles `GET /kiosk/{id}/status`                                     |
| `server/router/unauthorized-routes.go`       | Registers `/kiosk/` prefix as an unauthenticated route                           |
| `server/router/settings-router.go`           | Allowlists kiosk settings; bcrypt hashing on write; masking on read              |
| `server/router/space-router.go`              | Exposes `kioskEnabled` in space REST model                                       |
| `server/router/test/kiosk-router_test.go`    | New test file; 9 kiosk handler tests                                             |
| `server/router/test/settings-router_test.go` | Updated expected setting counts for public and admin read tests                  |
| `specs/openapi.yaml`                         | Documents kiosk endpoint, auth scheme, response schema                           |
| `ui/src/components/RuntimeConfig.ts`         | Added `featureKioskMode` and `kioskModeEnabled` fields                           |
| `ui/src/types/Space.ts`                      | Added `kioskEnabled` field to space model                                        |
| `ui/src/pages/admin/locations/[id].tsx`      | Kiosk section in space modal with URLs and copy buttons                          |
| `ui/src/pages/admin/settings/index.tsx`      | Kiosk section with org-enable toggle, secret management                          |
| `ui/src/pages/kiosk/[id].tsx`                | New kiosk page (color + mono variants, lang/secret/variant params)               |
| `ui/src/util/Validation.ts`                  | Added `excludeSpecial` param to `generatePassword`                               |
| `ui/i18n/translations.*.json` (all 14 files) | Added kiosk translation keys                                                     |

## Acceptance Criteria

The feature is complete when all of the following are true:

1. An admin can enable kiosk mode for a specific space from the existing space settings dialog.
2. An admin can configure a global kiosk secret from the Settings page.
3. Each kiosk-enabled space exposes URLs based on the space UUID for both the color and monochrome variants.
4. The kiosk page shows current availability, current booking, and next booking.
5. The kiosk page works on small 7" screens without horizontal scrolling.
6. The monochrome variant remains understandable on e-ink displays.
7. Booking data is not returned without valid kiosk authentication.
8. Owner visibility follows the existing `show_names` privacy setting.
9. Replacing the global kiosk secret invalidates prior kiosk access immediately.
