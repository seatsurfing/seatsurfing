import React from "react";

interface Props {
  width: number;
  height: number;
}

const STROKE_WIDTH = 3;
const CORNER_RADIUS = 6 - STROKE_WIDTH / 2;

type Point = [number, number];

const insetPolygon = (points: Point[], inset: number): Point[] => {
  const n = points.length;
  const lines = points.map((p, i) => {
    const q = points[(i + 1) % n];
    const dx = q[0] - p[0];
    const dy = q[1] - p[1];
    const len = Math.hypot(dx, dy);
    const nx = -dy / len;
    const ny = dx / len;
    return {
      p: [p[0] + nx * inset, p[1] + ny * inset] as Point,
      d: [dx, dy] as Point,
    };
  });
  return lines.map((l1, i) => {
    const l0 = lines[(i - 1 + n) % n];
    const cross = l0.d[0] * l1.d[1] - l0.d[1] * l1.d[0];
    const t =
      ((l1.p[0] - l0.p[0]) * l1.d[1] - (l1.p[1] - l0.p[1]) * l1.d[0]) / cross;
    return [l0.p[0] + l0.d[0] * t, l0.p[1] + l0.d[1] * t];
  });
};

const roundedPath = (points: Point[], radius: number): string => {
  const n = points.length;
  let path = "";
  points.forEach((p, i) => {
    const prev = points[(i - 1 + n) % n];
    const next = points[(i + 1) % n];
    const v1: Point = [prev[0] - p[0], prev[1] - p[1]];
    const v2: Point = [next[0] - p[0], next[1] - p[1]];
    const l1 = Math.hypot(v1[0], v1[1]);
    const l2 = Math.hypot(v2[0], v2[1]);
    const angle = Math.acos(
      Math.max(-1, Math.min(1, (v1[0] * v2[0] + v1[1] * v2[1]) / (l1 * l2))),
    );
    const tangent = Math.min(radius / Math.tan(angle / 2), l1 / 2, l2 / 2);
    const r = tangent * Math.tan(angle / 2);
    const a: Point = [
      p[0] + (v1[0] / l1) * tangent,
      p[1] + (v1[1] / l1) * tangent,
    ];
    const b: Point = [
      p[0] + (v2[0] / l2) * tangent,
      p[1] + (v2[1] / l2) * tangent,
    ];
    path += `${i === 0 ? "M" : "L"}${a[0]},${a[1]} A${r},${r} 0 0 1 ${b[0]},${b[1]} `;
  });
  return path + "Z";
};

const SpaceTrapezoidOutline: React.FC<Props> = ({ width, height }) => {
  const points = insetPolygon(
    [
      [width * 0.2, 0],
      [width * 0.8, 0],
      [width, height],
      [0, height],
    ],
    STROKE_WIDTH / 2,
  );
  return (
    <svg
      className="space-trapezoid-outline"
      width="100%"
      height="100%"
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none"
      aria-hidden="true"
    >
      <path d={roundedPath(points, CORNER_RADIUS)} strokeWidth={STROKE_WIDTH} />
    </svg>
  );
};

export default SpaceTrapezoidOutline;
