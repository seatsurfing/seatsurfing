# Feature Spec: User-Scoped Booking Operations in the Plugin Host API

## Overview

Plugins run as separate processes and talk to the core through the gRPC host API (`server/api/hostapi.go`). Until now, that API offered only single-entity lookups for bookings, spaces and locations. A plugin could not search for free spaces, create a booking or list a user's bookings without re-implementing the booking rules.

This spec adds five **user-scoped** host API methods. Each method acts as a given user and runs through the same code path as the matching REST endpoint. A plugin acting for a user therefore gets exactly that user's permissions and booking restrictions. This lets plugins offer booking functionality on behalf of users without duplicating the booking rules.

It also forwards the request's `Host` and `RemoteAddr` to plugins.

## Goals

- Plugins can search locations and spaces, create bookings and list upcoming bookings on behalf of a user.
- All booking rules keep a single implementation, shared by the REST API and the host API:
  - maximum bookings, advance days and duration limits
  - allowed booker groups
  - bookable weekdays
  - conflicts
  - approval
  - disabled spaces and locations
  - timezone handling
  - follow-up actions: mails, CalDAV and plugin hooks
- Plugins can tell which host (organization domain) a forwarded HTTP request was sent to.

## Non-Goals

- Booking on behalf of another user through the host API. `POST /booking/` with `userEmail` stays REST-only.
- Updating or deleting bookings through the host API.

## Refactoring in `server/router`

The REST handlers become thin wrappers around exported functions:

| Function                                                                                             | Extracted from                          |
| ---------------------------------------------------------------------------------------------------- | --------------------------------------- |
| `(*BookingRouter).CreateBookingForUser(user, *CreateBookingRequest) (*Booking, *BookingCreateError)` | `POST /booking/`                        |
| `GetUpcomingBookingsForUser(user) ([]*BookingDetails, error)`                                        | `GET /booking/`                         |
| `(*SpaceRouter).GetSpaceAvailabilityForUser(user, location, spaceID, enter, leave, attributes)`      | `GET /location/{id}/space/availability` |
| `(*LocationRouter).SearchLocationsForUser(user, *SearchLocationRequest)`                             | `POST /location/search`                 |

`BookingCreateError` carries the HTTP status and the `X-Error-Code` value (`ResponseCode*`). `BookingCreateError.Send(w)` writes it the way the handler always has, so REST responses are unchanged.

## Host API Additions

| Method                                                                                                                     | Behaves like                                                                               |
| -------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| `SearchLocationsForUser(userID, enter, leave, []SearchAttributeFilter) ([]*LocationInfo, error)`                           | `POST /location/search`, plus location attribute values and an `Allowed` flag for the user |
| `GetSpaceAvailabilityForUser(userID, locationID, enter, leave, []SearchAttributeFilter) ([]*SpaceAvailabilityInfo, error)` | `GET /location/{id}/space/availability`, without other users' bookings                     |
| `GetSpaceAttributes(organizationID) ([]*SpaceAttributeDefinition, error)`                                                  | `GET /space-attribute/`                                                                    |
| `CreateBookingForUser(userID, spaceID, enter, leave, subject) (*BookingCreateResult, error)`                               | `POST /booking/` for the user themselves                                                   |
| `GetUpcomingBookingsForUser(userID) ([]*BookingDetails, error)`                                                            | `GET /booking/`                                                                            |

Rules:

- **Active users only.** The user must exist and not be disabled, the same requirement `VerifyAuthMiddleware` applies to REST calls. Otherwise the method returns an error.
- **Own organization only.** Locations and spaces of other organizations are rejected: an error for availability, and `403` in `BookingCreateResult` for booking creation.
- **Validation failures are results, not errors.** When `CreateBookingForUser` rejects a booking, it returns `StatusCode` and `ErrorCode` in `BookingCreateResult` (for example `409`/`1001` for a slot conflict) and a nil error. The error return is reserved for transport failures and unknown or disabled users.
- **Times are wall-clock times.** Their timezone is replaced by the location's, as the REST API does. Callers build them in UTC, which is what survives the protobuf `Timestamp` round trip.
- **Filters are validated** with the same rules as the REST request bodies: valid attribute IDs and comparators only.

The new wire messages are additive (`hostapi.proto`), so existing plugins remain compatible.

## Forwarded Request Metadata

`api.PluginHTTPRequest` gains two fields, `Host` and `RemoteAddr`, filled from `http.Request.Host` and `http.Request.RemoteAddr` (`plugin.proto` `HttpRequest` fields 7 and 8). Go does not keep `Host` in the header map, so plugins had no way to see it before.

## Tests

- **`server/app/test/hostapi-booking_test.go`** calls the real host API implementation (`app.NewHostAPI()`) and covers:
  - creating and listing bookings
  - slot conflicts
  - the maximum-bookings limit
  - invalid durations
  - disabled spaces
  - allowed booker groups for spaces and locations
  - approval
  - availability and attribute filters
  - invalid comparators
  - location search, including `numFreeSpaces`
  - other organizations
  - disabled and unknown users
- **`server/api/hostapi_booking_test.go`**: protobuf round-trip tests for the new messages.
- **Regression:** the existing booking, space and location router tests cover the refactored REST handlers.
