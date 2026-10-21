# Feature Spec: Booking Attributes

## Overview

Organizations can define up to three custom booking attributes (label, type, required/optional). In this first step, the attributes are captured on the public booking page and stored with the resulting booking. Approvers see the values in the approvals list, and admins see them in the booking details.

## Goals

- Let admins define up to three booking attributes per organization.
- Support the types single-line text, multi-line text, number and checkbox.
- Let admins mark each attribute as required or optional.
- Capture the attribute values when a public booking is requested.
- Show the values in the approvals list and in the admin booking details.

## Non-Goals

- No attributes for regular (authenticated) bookings yet.
- No editing of captured values after the booking was requested.
- No attribute values in notification emails, the public booking details page, exports of the booking list or the plugin API.

## Product Decisions

- The limit of three attributes is enforced by the server, per organization.
- Attributes are managed in the "Public booking" section of the organization settings and require `org_settings` admin permission.
- Values are validated when the booking is requested (before the confirmation email is sent) and stored once the booker confirms the request.
- Values are a snapshot taken when the booking is stored: deleting, renaming, retyping or reordering an attribute does not change the values captured for existing bookings. Each value keeps the label, type and position of its attribute at the time it was captured.
- Empty optional values are not stored. Checkbox values are stored as `1` (checked) or `0` (unchecked); a required checkbox must be checked.
- Numbers must be positive integers (greater than zero, no decimals).

## Existing System Constraints

- Public bookings are requested via `POST /public-booking/{orgId}/request` and stored in an `AuthState` (`PublicBookingRequestPayload`) until confirmed via `POST /public-booking/confirm/{id}`, which creates the `bookings` and `public_bookings` rows (`server/router/public-booking-router.go`).
- Pending approvals are served by `GET /booking/pendingapprovals/` and single bookings by `GET /booking/{id}`, both mapped by `BookingRouter.copyToRestModel`.
- Public bookings cannot be edited (`PUT /booking/{id}` returns 403 for them).

## Data Model Changes

```sql
CREATE TABLE IF NOT EXISTS booking_attributes (
  id uuid DEFAULT uuid_generate_v4(),
  organization_id uuid NOT NULL,
  label VARCHAR NOT NULL,
  type INTEGER NOT NULL,             -- 1 = text, 2 = multi-line text, 3 = number, 4 = checkbox
  required boolean NOT NULL DEFAULT FALSE,
  position INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (id));

CREATE TABLE IF NOT EXISTS booking_attribute_values (
  booking_id uuid NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
  attribute_id uuid NOT NULL,        -- no foreign key: values outlive deleted attributes
  label VARCHAR NOT NULL,            -- attribute label at the time the value was stored
  type INTEGER NOT NULL,             -- attribute type at the time the value was stored
  position INTEGER NOT NULL DEFAULT 0, -- attribute position at the time the value was stored
  value VARCHAR NOT NULL DEFAULT '',
  PRIMARY KEY (booking_id, attribute_id));
```

Both tables are created on startup; no schema version bump is needed. `booking_attributes` rows are removed when the organization is deleted, values cascade with their booking only. A value is only stored if the attribute belongs to the booking's organization. The attribute's label, type and position are copied into `booking_attribute_values` when the value is stored, so values are shown as they were captured, even after the attribute has been changed or deleted.

`PublicBookingRequestPayload` gets an `attributes` map (attribute ID → value).

## API Changes

### Attribute management (`org_settings` admin)

| Method   | Path                           | Response                |
| -------- | ------------------------------ | ----------------------- |
| `GET`    | `/booking-attribute/`          | 200, list of attributes |
| `POST`   | `/booking-attribute/`          | 201, `X-Object-ID`      |
| `PUT`    | `/booking-attribute/{id}`      | 204                     |
| `PUT`    | `/booking-attribute/{id}/move` | 204                     |
| `DELETE` | `/booking-attribute/{id}`      | 204                     |

Request body: `{"label": "Company", "type": 1, "required": true}`. Creating a fourth attribute returns 400 with `X-Error-Code: 1015`. New attributes are appended at the end; `PUT /booking-attribute/{id}/move` with `{"direction": "up"}` or `{"direction": "down"}` swaps the attribute with its neighbour (no-op at the first/last position) and renumbers `position`. The list, the public booking form and the stored values are ordered by `position`. Attributes of another organization return 404.

### Public booking

- `GET /public-booking/{orgId}/spaces` additionally returns `attributes: [{id, label, type, required}]`.
- `POST /public-booking/{orgId}/request` accepts `attributes: {"<attributeId>": "<value>"}`. Missing required values, unknown attribute IDs or invalid values return 400 with `X-Error-Code: 1014`. Limits: single-line text 256 characters without line breaks, multi-line text 4096 characters, numbers must be positive integers matching `^0*[1-9][0-9]*$` (at most 32 characters, no sign, decimals or exponents), checkboxes accept `1`/`true` (checked) or `0`/`false`/empty (unchecked).

### Bookings

`GET /booking/{id}` and `GET /booking/pendingapprovals/` return `attributes: [{attributeId, label, type, value}]` for public bookings with stored values (`label` and `type` are those stored with the value; values are ordered by the stored position); the field is omitted otherwise.

## UI Changes

- Settings → Public booking: table of attributes with add/edit/delete (modal), add button disabled once three attributes exist.
- Public booking page: one input per attribute below the subject (text input, textarea, number input or checkbox), optional fields are marked as such.
- Approvals: an info button in the last column before approve/decline, shown for bookings with attribute values, opens a modal listing them.
- Booking details: read-only field per stored attribute value. Here and in the approvals list, values are rendered according to their stored type (checkbox values as yes/no), even if the attribute's type was changed afterwards.

## Integration Points

- Business rules live in `server/service/booking-attribute-service.go` (`BookingAttributeService`), errors are returned as `BookingError` and mapped by `sendBookingError`.
- Multi-tenancy: management endpoints act on the caller's organization only; values are joined via the booking's `organization_id`.
