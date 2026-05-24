import React from "react";
import { Button, Dropdown } from "react-bootstrap";
import { TranslationFunc, withTranslation } from "@/components/withTranslation";

// Element types and their default sizes
const ELEMENT_DEFAULTS: Record<
  string,
  { width: number; height: number; thickness?: number }
> = {
  plant: { width: 32, height: 32 },
  window: { width: 60, height: 8 },
  door: { width: 40, height: 40 },
  toilet: { width: 40, height: 40 },
};

const SNAP_THRESHOLD = 10;
const PADDING = 40;
const MIN_CANVAS_W = 400;
const MIN_CANVAS_H = 300;
const DEFAULT_WALL_THICKNESS = 8;

type ElementType = "wall" | "plant" | "window" | "door" | "toilet";
type Mode =
  | "select"
  | "draw-wall"
  | `add-entity:${Exclude<ElementType, "wall">}`;

interface WallElement {
  id: string;
  type: "wall";
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  thickness: number;
}

interface EntityElement {
  id: string;
  type: Exclude<ElementType, "wall">;
  x: number;
  y: number;
  width: number;
  height: number;
  rotation: number;
}

type FloorPlanElementDef = WallElement | EntityElement;

interface FloorPlanDesign {
  version: number;
  elements: FloorPlanElementDef[];
}

interface Props {
  designData: string;
  onChange: (designData: string) => void;
  t: TranslationFunc;
}

interface State {
  elements: FloorPlanElementDef[];
  mode: Mode;
  selectedId: string | null;
  // draw-wall state
  wallStart: { x: number; y: number } | null;
  mousePos: { x: number; y: number } | null;
  snapTarget: { x: number; y: number } | null;
  // drag/resize state
  dragging: boolean;
  dragStartMouse: { x: number; y: number } | null;
  dragStartElement: {
    x: number;
    y: number;
    x1?: number;
    y1?: number;
    x2?: number;
    y2?: number;
  } | null;
  dragHandle: string | null; // 'body' | 'ep1' | 'ep2' | 'nw' | 'ne' | 'se' | 'sw'
  orthoConstrained: boolean;
}

function generateId(): string {
  return Math.random().toString(36).slice(2, 10);
}

function parseDesign(designData: string): FloorPlanElementDef[] {
  if (!designData) return [];
  try {
    const parsed: FloorPlanDesign = JSON.parse(designData);
    return parsed.elements || [];
  } catch {
    return [];
  }
}

function serializeDesign(elements: FloorPlanElementDef[]): string {
  const design: FloorPlanDesign = { version: 1, elements };
  return JSON.stringify(design);
}

function distance(x1: number, y1: number, x2: number, y2: number): number {
  return Math.sqrt((x1 - x2) ** 2 + (y1 - y2) ** 2);
}

function snapPoint(
  x: number,
  y: number,
  elements: FloorPlanElementDef[],
  excludeId: string | null,
  threshold = SNAP_THRESHOLD,
): { x: number; y: number; snapped: boolean } {
  for (const el of elements) {
    if (el.id === excludeId) continue;
    if (el.type === "wall") {
      if (distance(x, y, el.x1, el.y1) <= threshold) {
        return { x: el.x1, y: el.y1, snapped: true };
      }
      if (distance(x, y, el.x2, el.y2) <= threshold) {
        return { x: el.x2, y: el.y2, snapped: true };
      }
    }
  }
  return { x, y, snapped: false };
}

/**
 * When ortho is active, snap the free axis to the nearest wall endpoint
 * coordinate along that axis (1-D check). Returns the snapped value or null.
 */
function snapAxisValue(
  value: number,
  axis: "x" | "y",
  elements: FloorPlanElementDef[],
  excludeId: string | null,
  threshold = SNAP_THRESHOLD,
): number | null {
  for (const el of elements) {
    if (el.id === excludeId) continue;
    if (el.type === "wall") {
      const v1 = axis === "x" ? el.x1 : el.y1;
      const v2 = axis === "x" ? el.x2 : el.y2;
      if (Math.abs(value - v1) <= threshold) return v1;
      if (Math.abs(value - v2) <= threshold) return v2;
    }
  }
  return null;
}

const WALL_SNAP_THRESHOLD = 20;

function closestPointOnSegment(
  px: number,
  py: number,
  x1: number,
  y1: number,
  x2: number,
  y2: number,
): { x: number; y: number; dist: number } {
  const dx = x2 - x1;
  const dy = y2 - y1;
  const lenSq = dx * dx + dy * dy;
  if (lenSq === 0) {
    return { x: x1, y: y1, dist: distance(px, py, x1, y1) };
  }
  const t = Math.max(0, Math.min(1, ((px - x1) * dx + (py - y1) * dy) / lenSq));
  const nearX = x1 + t * dx;
  const nearY = y1 + t * dy;
  return { x: nearX, y: nearY, dist: distance(px, py, nearX, nearY) };
}

function snapEntityToWall(
  entityType: string,
  cx: number,
  cy: number,
  width: number,
  height: number,
  elements: FloorPlanElementDef[],
): { x: number; y: number; rotation: number } | null {
  if (entityType !== "door" && entityType !== "window") return null;
  let bestDist = WALL_SNAP_THRESHOLD;
  let bestResult: { x: number; y: number; rotation: number } | null = null;
  for (const el of elements) {
    if (el.type !== "wall") continue;
    const pt = closestPointOnSegment(cx, cy, el.x1, el.y1, el.x2, el.y2);
    if (pt.dist < bestDist) {
      bestDist = pt.dist;
      const wallAngleRad = Math.atan2(el.y2 - el.y1, el.x2 - el.x1);
      const wallAngleDeg = wallAngleRad * (180 / Math.PI);
      let entityX: number;
      let entityY: number;
      if (entityType === "door") {
        // The door panel is the top edge (y = el.y) of the bounding box.
        // After rotate(wallAngleDeg, cx, cy) the panel sits on the wall.
        // Panel center is (0, -height/2) relative to entity center; rotate it
        // to find where the entity center must be so the panel lands on pt.
        const panelCx = pt.x - (height / 2) * Math.sin(wallAngleRad);
        const panelCy = pt.y + (height / 2) * Math.cos(wallAngleRad);
        entityX = panelCx - width / 2;
        entityY = panelCy - height / 2;
      } else {
        // Window: center of bounding box on the wall.
        entityX = pt.x - width / 2;
        entityY = pt.y - height / 2;
      }
      bestResult = {
        x: entityX,
        y: entityY,
        rotation: wallAngleDeg,
      };
    }
  }
  return bestResult;
}

function computeViewBox(
  elements: FloorPlanElementDef[],
  extraPoints?: { x: number; y: number }[],
): {
  x: number;
  y: number;
  w: number;
  h: number;
} {
  const hasContent =
    elements.length > 0 || (extraPoints && extraPoints.length > 0);
  if (!hasContent) {
    return { x: 0, y: 0, w: MIN_CANVAS_W, h: MIN_CANVAS_H };
  }
  let minX = Infinity,
    minY = Infinity,
    maxX = -Infinity,
    maxY = -Infinity;
  for (const el of elements) {
    if (el.type === "wall") {
      minX = Math.min(minX, el.x1, el.x2);
      minY = Math.min(minY, el.y1, el.y2);
      maxX = Math.max(maxX, el.x1, el.x2);
      maxY = Math.max(maxY, el.y1, el.y2);
    } else {
      minX = Math.min(minX, el.x);
      minY = Math.min(minY, el.y);
      maxX = Math.max(maxX, el.x + el.width);
      maxY = Math.max(maxY, el.y + el.height);
    }
  }
  if (extraPoints) {
    for (const pt of extraPoints) {
      minX = Math.min(minX, pt.x);
      minY = Math.min(minY, pt.y);
      maxX = Math.max(maxX, pt.x);
      maxY = Math.max(maxY, pt.y);
    }
  }
  if (minX === Infinity) {
    return { x: 0, y: 0, w: MIN_CANVAS_W, h: MIN_CANVAS_H };
  }
  const x = minX - PADDING;
  const y = minY - PADDING;
  const w = Math.max(maxX - minX + 2 * PADDING, MIN_CANVAS_W);
  const h = Math.max(maxY - minY + 2 * PADDING, MIN_CANVAS_H);
  return { x, y, w, h };
}

class FloorPlanDesigner extends React.Component<Props, State> {
  svgRef = React.createRef<SVGSVGElement>();

  constructor(props: Props) {
    super(props);
    this.state = {
      elements: parseDesign(props.designData),
      mode: "select",
      selectedId: null,
      wallStart: null,
      mousePos: null,
      snapTarget: null,
      dragging: false,
      dragStartMouse: null,
      dragStartElement: null,
      dragHandle: null,
      orthoConstrained: false,
    };
  }

  getSVGCoords(e: React.MouseEvent<SVGSVGElement>): { x: number; y: number } {
    const svg = this.svgRef.current;
    if (!svg) return { x: 0, y: 0 };
    const pt = svg.createSVGPoint();
    pt.x = e.clientX;
    pt.y = e.clientY;
    const ctm = svg.getScreenCTM();
    if (!ctm) return { x: 0, y: 0 };
    const svgPt = pt.matrixTransform(ctm.inverse());
    return { x: svgPt.x, y: svgPt.y };
  }

  getSVGCoordsFromClient(clientX: number, clientY: number): { x: number; y: number } {
    const svg = this.svgRef.current;
    if (!svg) return { x: 0, y: 0 };
    const pt = svg.createSVGPoint();
    pt.x = clientX;
    pt.y = clientY;
    const ctm = svg.getScreenCTM();
    if (!ctm) return { x: 0, y: 0 };
    const svgPt = pt.matrixTransform(ctm.inverse());
    return { x: svgPt.x, y: svgPt.y };
  }

  notifyChange = (elements: FloorPlanElementDef[]) => {
    this.props.onChange(serializeDesign(elements));
  };

  /**
   * Returns the best position for a wall endpoint: endpoint-snap takes priority,
   * then ortho constraint (horizontal/vertical) when within 15° of either axis.
   */
  getConstrainedPoint(
    rawPos: { x: number; y: number },
    excludeId: string | null,
  ): { x: number; y: number; snapped: boolean } {
    // 1. Endpoint snap on raw position
    const snap = snapPoint(rawPos.x, rawPos.y, this.state.elements, excludeId);
    if (snap.snapped) return { x: snap.x, y: snap.y, snapped: true };
    // 2. Apply ortho constraint when a wallStart exists
    let constrained = { x: rawPos.x, y: rawPos.y };
    let yFixed = false; // horizontal draw — Y locked, X free
    let xFixed = false; // vertical draw   — X locked, Y free
    if (this.state.wallStart) {
      const dx = rawPos.x - this.state.wallStart.x;
      const dy = rawPos.y - this.state.wallStart.y;
      if (dx !== 0 || dy !== 0) {
        const angleDeg =
          Math.atan2(Math.abs(dy), Math.abs(dx)) * (180 / Math.PI);
        const ORTHO_THRESHOLD = 15;
        if (angleDeg < ORTHO_THRESHOLD) {
          constrained = { x: rawPos.x, y: this.state.wallStart.y };
          yFixed = true;
        } else if (angleDeg > 90 - ORTHO_THRESHOLD) {
          constrained = { x: this.state.wallStart.x, y: rawPos.y };
          xFixed = true;
        }
      }
    }
    // 3. Second 2-D endpoint snap on ortho-constrained position (closes rectangles)
    if (constrained.x !== rawPos.x || constrained.y !== rawPos.y) {
      const snap2 = snapPoint(
        constrained.x,
        constrained.y,
        this.state.elements,
        excludeId,
      );
      if (snap2.snapped) return { x: snap2.x, y: snap2.y, snapped: true };
    }
    // 4. Axis snap: snap the free axis to a nearby wall-endpoint coordinate.
    //    Enables aligning parallel walls (e.g. drawing the 3rd wall of a rectangle
    //    so its endpoint lines up with the first wall's endpoint).
    if (yFixed) {
      const snappedX = snapAxisValue(
        constrained.x,
        "x",
        this.state.elements,
        excludeId,
      );
      if (snappedX !== null) {
        return { x: snappedX, y: constrained.y, snapped: true };
      }
    } else if (xFixed) {
      const snappedY = snapAxisValue(
        constrained.y,
        "y",
        this.state.elements,
        excludeId,
      );
      if (snappedY !== null) {
        return { x: constrained.x, y: snappedY, snapped: true };
      }
    }
    return { x: constrained.x, y: constrained.y, snapped: false };
  }

  handleSVGClick = (e: React.MouseEvent<SVGSVGElement>) => {
    if (e.target === e.currentTarget) {
      // Clicked on empty canvas
      if (this.state.mode === "select") {
        this.setState({ selectedId: null });
      }
    }
  };

  handleSVGMouseMove = (e: React.MouseEvent<SVGSVGElement>) => {
    // Drag moves are handled by the window-level listener to support
    // the cursor leaving the SVG while dragging.
    if (this.state.dragging) return;

    const pos = this.getSVGCoords(e);

    // Update mouse position for wall preview / snap indicator
    if (this.state.mode === "draw-wall") {
      const constrained = this.getConstrainedPoint(pos, null);
      const isOrtho =
        this.state.wallStart !== null &&
        !constrained.snapped &&
        (constrained.x !== pos.x || constrained.y !== pos.y);
      this.setState({
        mousePos: { x: constrained.x, y: constrained.y },
        snapTarget: constrained.snapped
          ? { x: constrained.x, y: constrained.y }
          : null,
        orthoConstrained: isOrtho,
      });
    }
  };

  applyDragMove = (pos: { x: number; y: number }) => {
    if (
      !this.state.dragging ||
      !this.state.dragStartMouse ||
      !this.state.dragStartElement
    )
      return;
    const dx = pos.x - this.state.dragStartMouse.x;
    const dy = pos.y - this.state.dragStartMouse.y;
    const handle = this.state.dragHandle;
    const startEl = this.state.dragStartElement;
    const elements = this.state.elements.map((el) => {
      if (el.id !== this.state.selectedId) return el;
      if (el.type === "wall") {
        if (handle === "ep1") {
          const snapped = snapPoint(
            startEl.x1! + dx,
            startEl.y1! + dy,
            this.state.elements,
            el.id,
          );
          return { ...el, x1: snapped.x, y1: snapped.y };
        } else if (handle === "ep2") {
          const snapped = snapPoint(
            startEl.x2! + dx,
            startEl.y2! + dy,
            this.state.elements,
            el.id,
          );
          return { ...el, x2: snapped.x, y2: snapped.y };
        } else {
          // body drag
          return {
            ...el,
            x1: startEl.x1! + dx,
            y1: startEl.y1! + dy,
            x2: startEl.x2! + dx,
            y2: startEl.y2! + dy,
          };
        }
      } else {
        // Entity element
        if (handle === "rotate") {
          const cx = startEl.x1!;
          const cy = startEl.y1!;
          const startAngle = startEl.x!;
          const startRotation = startEl.y!;
          const currentAngle = Math.atan2(pos.y - cy, pos.x - cx);
          const deltaAngle = (currentAngle - startAngle) * (180 / Math.PI);
          let newRotation = startRotation + deltaAngle;
          const snapped = Math.round(newRotation / 90) * 90;
          if (Math.abs(newRotation - snapped) < 10) newRotation = snapped;
          return { ...el, rotation: newRotation };
        } else if (handle === "body") {
          const newX = startEl.x! + dx;
          const newY = startEl.y! + dy;
          const wallSnap = snapEntityToWall(
            el.type,
            newX + el.width / 2,
            newY + el.height / 2,
            el.width,
            el.height,
            this.state.elements,
          );
          if (wallSnap) {
            return {
              ...el,
              x: wallSnap.x,
              y: wallSnap.y,
              rotation: wallSnap.rotation,
            };
          }
          return { ...el, x: newX, y: newY };
        } else if (handle === "se") {
          return {
            ...el,
            width: Math.max(10, startEl.x2! + dx),
            height: Math.max(10, startEl.y2! + dy),
          };
        } else if (handle === "sw") {
          const newW = Math.max(10, startEl.x2! - dx);
          return {
            ...el,
            x: startEl.x! + (startEl.x2! - newW),
            width: newW,
            height: Math.max(10, startEl.y2! + dy),
          };
        } else if (handle === "ne") {
          const newH = Math.max(10, startEl.y2! - dy);
          return {
            ...el,
            y: startEl.y! + (startEl.y2! - newH),
            width: Math.max(10, startEl.x2! + dx),
            height: newH,
          };
        } else if (handle === "nw") {
          const newW = Math.max(10, startEl.x2! - dx);
          const newH = Math.max(10, startEl.y2! - dy);
          return {
            ...el,
            x: startEl.x! + (startEl.x2! - newW),
            y: startEl.y! + (startEl.y2! - newH),
            width: newW,
            height: newH,
          };
        }
      }
      return el;
    });
    this.setState({ elements });
  };

  finalizeDrag = () => {
    if (!this.state.dragging) return;
    const elements = [...this.state.elements];
    this.setState({
      dragging: false,
      dragStartMouse: null,
      dragStartElement: null,
      dragHandle: null,
    });
    this.notifyChange(elements);
    this.detachWindowDragListeners();
  };

  attachWindowDragListeners = () => {
    window.addEventListener("mousemove", this.handleWindowMouseMove);
    window.addEventListener("mouseup", this.handleWindowMouseUp);
  };

  detachWindowDragListeners = () => {
    window.removeEventListener("mousemove", this.handleWindowMouseMove);
    window.removeEventListener("mouseup", this.handleWindowMouseUp);
  };

  handleWindowMouseMove = (e: MouseEvent) => {
    const pos = this.getSVGCoordsFromClient(e.clientX, e.clientY);
    this.applyDragMove(pos);
  };

  handleWindowMouseUp = () => {
    this.finalizeDrag();
  };

  handleSVGMouseDown = (e: React.MouseEvent<SVGSVGElement>) => {
    if (this.state.mode !== "select") return;
    // Only start drag if clicking on an element handle
  };

  handleSVGMouseUp = (e: React.MouseEvent<SVGSVGElement>) => {
    this.finalizeDrag();
  };

  handleCanvasClick = (e: React.MouseEvent<SVGSVGElement>) => {
    const pos = this.getSVGCoords(e);

    if (this.state.mode === "draw-wall") {
      if (e.button === 2) return; // right click handled separately
      const constrained = this.getConstrainedPoint(pos, null);
      const pt = { x: constrained.x, y: constrained.y };
      if (!this.state.wallStart) {
        this.setState({ wallStart: pt });
      } else {
        // Complete the wall
        const newWall: WallElement = {
          id: generateId(),
          type: "wall",
          x1: this.state.wallStart.x,
          y1: this.state.wallStart.y,
          x2: pt.x,
          y2: pt.y,
          thickness: DEFAULT_WALL_THICKNESS,
        };
        const elements = [...this.state.elements, newWall];
        this.setState({
          elements,
          wallStart: pt,
          snapTarget: null,
          orthoConstrained: false,
        });
        this.notifyChange(elements);
      }
      return;
    }

    const modeStr = this.state.mode as string;
    if (modeStr.startsWith("add-entity:")) {
      const entityType = modeStr.replace("add-entity:", "") as Exclude<
        ElementType,
        "wall"
      >;
      const defaults = ELEMENT_DEFAULTS[entityType];
      const wallSnap = snapEntityToWall(
        entityType,
        pos.x,
        pos.y,
        defaults.width,
        defaults.height,
        this.state.elements,
      );
      const newEl: EntityElement = {
        id: generateId(),
        type: entityType,
        x: wallSnap ? wallSnap.x : pos.x - defaults.width / 2,
        y: wallSnap ? wallSnap.y : pos.y - defaults.height / 2,
        width: defaults.width,
        height: defaults.height,
        rotation: wallSnap ? wallSnap.rotation : 0,
      };
      const elements = [...this.state.elements, newEl];
      this.setState({ elements, mode: "select", selectedId: newEl.id });
      this.notifyChange(elements);
    }
  };

  handleSVGContextMenu = (e: React.MouseEvent<SVGSVGElement>) => {
    e.preventDefault();
    if (this.state.mode === "draw-wall") {
      this.setState({
        wallStart: null,
        mousePos: null,
        snapTarget: null,
        orthoConstrained: false,
      });
    }
  };

  handleKeyDown = (e: KeyboardEvent) => {
    const activeTag = (document.activeElement?.tagName ?? "").toLowerCase();
    const isEditing =
      activeTag === "input" ||
      activeTag === "textarea" ||
      activeTag === "select";
    if (e.key === "Escape") {
      if (this.state.mode === "draw-wall" && this.state.wallStart) {
        this.setState({
          wallStart: null,
          mousePos: null,
          snapTarget: null,
          orthoConstrained: false,
        });
      }
    }
    if (
      (e.key === "Delete" || e.key === "Backspace") &&
      !isEditing &&
      this.state.selectedId
    ) {
      this.deleteSelected();
    }
  };

  componentDidMount() {
    window.addEventListener("keydown", this.handleKeyDown);
  }

  componentDidUpdate(prevProps: Props) {
    if (
      prevProps.designData !== this.props.designData &&
      !this.state.dragging
    ) {
      this.setState({ elements: parseDesign(this.props.designData) });
    }
  }

  componentWillUnmount() {
    window.removeEventListener("keydown", this.handleKeyDown);
    this.detachWindowDragListeners();
  }

  deleteSelected = () => {
    if (!this.state.selectedId) return;
    const elements = this.state.elements.filter(
      (el) => el.id !== this.state.selectedId,
    );
    this.setState({ elements, selectedId: null });
    this.notifyChange(elements);
  };

  startDrag = (e: React.MouseEvent, elementId: string, handle: string) => {
    e.stopPropagation();
    const pos = this.getSVGCoords(e as React.MouseEvent<SVGSVGElement>);
    const el = this.state.elements.find((x) => x.id === elementId);
    if (!el) return;
    let startEl: State["dragStartElement"];
    if (el.type === "wall") {
      startEl = { x1: el.x1, y1: el.y1, x2: el.x2, y2: el.y2, x: 0, y: 0 };
    } else {
      startEl = { x: el.x, y: el.y, x2: el.width, y2: el.height };
    }
    this.setState({
      dragging: true,
      dragStartMouse: pos,
      dragStartElement: startEl,
      dragHandle: handle,
      selectedId: elementId,
    });
    this.attachWindowDragListeners();
  };

  startRotateHandle = (e: React.MouseEvent, elementId: string) => {
    e.stopPropagation();
    const pos = this.getSVGCoords(e as React.MouseEvent<SVGSVGElement>);
    const el = this.state.elements.find((x) => x.id === elementId);
    if (!el || el.type === "wall") return;
    const cx = el.x + el.width / 2;
    const cy = el.y + el.height / 2;
    const startAngle = Math.atan2(pos.y - cy, pos.x - cx);
    this.setState({
      dragging: true,
      dragStartMouse: pos,
      // Encode: x=startAngle(rad), y=startRotation(deg), x1/y1=entity center
      dragStartElement: {
        x: startAngle,
        y: el.rotation,
        x1: cx,
        y1: cy,
        x2: 0,
        y2: 0,
      },
      dragHandle: "rotate",
      selectedId: elementId,
    });
    this.attachWindowDragListeners();
  };

  renderWall = (el: WallElement) => {
    const isSelected = el.id === this.state.selectedId;
    return (
      <g key={el.id}>
        <line
          x1={el.x1}
          y1={el.y1}
          x2={el.x2}
          y2={el.y2}
          stroke="#555555"
          strokeWidth={el.thickness}
          strokeLinecap="round"
          style={{ cursor: "pointer" }}
          onMouseDown={(e) => {
            if (this.state.mode === "select") {
              this.setState({ selectedId: el.id });
              this.startDrag(e, el.id, "body");
            }
          }}
        />
        {isSelected && (
          <>
            {/* Endpoint 1 handle */}
            <circle
              cx={el.x1}
              cy={el.y1}
              r={6}
              fill="#2196f3"
              stroke="white"
              strokeWidth={2}
              style={{ cursor: "grab" }}
              onMouseDown={(e) => this.startDrag(e, el.id, "ep1")}
            />
            {/* Endpoint 2 handle */}
            <circle
              cx={el.x2}
              cy={el.y2}
              r={6}
              fill="#2196f3"
              stroke="white"
              strokeWidth={2}
              style={{ cursor: "grab" }}
              onMouseDown={(e) => this.startDrag(e, el.id, "ep2")}
            />
            {/* Selection outline */}
            <rect
              x={Math.min(el.x1, el.x2) - 4}
              y={Math.min(el.y1, el.y2) - 4}
              width={Math.abs(el.x2 - el.x1) + 8}
              height={Math.abs(el.y2 - el.y1) + 8}
              fill="none"
              stroke="#2196f3"
              strokeWidth={1}
              strokeDasharray="4,4"
            />
          </>
        )}
      </g>
    );
  };

  renderEntity = (el: EntityElement) => {
    const isSelected = el.id === this.state.selectedId;
    const cx = el.x + el.width / 2;
    const cy = el.y + el.height / 2;
    const transform =
      el.rotation !== 0 ? `rotate(${el.rotation} ${cx} ${cy})` : undefined;

    let shape: React.ReactNode;
    switch (el.type) {
      case "plant":
        shape = (
          <ellipse
            cx={cx}
            cy={cy}
            rx={el.width / 2}
            ry={el.height / 2}
            fill="#4caf50"
            stroke="#388e3c"
            strokeWidth={1}
          />
        );
        break;
      case "window":
        shape = (
          <rect
            x={el.x}
            y={el.y}
            width={el.width}
            height={el.height}
            fill="#90caf9"
            stroke="#1565c0"
            strokeWidth={1}
          />
        );
        break;
      case "door": {
        const path = `M ${el.x},${el.y} L ${el.x + el.width},${el.y} A ${el.width},${el.height} 0 0,1 ${el.x},${el.y + el.height} Z`;
        shape = (
          <path d={path} fill="#ffcc80" stroke="#e65100" strokeWidth={1} />
        );
        break;
      }
      case "toilet": {
        const tankH = el.height * 0.35;
        const seatCy = el.y + tankH + (el.height - tankH) * 0.5;
        shape = (
          <>
            <rect
              x={el.x}
              y={el.y}
              width={el.width}
              height={tankH}
              fill="#eeeeee"
              stroke="#9e9e9e"
              strokeWidth={1}
            />
            <ellipse
              cx={cx}
              cy={seatCy}
              rx={el.width * 0.4}
              ry={(el.height - tankH) * 0.45}
              fill="#eeeeee"
              stroke="#9e9e9e"
              strokeWidth={1}
            />
          </>
        );
        break;
      }
    }

    const handleSize = 7;
    const corners = [
      { key: "nw", x: el.x, y: el.y },
      { key: "ne", x: el.x + el.width, y: el.y },
      { key: "se", x: el.x + el.width, y: el.y + el.height },
      { key: "sw", x: el.x, y: el.y + el.height },
    ];

    return (
      <g
        key={el.id}
        transform={transform}
        style={{ cursor: "pointer" }}
        onMouseDown={(e) => {
          if (this.state.mode === "select") {
            this.setState({ selectedId: el.id });
            this.startDrag(e, el.id, "body");
          }
        }}
      >
        {shape}
        {isSelected && (
          <>
            {/* Rotation handle */}
            <line
              x1={cx}
              y1={el.y - 3}
              x2={cx}
              y2={el.y - 20}
              stroke="#2196f3"
              strokeWidth={1.5}
              pointerEvents="none"
            />
            <circle
              cx={cx}
              cy={el.y - 26}
              r={6}
              fill="white"
              stroke="#2196f3"
              strokeWidth={2}
              style={{ cursor: "crosshair" }}
              onMouseDown={(e) => this.startRotateHandle(e, el.id)}
            />
            <rect
              x={el.x - 3}
              y={el.y - 3}
              width={el.width + 6}
              height={el.height + 6}
              fill="none"
              stroke="#2196f3"
              strokeWidth={1.5}
              strokeDasharray="5,5"
            />
            {corners.map((corner) => (
              <rect
                key={corner.key}
                x={corner.x - handleSize / 2}
                y={corner.y - handleSize / 2}
                width={handleSize}
                height={handleSize}
                fill="#2196f3"
                stroke="white"
                strokeWidth={1}
                style={{ cursor: "nw-resize" }}
                onMouseDown={(e) => {
                  e.stopPropagation();
                  this.startDrag(e, el.id, corner.key);
                }}
              />
            ))}
          </>
        )}
      </g>
    );
  };

  render() {
    const { t } = this.props;
    const {
      elements,
      mode,
      wallStart,
      mousePos,
      snapTarget,
      selectedId,
      orthoConstrained,
    } = this.state;

    const extraPoints: { x: number; y: number }[] = [];
    if (mode === "draw-wall" && wallStart) extraPoints.push(wallStart);
    if (mode === "draw-wall" && mousePos) extraPoints.push(mousePos);
    const vb = computeViewBox(
      elements,
      extraPoints.length > 0 ? extraPoints : undefined,
    );
    const viewBox = `${vb.x} ${vb.y} ${vb.w} ${vb.h}`;
    const modeStr = mode as string;
    const isAddEntity = modeStr.startsWith("add-entity:");

    return (
      <div className="floor-plan-designer">
        {/* Toolbar */}
        <div className="fpd-toolbar mb-2">
          <div className="btn-group me-2">
            <Button
              size="sm"
              variant={mode === "select" ? "primary" : "outline-secondary"}
              onClick={() =>
                this.setState({
                  mode: "select",
                  wallStart: null,
                  mousePos: null,
                  snapTarget: null,
                })
              }
            >
              {t("selectMove")}
            </Button>
            <Button
              size="sm"
              variant={mode === "draw-wall" ? "primary" : "outline-secondary"}
              onClick={() =>
                this.setState({
                  mode: "draw-wall",
                  selectedId: null,
                })
              }
            >
              {t("drawWall")}
            </Button>
          </div>
          <Dropdown as="span" className="me-2">
            <Dropdown.Toggle
              size="sm"
              variant={isAddEntity ? "primary" : "outline-secondary"}
              id="fpd-add-entity"
            >
              {t("addEntity")}
            </Dropdown.Toggle>
            <Dropdown.Menu>
              {(
                ["plant", "window", "door", "toilet"] as Exclude<
                  ElementType,
                  "wall"
                >[]
              ).map((type) => (
                <Dropdown.Item
                  key={type}
                  onClick={() =>
                    this.setState({
                      mode: `add-entity:${type}`,
                      selectedId: null,
                      wallStart: null,
                    })
                  }
                >
                  {t(`entity${type.charAt(0).toUpperCase() + type.slice(1)}`)}
                </Dropdown.Item>
              ))}
            </Dropdown.Menu>
          </Dropdown>
          <Button
            size="sm"
            variant="outline-secondary"
            disabled={!selectedId}
            onClick={this.deleteSelected}
          >
            {t("deleteSelected")}
          </Button>
        </div>

        {/* Canvas */}
        <div className="mapScrollContainer">
          <svg
            ref={this.svgRef}
            viewBox={viewBox}
            width={vb.w}
            height={vb.h}
            style={{
              display: "block",
              cursor:
                mode === "draw-wall"
                  ? "crosshair"
                  : isAddEntity
                    ? "cell"
                    : "default",
            }}
            onClick={this.handleCanvasClick}
            onMouseMove={this.handleSVGMouseMove}
            onMouseUp={this.handleSVGMouseUp}
            onContextMenu={this.handleSVGContextMenu}
          >
            {/* Dot grid background */}
            <defs>
              <pattern
                id="fpd-grid"
                width="20"
                height="20"
                patternUnits="userSpaceOnUse"
                patternTransform={`translate(${-vb.x % 20} ${-vb.y % 20})`}
              >
                <circle cx="0" cy="0" r="1" fill="#cccccc" />
              </pattern>
            </defs>
            <rect
              x={vb.x}
              y={vb.y}
              width={vb.w}
              height={vb.h}
              fill="url(#fpd-grid)"
            />

            {/* Elements */}
            {elements.map((el) => {
              if (el.type === "wall") return this.renderWall(el as WallElement);
              return this.renderEntity(el as EntityElement);
            })}

            {/* Wall preview line */}
            {mode === "draw-wall" && wallStart && mousePos && (
              <line
                x1={wallStart.x}
                y1={wallStart.y}
                x2={mousePos.x}
                y2={mousePos.y}
                stroke={orthoConstrained ? "#4caf50" : "#2196f3"}
                strokeWidth={DEFAULT_WALL_THICKNESS}
                strokeLinecap="round"
                strokeDasharray="8,4"
                pointerEvents="none"
              />
            )}

            {/* Snap indicator */}
            {snapTarget && (
              <circle
                cx={snapTarget.x}
                cy={snapTarget.y}
                r={6}
                fill="#2196f3"
                pointerEvents="none"
              />
            )}

            {/* First click indicator */}
            {mode === "draw-wall" && wallStart && (
              <circle
                cx={wallStart.x}
                cy={wallStart.y}
                r={5}
                fill="#ff9800"
                pointerEvents="none"
              />
            )}
          </svg>
        </div>
      </div>
    );
  }
}

export default withTranslation(FloorPlanDesigner as any);
