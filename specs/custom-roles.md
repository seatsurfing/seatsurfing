# Feature Spec: Custom Roles & Permissions

## Overview

Replace Seatsurfing's fixed, ordered role ladder (`UserRoleUser` … `UserRoleSuperAdmin`) with an organization-scoped role model in which administrators define named roles and, per functionality, choose an access level. A user may hold any number of roles; their effective access to a functionality is the highest level any of their roles grants.

Existing installations keep exactly the access they have today: the schema upgrade seeds built-in roles that reproduce the current ladder and assigns them according to each user's legacy `role` value.

## Goals

- Let organization administrators define custom roles with a per-functionality access level.
- Allow a role to grant, for example, group administration without also granting users, settings and auth providers (issue #2470).
- Allow service accounts to be granted a narrow permission set instead of implicit organization-admin access.
- Preserve the effective access of every existing user across the upgrade.
- Make it impossible for an organization to lock itself out of administrative access.
- Let plugins contribute their own permissions to the same model.
- Provide the role-assignment substrate that OIDC claim mapping (issues #1447, #2353) builds on.

## Non-Goals

- No scoped assignments in this version (for example "floor plan administrator for Building A only"). The `user_roles` table reserves columns for it.
- No group-scoped booking _policies_ — advance-booking windows, collective quotas, group-exclusive spaces (issues #2090, #1365, #557). A permission governs endpoint access; a policy governs limits once access is granted. These are separate features.
- No per-endpoint or per-HTTP-method custom permissions. The catalogue is a fixed set of functionalities.
- No delegation of role management to a scope narrower than the organization.

## Product Decisions

- **Roles are organization-scoped.** There are no cross-organization roles.
- **Assignment is many-to-many**, with the effective level being the maximum granted by any assigned role.
- **Levels are declared per permission.** A uniform four-level ladder would present administrators with choices that do nothing, so each permission enumerates only the levels meaningful for it. Most enumerate `none` and `admin` only.
- **Super admin is removed from the product.** It was an operator concept that leaked into the user role ladder; self-hosted installations never created one.
- **Service accounts become an account type, not a role.** This removes the defect whereby `UserRoleServiceAccountRO/RW` (21/22) sort above `UserRoleOrgAdmin` (20) and therefore satisfy `IsOrgAdmin`.
- **Baseline access is not configurable.** Every authenticated, non-disabled user can always manage their own bookings, buddies, preferences, profile and MFA, and can always read locations, spaces, attributes and availability — without these, the booking UI cannot function.
- **The organization can never be left without an administrator.** Enforced transactionally on every mutation, with a startup repair as a backstop.

## Existing System Constraints

### Current role model

`server/api/entities.go`:

```go
type UserRole int

const (
	UserRoleUser             UserRole = 0
	UserRoleSpaceAdmin       UserRole = 10
	UserRoleOrgAdmin         UserRole = 20
	UserRoleServiceAccountRO UserRole = 21
	UserRoleServiceAccountRW UserRole = 22
	UserRoleSuperAdmin       UserRole = 90
)
```

`User.Role` is a single ordered integer. Predicates in `server/repository/user-repository.go` are `>=` comparisons:

```go
func (r *UserStore) IsSpaceAdmin(user *User) bool { return int(user.Role) >= int(UserRoleSpaceAdmin) }
func (r *UserStore) IsOrgAdmin(user *User) bool   { return int(user.Role) >= int(UserRoleOrgAdmin) }
func (r *UserStore) IsSuperAdmin(user *User) bool { return int(user.Role) >= int(UserRoleSuperAdmin) }
```

### Current authorization helpers

`server/router/routes.go` exposes three helpers, used at approximately 113 call sites across the routers:

- `CanAccessOrg(user, orgID)` — same organization, or super admin.
- `CanSpaceAdminOrg(user, orgID)` — same organization and `IsSpaceAdmin`, or super admin.
- `CanAdminOrg(user, orgID)` — same organization and `IsOrgAdmin`, or super admin.

`GetRequestUser(r)` loads the user from the database on each call (cache-backed). Authorization never trusts the JWT: `handleTokenAuth` reads only `claims.UserID` and `claims.SessionID`, then re-loads the user.

### Service account authentication

`handleServiceAccountAuth` (`server/router/routes.go`) accepts HTTP Basic (`{orgID}_{email}`) and opaque bearer API tokens. It requires `Role` to be `UserRoleServiceAccountRO` or `UserRoleServiceAccountRW`, and rejects non-`GET` requests for the read-only variant. That method check is the only write protection for read-only service accounts and is retained unchanged.

### Approval narrowing

Booking approval already has an orthogonal narrowing mechanism independent of roles — `isValidApproverForSpace` (`server/router/booking-router.go`) requires the approver to be a member of one of the space's approver groups, when any are configured. Access levels add nothing here, so the `approvals` permission is `none`/`admin` only.

### Schema migrations

A single global integer version, orchestrated by `RunDBSchemaUpdates()` in `server/repository/db-updates.go` (currently `targetVersion = 52`). Each repository implements `RunSchemaUpgrade(curVersion, targetVersion int)` with sequential `if curVersion < N` blocks using idempotent DDL. Base tables are created lazily in each `Get*Repository()` singleton.

The precedent for this feature's backfill is the version-13 upgrade in `user-repository.go`, which converted the `org_admin`/`super_admin` booleans into the integer `role` column.

### Plugins

Plugins are standalone gRPC servers implementing `api.SeatsurfingPlugin`, dialed by the host from `PLUGINS_CONFIG`. The host authenticates the request and forwards only `UserID` in `api.PluginHTTPRequest`; the plugin resolves authorization itself through the `HostAPI` callback interface, which today exposes `IsOrgAdmin` and `IsSuperAdmin` (but not `IsSpaceAdmin`). Both first-party plugins carry a private copy of `CanAdminOrg` in `plugin-helpers.go`.

Plugin admin-UI menu items declare `Visibility: "admin" | "spaceadmin"`, gated by `settings-router.go` behind `CanSpaceAdminOrg`.

## Data Model Changes

Schema version 53.

```sql
CREATE TABLE IF NOT EXISTS roles (
  id              uuid DEFAULT uuid_generate_v4(),
  organization_id uuid NOT NULL,
  name            VARCHAR NOT NULL,
  description     VARCHAR NOT NULL DEFAULT '',
  system          BOOLEAN NOT NULL DEFAULT FALSE,
  PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_org_name ON roles (organization_id, LOWER(name));

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id    uuid NOT NULL,
  permission VARCHAR NOT NULL,
  level      INT NOT NULL,
  PRIMARY KEY (role_id, permission)
);

CREATE TABLE IF NOT EXISTS user_roles (
  user_id    uuid NOT NULL,
  role_id    uuid NOT NULL,
  scope_type VARCHAR NOT NULL DEFAULT '',
  scope_id   VARCHAR NOT NULL DEFAULT '',
  source     VARCHAR NOT NULL DEFAULT 'manual',
  PRIMARY KEY (user_id, role_id, scope_type, scope_id)
);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (user_id);
```

`scope_type` and `scope_id` are reserved for a later scoped-assignment feature and are always empty in this version. Empty strings rather than `NULL` keep them usable in the primary key.

`source` distinguishes `manual` assignments from those reconciled from an identity provider, so that OIDC reconciliation can revoke only what it granted.

Unknown `permission` values are tolerated on read and resolve to no access, so that a role referencing an offline plugin's permission does not fail to load.

```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS account_type INT NOT NULL DEFAULT 0;
```

`account_type`: `0` = person, `1` = service account.

The `users.role` column is retained but no longer consulted for authorization. It is dropped in a later schema version, once the new model has shipped.

### Backfill

For each organization, four roles are created:

| Role                       | `system` | Permissions                                                                                                                            | Assigned to legacy role |
| -------------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- |
| Organization Administrator | yes      | every permission at `admin`                                                                                                            | 20, and 90              |
| Floor Plan Administrator   | no       | `areas`, `space_attributes`, `bookings`, `approvals` at `admin`; `analytics`, `presence_report` at `read`; `users`, `groups` at `read` | 10                      |
| API access                 | no       | every permission at `admin`                                                                                                            | 21 and 22               |

Both service account kinds receive the same permissions. The read-only
restriction is enforced by the HTTP method check in
`handleServiceAccountAuth`, not by permission level: granting a lower level
would revoke access such accounts have today, since several endpoints a
read-only account can currently reach are organization-admin gated.

Users with legacy role `0` receive no assignment; baseline access covers them. Users with role `21` or `22` additionally get `account_type = 1`.

Former super administrators become organization administrators of their own organization; each conversion is logged.

Only the Organization Administrator role is marked `system`. The other three are ordinary roles that administrators may edit or delete afterwards — reproducing today's behaviour is a starting point, not a constraint.

## Permission Catalogue

Levels: `none = 0`, `read = 10`, `write = 20`, `admin = 30`.

| Key                | Levels                   | Functionality                                                             |
| ------------------ | ------------------------ | ------------------------------------------------------------------------- |
| `areas`            | none, admin              | Locations, floor plans, maps, spaces, approvers, allowed bookers          |
| `space_attributes` | none, admin              | Space and location attributes                                             |
| `bookings`         | none, read, admin        | Other users' bookings; `read` grants visibility without modification      |
| `approvals`        | none, admin              | Approving pending bookings (further narrowed by approver groups)          |
| `analytics`        | none, read               | Aggregate utilization statistics and dashboard tiles                      |
| `presence_report`  | none, read               | Who was present and when - personal data, hence separate from `analytics` |
| `users`            | none, read, admin        | User CRUD, password / TOTP / passkey reset                                |
| `groups`           | none, read, write, admin | `write` grants membership management only; `admin` adds create and delete |
| `roles`            | none, read, admin        | Managing roles and their assignment                                       |
| `service_accounts` | none, admin              | Service accounts and API tokens                                           |
| `auth_providers`   | none, admin              | OAuth2 / OIDC authentication providers                                    |
| `org_settings`     | none, admin              | Booking settings, organization profile, domains, feature toggles          |
| `audit_log`        | none, read               | Authentication attempts                                                   |

Plugins register additional keys under a `plugin.` prefix.

### Organization-wide restrictions

The settings `hide_reports` and `hide_stats` are enforced after the permission
check and withhold the presence report and the utilization statistics
respectively, from every user including organization administrators. They
express something a permission cannot: an administrator can always grant
themselves a permission, so only an organization-level setting can state that
nobody evaluates presence data - the form a works council agreement takes.

They are not a security boundary, since `org_settings` at full access can
switch them off, but that is a deliberate act rather than a side effect of a
role edit. A withheld endpoint answers 404, where a missing permission answers 403.

### Baseline access

Granted unconditionally to every authenticated, non-disabled user and not represented in the catalogue: own bookings (create, read, update, delete), recurring bookings, buddies, own preferences, own profile, own MFA and passkeys, and read access to locations, spaces, space attributes and availability.

## API Changes

New router at `/role/`:

| Method | Path                | Required                                                                                         |
| ------ | ------------------- | ------------------------------------------------------------------------------------------------ |
| GET    | `/role/permissions` | authenticated — returns the catalogue, including plugin-registered keys and their allowed levels |
| GET    | `/role/`            | `roles:read`                                                                                     |
| GET    | `/role/{id}`        | `roles:read`                                                                                     |
| POST   | `/role/`            | `roles:admin`                                                                                    |
| PUT    | `/role/{id}`        | `roles:admin`                                                                                    |
| DELETE | `/role/{id}`        | `roles:admin`                                                                                    |
| GET    | `/role/{id}/users`  | `roles:admin`                                                                                    |

Extended on the existing user router:

| Method | Path                     | Required                                        |
| ------ | ------------------------ | ----------------------------------------------- |
| GET    | `/user/{id}/roles`       | `roles:read`                                    |
| PUT    | `/user/{id}/roles`       | `roles:admin`                                   |
| GET    | `/user/{id}/permissions` | `roles:read`, or the requesting user themselves |

`GET /user/me` gains a `permissions` object mapping permission key to effective level, and loses `superAdmin`.

New error codes in the `7xxx` range (roles), defined in `server/router/routes.go`:

- `7001` — the change would leave the organization without an administrator
- `7002` — cannot grant a level exceeding your own
- `7003` — cannot modify a system role
- `7004` — cannot remove your own role-management access

Removed endpoints (super admin): `GET /organization/`, `POST /organization/`. Cross-organization behaviour is removed from `PUT /user/{id}`, `POST /user/`, the API-token endpoints and `GET /booking/report/presence`.

## UI Changes

- New `pages/admin/roles/index.tsx` (role list) and `pages/admin/roles/[id].tsx` (permission matrix editor). The editor renders one row per permission, with a control offering only that permission's declared levels.
- `pages/admin/users/[id].tsx` replaces the single-role select with a multi-select of roles.
- `pages/admin/organizations/*` is deleted along with its sidebar entry.
- `RuntimeConfig.INFOS` replaces `superAdmin` / `spaceAdmin` / `orgAdmin` with a `permissions` map, plus a `hasPermission(key, level)` helper.
- Sidebar entries, the booking-UI admin link, the post-login redirect guard and the dashboard tiles all switch to permission checks.
- A `withPermission(Page, key, level)` HOC provides page-level guards, which admin pages currently lack entirely.
- New translation keys in `i18n/translations.en-GB.json`, propagated by `add-missing-translations.sh`.

## Integration Points

### Authorization

`server/api/permissions.go` defines the `Permission` and `PermissionLevel` types and the catalogue registry. `server/router/permissions.go` gains:

```go
func GetEffectivePermissions(user *User, organizationID string) map[Permission]PermissionLevel
func HasPermission(user *User, organizationID string, p Permission, min PermissionLevel) bool
```

Resolution is memoized per request and cached by user ID in the existing dual-mode cache, invalidated on any role, assignment or user mutation. `CanSpaceAdminOrg` and `CanAdminOrg` are removed once every call site is migrated; `CanAccessOrg` remains as an organization-membership test.

### Plugins

`SeatsurfingPlugin` gains `GetPermissionDefinitions() []PermissionDefinition`, called during plugin registration so that plugin permissions appear in the matrix editor. `AdminUIMenuItem` gains `RequiredPermission` and `RequiredLevel`, superseding `Visibility`. `HostAPI` gains `HasPermission(userID, organizationID, permission string, minLevel int) (bool, error)`, replacing `IsOrgAdmin` and `IsSuperAdmin`.

Because plugins are independently built binaries, both skew directions degrade gracefully: a plugin that declares no permissions falls back to the legacy `Visibility` mapping, and a plugin built against a newer interface tolerates the host not implementing the new calls.

### Lock-out prevention

1. The Organization Administrator role is `system` and cannot be edited or deleted.
2. `assertOrgRetainsAdmin(organizationID)` runs inside the transaction of every role, assignment, user-delete and user-disable mutation, requiring at least one enabled, non-service-account user with `roles:admin` and `users:admin`.
3. No user may grant a level exceeding their own effective level for that permission.
4. No user may remove their own last assignment granting `roles:admin`.
5. `assertEveryOrgHasAdmin()` runs at startup and repairs any organization left without an administrator by re-assigning the Organization Administrator role to its oldest enabled user, logging the repair.
6. `GET /user/{id}/permissions` lets the UI show the resolved effect of a change before it is saved.
