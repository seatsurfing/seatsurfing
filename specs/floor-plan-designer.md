# Floor Plan Designer

## Overview

Allow admins to design simple floor plans directly within the admin web UI, without needing an external image file. The designed floor plan is stored in a structured JSON format and rendered on demand as SVG by the backend. The resulting SVG is served through the existing `/location/{id}/map` endpoint, making the feature transparent to the booking UI and all other existing consumers of floor plan data.

## Goals

- Draw, resize, move, and delete wall segments on an interactive canvas.
- Wall endpoints snap to nearby wall endpoints when within a configurable threshold.
- Place, move, resize, and delete static orientation entities: plants, windows, doors, and toilets.
- The canvas auto-sizes to the bounding box of all drawn elements (no fixed canvas size upfront).
- Designed floor plans are stored in editable structured JSON and can be modified after saving.
- The backend renders design JSON to SVG and returns it via the existing `GET /location/{id}/map` endpoint, so no other parts of the application (booking UI, space placement editor) need changes.
- The location edit form lets admins choose between "Upload file" (existing) and "Design floor plan" (new) modes.

## Non-Goals

- Multi-story or layered floor plans.
- Freehand/curved wall drawing.
- Undo/redo history within the designer.
- Real-time collaboration or multi-user editing.
- SVG or DXF import (use the existing upload path for externally created files).
- Printing or PDF export beyond what the generated SVG already provides.

## Product Decisions

- The designer canvas is SVG-based (not `<canvas>`), consistent with the existing `SpaceRect` component pattern.
- Two-click wall placement (click first endpoint, click second endpoint) rather than click-and-drag, to make precise placement easier.
- Snap threshold is 10 logical pixels.
- The backend generates SVG from design JSON using inline Go string generation — no additional Go SVG library is introduced.
- Switching between "upload" and "designer" modes is lossless until the form is saved: switching away from a mode does not immediately delete that mode's data.

## Existing System Constraints

### Floor Plan Storage (Backend)

The `locations` table stores the uploaded floor plan as binary data:

```sql
CREATE TABLE locations (
  id            uuid DEFAULT uuid_generate_v4(),
  organization_id uuid NOT NULL,
  name          VARCHAR NOT NULL,
  map_mimetype  VARCHAR DEFAULT '',   -- 'png', 'jpeg', 'gif', 'svg+xml', or ''
  map_data      BYTEA,               -- raw image bytes
  map_width     INTEGER DEFAULT 0,
  map_height    INTEGER DEFAULT 0,
  map_scale     real NOT NULL DEFAULT 1.0,
  description   VARCHAR DEFAULT '',
  max_concurrent_bookings INTEGER DEFAULT 0,
  tz            VARCHAR DEFAULT '',
  enabled       boolean NOT NULL DEFAULT TRUE,
  PRIMARY KEY (id)
);
```

`map_mimetype` is empty when no map has been uploaded. Absence of a map is the default state for new locations.

### Map API Endpoints

- `POST /location/{id}/map` — accepts a raw binary body (PNG, JPEG, GIF, or SVG), decodes dimensions, stores in `map_data`. Implemented in `router/location-router.go` → `setMap`.
- `GET /location/{id}/map` — returns `{ width, height, mimeType, scale, data: base64 }`. Implemented in `router/location-router.go` → `getMap`.

### Location Update API

`PUT /location/{id}` accepts a `CreateLocationRequest` body (`router/location-router.go`):

```go
type CreateLocationRequest struct {
    Name                  string   `json:"name" validate:"required,max=128"`
    Description           string   `json:"description" validate:"max=512"`
    MaxConcurrentBookings uint     `json:"maxConcurrentBookings"`
    Timezone              string   `json:"timezone" validate:"max=32"`
    Enabled               bool     `json:"enabled"`
    MapScale              float64  `json:"mapScale"`
    AllowedBookerGroupIDs []string `json:"allowedBookerGroupIds" validate:"dive,uuid"`
}
```

### Frontend Location Type

`ui/src/types/Location.ts` exposes the `Location` entity class. Relevant fields: `mapWidth`, `mapHeight`, `mapMimeType`, `mapScale`. The `setMap(file: File)` method posts to `POST /location/{id}/map`.

### Location Edit Form

`ui/src/pages/admin/locations/[id].tsx` contains `EditLocation`, a class component. The floor plan upload is done in `saveSpaces` (after the main entity save). `this.mapData` caches the loaded map data used as a CSS `background-image` for the space-placement overlay.

### Current DB Schema Version

The current `targetVersion` in `server/repository/db-updates.go` is **43**. All new migrations must use version **44+**.

### Authorization Helpers

- `CanAccessOrg(user, orgID)` — any authenticated member of the org.
- `CanSpaceAdminOrg(user, orgID)` — space admin or org admin.

---

## Data Model Changes

### New column: `locations.map_type`

Distinguishes how the floor plan was created:

| Value        | Meaning                                         |
|--------------|-------------------------------------------------|
| `''` (empty) | Uploaded image file — existing behavior, backward compatible |
| `'designed'` | Created by the floor plan designer              |

```sql
-- Migration at version 44
ALTER TABLE locations ADD COLUMN IF NOT EXISTS map_type VARCHAR NOT NULL DEFAULT '';
```

### New table: `location_floor_plans`

Stores the designer's structured data as JSON text, one row per location:

```sql
-- Migration at version 44
CREATE TABLE IF NOT EXISTS location_floor_plans (
    location_id     uuid PRIMARY KEY REFERENCES locations(id) ON DELETE CASCADE,
    organization_id uuid NOT NULL,
    design_data     TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_location_floor_plans_org ON location_floor_plans(organization_id);
```

### Design JSON Schema

`design_data` is a JSON string with the following structure:

```json
{
  "version": 1,
  "elements": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "type": "wall",
      "x1": 0,
      "y1": 0,
      "x2": 300,
      "y2": 0,
      "thickness": 8
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "type": "plant",
      "x": 50,
      "y": 50,
      "width": 32,
      "height": 32,
      "rotation": 0
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "type": "window",
      "x": 100,
      "y": 0,
      "width": 60,
      "height": 8,
      "rotation": 0
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "type": "door",
      "x": 200,
      "y": 0,
      "width": 40,
      "height": 40,
      "rotation": 0
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440004",
      "type": "toilet",
      "x": 250,
      "y": 50,
      "width": 40,
      "height": 40,
      "rotation": 0
    }
  ]
}
```

**Element types and default sizes (logical pixels):**

| Type      | Geometry fields                              | Default size  | Notes                                           |
|-----------|----------------------------------------------|---------------|-------------------------------------------------|
| `wall`    | `x1, y1, x2, y2, thickness`                  | thickness = 8 | Defined by two endpoints, not a bounding box    |
| `plant`   | `x, y, width, height, rotation`              | 32 × 32       | Center is `(x + width/2, y + height/2)`         |
| `window`  | `x, y, width, height, rotation`              | 60 × 8        |                                                 |
| `door`    | `x, y, width, height, rotation`              | 40 × 40       | Quarter-circle arc indicates swing direction    |
| `toilet`  | `x, y, width, height, rotation`              | 40 × 40       |                                                 |

All coordinate values are numbers (integers or floats). `rotation` is degrees clockwise.

---

## API Changes

### New repository file: `server/repository/location-floor-plan-repository.go`

Follows the singleton pattern. Key methods:

```go
type LocationFloorPlan struct {
    LocationID     string
    OrganizationID string
    DesignData     string // JSON text
}

func GetLocationFloorPlanRepository() *LocationFloorPlanRepository
func (r *LocationFloorPlanRepository) GetDesign(locationID string) (*LocationFloorPlan, error)
func (r *LocationFloorPlanRepository) SetDesign(e *LocationFloorPlan) error // upsert
func (r *LocationFloorPlanRepository) Delete(locationID string) error
func (r *LocationFloorPlanRepository) RunSchemaUpgrade(curVersion, targetVersion int)
```

`SetDesign` uses `INSERT ... ON CONFLICT (location_id) DO UPDATE SET design_data = EXCLUDED.design_data, organization_id = EXCLUDED.organization_id`.

### Modified: `GET /location/{id}/map`

When `locations.map_type = 'designed'`:

1. Call `GetLocationFloorPlanRepository().GetDesign(locationID)` to retrieve the JSON.
2. Parse all element coordinates to compute the bounding box:
   - For `wall` elements: include both `(x1, y1)` and `(x2, y2)`.
   - For other elements: include `(x, y)` and `(x + width, y + height)`.
3. Add 40 px padding on all sides. The resulting rectangle is the SVG `viewBox` and the `width`/`height` of the response.
4. Render each element to SVG markup using Go string generation (no external library):
   - `wall` → `<rect>` centered on the line, rotated appropriately, or `<line>` with `stroke-width` equal to `thickness`.
   - `plant` → `<ellipse>` in green (`#4caf50`).
   - `window` → `<rect>` in light blue (`#90caf9`) with a thin stroke.
   - `door` → `<path>` describing a quarter-circle arc.
   - `toilet` → `<rect>` with an inner arc for the seat.
   - All transforms apply `rotation` around the element's center.
5. Wrap elements in `<svg xmlns="http://www.w3.org/2000/svg" viewBox="..." width="..." height="...">`.
6. Return the SVG bytes as the response with `mimeType: "svg+xml"` and the computed `width`, `height`, and `scale: 1.0`.

When `map_type` is empty, behavior is unchanged (existing code path).

### New: `GET /location/{id}/floorplan-design`

Returns the raw design JSON for the editor to load.

- **Auth**: `CanAccessOrg`
- **Response 200**:
  ```json
  { "designData": "{ \"version\": 1, \"elements\": [] }" }
  ```
- **Response 404**: If no design record exists (new location), return `{ "designData": "" }` with status 200 (empty design).

### New: `POST /location/{id}/floorplan-design`

Saves the design JSON.

- **Auth**: `CanSpaceAdminOrg`
- **Request body**:
  ```json
  { "designData": "{ \"version\": 1, \"elements\": [...] }" }
  ```
  `designData` must be valid JSON (server validates by attempting `json.Valid`).
- **Response 204**: Success.
- **Response 400**: `designData` is not valid JSON.

### Modified: `PUT /location/{id}` and `POST /location/`

Add optional `mapType` field to `CreateLocationRequest`:

```go
type CreateLocationRequest struct {
    // ... existing fields ...
    MapType string `json:"mapType" validate:"omitempty,oneof='' designed"`
}
```

The `update` and `create` handlers write `MapType` to `locations.map_type`. When `mapType` transitions from `"designed"` to `""`, the handler also clears `map_data`, `map_width`, `map_height`, and `map_mimetype` to avoid serving stale data. When transitioning from `""` to `"designed"`, the existing `map_data` is left untouched (it will simply not be served while `map_type = 'designed'`).

### Route registration

New routes added in `SetupRoutes` for `LocationRouter`:

```go
s.HandleFunc("/{id}/floorplan-design", router.getFloorPlanDesign).Methods("GET")
s.HandleFunc("/{id}/floorplan-design", router.setFloorPlanDesign).Methods("POST")
```

### Register new repository in `db-updates.go`

Add `GetLocationFloorPlanRepository()` to the `repositories` slice. Add a `targetVersion = 44` block for the two DDL statements (new column + new table).

---

## UI Changes

### `ui/src/types/Location.ts`

Add field:

```ts
mapType: string; // '' or 'designed'
```

Update `serialize()` to include `mapType`. Update `deserialize()` to read `mapType`. Add methods:

```ts
async getFloorPlanDesign(): Promise<string> {
    return Ajax.get(this.getBackendUrl() + this.id + "/floorplan-design").then(
        (result) => result.json.designData as string
    );
}

async setFloorPlanDesign(designData: string): Promise<void> {
    return Ajax.postData(this.getBackendUrl() + this.id + "/floorplan-design", { designData }).then(
        () => undefined
    );
}
```

### `ui/src/pages/admin/locations/[id].tsx`

#### New state fields

```ts
interface State {
    // ... existing fields ...
    mapType: 'upload' | 'designed';
    designData: string;
}
```

Initial values: `mapType: 'upload'`, `designData: ''`.

#### `loadData`

After loading the location entity, if `entity.mapType === 'designed'`, call `entity.getFloorPlanDesign()` and store the result in `this.state.designData`. Set `this.state.mapType` from `entity.mapType` (map `'designed'` → `'designed'`, anything else → `'upload'`).

#### Floor plan source selection (render method)

Replace the existing single `<Form.Control type="file" ...>` field with:

1. Two radio buttons for the source mode:
   ```tsx
   <Form.Check
     type="radio" id="map-type-upload" name="mapType"
     label={this.props.t("uploadFile")}
     checked={this.state.mapType === 'upload'}
     onChange={() => this.setState({ mapType: 'upload' })}
   />
   <Form.Check
     type="radio" id="map-type-designed" name="mapType"
     label={this.props.t("designFloorPlan")}
     checked={this.state.mapType === 'designed'}
     onChange={() => this.setState({ mapType: 'designed' })}
   />
   ```

2. Conditionally render the existing file input (`mapType === 'upload'`) or the `<FloorPlanDesigner>` component (`mapType === 'designed'`).

3. The `required={!this.entity.id}` attribute on the file input is only applied when `mapType === 'upload'`.

#### `onSubmit` / save flow

After `this.entity.save()` resolves:

- If `mapType === 'upload'` and `this.state.files` is set: upload the file via `this.entity.setMap(file)` as before.
- If `mapType === 'designed'`: call `this.entity.setFloorPlanDesign(this.state.designData)`.
- Include `mapType` in the entity serialization so `PUT /location/{id}` receives it.

#### The designer in the existing floor plan section

The `floorPlan` section (rendered below the form when `this.entity.id` exists) continues to show the interactive space-placement overlay using the background image from `this.mapData`. When `mapType === 'designed'`, `loadData` still loads the map via `entity.getMap()` and stores it in `this.mapData`, so the space-placement overlay continues to work without changes (the backend serves the SVG through the same endpoint).

### New component: `ui/src/components/FloorPlanDesigner.tsx`

```ts
interface Props {
    designData: string;      // JSON string; '' for a new/empty canvas
    onChange: (designData: string) => void;
}
```

The component is a class component consistent with the rest of the codebase. It maintains its own `elements` array in state (parsed from `designData` on mount).

On every mutation it serializes the elements array back to a JSON string and calls `this.props.onChange(json)`.

#### Toolbar (above canvas)

| Control                           | Action                                              |
|-----------------------------------|-----------------------------------------------------|
| "Select / Move" button            | Sets active mode to `select`                        |
| "Draw Wall" button                | Sets active mode to `draw-wall`                     |
| "Add entity" dropdown             | Items: Plant, Window, Door, Toilet — sets mode to `add-entity:<type>` |
| "Delete selected" button          | Removes the selected element; disabled when nothing selected |

#### Canvas

The canvas is an `<svg>` element inside a `<div className="mapScrollContainer">`. The SVG has no fixed size; its `viewBox` and `width`/`height` are computed from the bounding box of all elements plus 40 px padding. When the canvas is empty, a minimum size of 400 × 300 is used.

A light gray dot grid (SVG `<pattern>`) provides orientation and is not part of the exported design.

#### Mode: `select`

- Click on an element: selects it (highlighted with a dashed blue border / handles).
- Drag a selected wall endpoint: moves that endpoint; snap applies.
- Drag a selected entity body: moves it.
- Corner handles on selected entities: resize.
- Click on empty canvas area: deselects.
- `Delete` key (or the toolbar button): removes the selected element.

#### Mode: `draw-wall`

- First click: records `(x1, y1)`. If click is within snap threshold of an existing wall endpoint, snap to it.
- Mouse move: renders a preview line from `(x1, y1)` to the cursor. Preview endpoint snaps visually.
- Second click: records `(x2, y2)` with snap. Adds wall element. Resets to first-click state (allowing chained wall drawing).
- Right-click or `Escape`: cancels the in-progress wall, returns to idle within `draw-wall` mode.

#### Mode: `add-entity:<type>`

- Click on canvas: places the entity centered at the click position with the default size for its type. Returns to `select` mode immediately after placement.

#### Snap behavior

```
function snapPoint(x, y, elements, excludeElementId, threshold = 10):
  for each wall element (excluding excludeElementId):
    if distance((x,y), (element.x1, element.y1)) <= threshold:
      return (element.x1, element.y1)
    if distance((x,y), (element.x2, element.y2)) <= threshold:
      return (element.x2, element.y2)
  return (x, y)  // no snap
```

A snap target is highlighted with a blue filled circle (radius 6 px) rendered above all other elements.

#### SVG element rendering

| Element  | SVG representation                                                                 |
|----------|------------------------------------------------------------------------------------|
| `wall`   | `<line x1 y1 x2 y2 stroke="#555" strokeWidth={thickness} strokeLinecap="round"/>` |
| `plant`  | `<ellipse>` in `#4caf50` with a slightly darker stroke                             |
| `window` | `<rect>` in `#90caf9`, thin stroke `#1565c0`                                      |
| `door`   | `<path>` — a right-angle triangle plus quarter-circle arc in `#ffcc80`             |
| `toilet` | `<rect>` (tank) plus `<ellipse>` (seat) in `#eeeeee` with a gray stroke           |

Selected element: wrapped in an `<g>` with a dashed blue `<rect>` outline and resize handle circles at corners.

#### New CSS file: `ui/src/styles/FloorPlanDesigner.css`

Contains styles for the toolbar, canvas container, snap indicator, and selection handles.

### New i18n keys

Add to `i18n/translations.en-GB.json` (then run `./add-missing-translations.sh`):

| Key                    | English value                                |
|------------------------|----------------------------------------------|
| `uploadFile`           | `Upload file`                                |
| `designFloorPlan`      | `Design floor plan`                          |
| `drawWall`             | `Draw wall`                                  |
| `addEntity`            | `Add entity`                                 |
| `selectMove`           | `Select / Move`                              |
| `deleteSelected`       | `Delete selected`                            |
| `entityPlant`          | `Plant`                                      |
| `entityWindow`         | `Window`                                     |
| `entityDoor`           | `Door`                                       |
| `entityToilet`         | `Toilet`                                     |

---

## Integration Points

### Booking UI and space search

Both consume `GET /location/{id}/map`. Because the backend serves the designer's SVG through the same endpoint, no changes are required in the booking UI or search flow.

### Space placement editor

The space-placement overlay in `[id].tsx` uses `this.mapData` as a CSS `background-image` (base64 data URL). Since `getMap()` returns the SVG data via the same API, this continues to work without modification.

### Plugin hooks

If a plugin implements `OnLocationUpdated`, it should fire on the `PUT /location/{id}` call as it does today. Saving the floor plan design via `POST /location/{id}/floorplan-design` does not need to trigger a plugin hook.

---

## Security Considerations

- The `designData` JSON is stored as opaque text and never executed server-side. During SVG generation, all values extracted from the JSON (coordinates, dimensions) are parsed as numbers and formatted with `strconv.FormatFloat`/`strconv.Itoa` — no user-supplied strings are interpolated into SVG attribute values.
- The generated SVG must not include `<script>`, `<foreignObject>`, event handler attributes, or any other XSS vectors. The Go SVG renderer only emits a fixed set of known SVG elements with numeric attributes.
- `POST /location/{id}/floorplan-design` validates that `designData` is valid JSON (`json.Valid`) before storing it, and requires `CanSpaceAdminOrg` authorization.
- `GET /location/{id}/floorplan-design` requires `CanAccessOrg` to prevent leaking floor plan data across organizations.
- All queries on `location_floor_plans` filter by `organization_id` to enforce multi-tenancy.
