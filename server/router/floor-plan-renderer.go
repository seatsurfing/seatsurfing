package router

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

const floorPlanPadding = 40.0
const floorPlanMinW = 400.0
const floorPlanMinH = 300.0

type floorPlanDesign struct {
	Version  int                `json:"version"`
	Elements []floorPlanElement `json:"elements"`
}

type floorPlanElement struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	X1        float64 `json:"x1"`
	Y1        float64 `json:"y1"`
	X2        float64 `json:"x2"`
	Y2        float64 `json:"y2"`
	Thickness float64 `json:"thickness"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Rotation  float64 `json:"rotation"`
}

type floorPlanBounds struct {
	minX, minY, maxX, maxY float64
}

func renderFloorPlanSVG(designData string) ([]byte, uint, uint, error) {
	var design floorPlanDesign
	if err := json.Unmarshal([]byte(designData), &design); err != nil {
		return nil, 0, 0, err
	}

	bounds := computeBounds(design.Elements)
	viewX := bounds.minX - floorPlanPadding
	viewY := bounds.minY - floorPlanPadding
	viewW := math.Max(bounds.maxX-bounds.minX+2*floorPlanPadding, floorPlanMinW)
	viewH := math.Max(bounds.maxY-bounds.minY+2*floorPlanPadding, floorPlanMinH)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="%s %s %s %s" width="%s" height="%s">`,
		fmtF(viewX), fmtF(viewY), fmtF(viewW), fmtF(viewH),
		fmtF(viewW), fmtF(viewH),
	))

	for i := range design.Elements {
		sb.WriteString(renderElement(&design.Elements[i]))
	}

	sb.WriteString(`</svg>`)
	return []byte(sb.String()), uint(math.Round(viewW)), uint(math.Round(viewH)), nil
}

func computeBounds(elements []floorPlanElement) floorPlanBounds {
	if len(elements) == 0 {
		return floorPlanBounds{0, 0, 400, 300}
	}
	b := floorPlanBounds{
		minX: math.MaxFloat64,
		minY: math.MaxFloat64,
		maxX: -math.MaxFloat64,
		maxY: -math.MaxFloat64,
	}
	for i := range elements {
		e := &elements[i]
		if e.Type == "wall" {
			b.expand(e.X1, e.Y1)
			b.expand(e.X2, e.Y2)
		} else {
			b.expand(e.X, e.Y)
			b.expand(e.X+e.Width, e.Y+e.Height)
		}
	}
	return b
}

func (b *floorPlanBounds) expand(x, y float64) {
	if x < b.minX {
		b.minX = x
	}
	if x > b.maxX {
		b.maxX = x
	}
	if y < b.minY {
		b.minY = y
	}
	if y > b.maxY {
		b.maxY = y
	}
}

func renderElement(e *floorPlanElement) string {
	switch e.Type {
	case "wall":
		return renderWall(e)
	case "plant":
		return renderPlant(e)
	case "window":
		return renderWindow(e)
	case "door":
		return renderDoor(e)
	case "toilet":
		return renderToilet(e)
	default:
		return ""
	}
}

func renderWall(e *floorPlanElement) string {
	thickness := e.Thickness
	if thickness <= 0 {
		thickness = 8
	}
	return fmt.Sprintf(
		`<line x1="%s" y1="%s" x2="%s" y2="%s" stroke="#555555" stroke-width="%s" stroke-linecap="round"/>`,
		fmtF(e.X1), fmtF(e.Y1), fmtF(e.X2), fmtF(e.Y2), fmtF(thickness),
	)
}

func renderPlant(e *floorPlanElement) string {
	cx := e.X + e.Width/2
	cy := e.Y + e.Height/2
	rx := e.Width / 2
	ry := e.Height / 2
	transform := rotateTransform(e.Rotation, cx, cy)
	return fmt.Sprintf(
		`<ellipse cx="%s" cy="%s" rx="%s" ry="%s" fill="#4caf50" stroke="#388e3c" stroke-width="1"%s/>`,
		fmtF(cx), fmtF(cy), fmtF(rx), fmtF(ry), transform,
	)
}

func renderWindow(e *floorPlanElement) string {
	cx := e.X + e.Width/2
	cy := e.Y + e.Height/2
	transform := rotateTransform(e.Rotation, cx, cy)
	return fmt.Sprintf(
		`<rect x="%s" y="%s" width="%s" height="%s" fill="#90caf9" stroke="#1565c0" stroke-width="1"%s/>`,
		fmtF(e.X), fmtF(e.Y), fmtF(e.Width), fmtF(e.Height), transform,
	)
}

func renderDoor(e *floorPlanElement) string {
	// Quarter-circle pie slice representing door swing
	// Hinge at top-left (e.X, e.Y), door panel goes right, arc sweeps downward
	x := e.X
	y := e.Y
	w := e.Width
	h := e.Height
	cx := x + w/2
	cy := y + h/2
	transform := rotateTransform(e.Rotation, cx, cy)
	// Path: M hinge, L door panel end, A arc back to bottom of hinge, Z
	path := fmt.Sprintf("M %s,%s L %s,%s A %s,%s 0 0,1 %s,%s Z",
		fmtF(x), fmtF(y),
		fmtF(x+w), fmtF(y),
		fmtF(w), fmtF(h),
		fmtF(x), fmtF(y+h),
	)
	return fmt.Sprintf(
		`<path d="%s" fill="#ffcc80" stroke="#e65100" stroke-width="1"%s/>`,
		path, transform,
	)
}

func renderToilet(e *floorPlanElement) string {
	x := e.X
	y := e.Y
	w := e.Width
	h := e.Height
	cx := x + w/2
	cy := y + h/2
	transform := rotateTransform(e.Rotation, cx, cy)
	// Tank: upper portion
	tankH := h * 0.35
	// Seat (bowl): ellipse in lower portion
	seatCy := y + tankH + (h-tankH)*0.5
	seatRx := w * 0.4
	seatRy := (h - tankH) * 0.45
	return fmt.Sprintf(
		`<rect x="%s" y="%s" width="%s" height="%s" fill="#eeeeee" stroke="#9e9e9e" stroke-width="1"%s/>`+
			`<ellipse cx="%s" cy="%s" rx="%s" ry="%s" fill="#eeeeee" stroke="#9e9e9e" stroke-width="1"%s/>`,
		fmtF(x), fmtF(y), fmtF(w), fmtF(tankH), transform,
		fmtF(cx), fmtF(seatCy), fmtF(seatRx), fmtF(seatRy), transform,
	)
}

func rotateTransform(rotation, cx, cy float64) string {
	if rotation == 0 {
		return ""
	}
	return fmt.Sprintf(` transform="rotate(%s %s %s)"`, fmtF(rotation), fmtF(cx), fmtF(cy))
}

func fmtF(f float64) string {
	if f == math.Trunc(f) {
		return fmt.Sprintf("%d", int(f))
	}
	return fmt.Sprintf("%.4g", f)
}
