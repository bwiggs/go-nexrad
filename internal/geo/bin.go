package geo

import (
	"fmt"
	"strings"

	"github.com/twpayne/go-proj/v10"
)

const coordFmt = "[%.4f,%.4f]"

type Poly []proj.Coord

type Bin struct {
	Coords Poly
	Value  float32
}

func NewBin(a proj.Coord, b proj.Coord, c proj.Coord, d proj.Coord, value float32) *Bin {
	return &Bin{
		Coords: []proj.Coord{a, b, c, d},
		Value:  value,
	}
}

func (b *Bin) AppendFeature(builder *strings.Builder) {
	fmt.Fprint(builder, "{\"type\":\"Feature\",\"geometry\":{\"type\":\"Polygon\",\"coordinates\":[[")

	// A, B, D, C, A
	fmt.Fprintf(builder, coordFmt, b.Coords[0].X(), b.Coords[0].Y())
	fmt.Fprint(builder, ",")
	fmt.Fprintf(builder, coordFmt, b.Coords[1].X(), b.Coords[1].Y())
	fmt.Fprint(builder, ",")
	fmt.Fprintf(builder, coordFmt, b.Coords[3].X(), b.Coords[3].Y())
	fmt.Fprint(builder, ",")
	fmt.Fprintf(builder, coordFmt, b.Coords[2].X(), b.Coords[2].Y())
	fmt.Fprint(builder, ",")
	fmt.Fprintf(builder, coordFmt, b.Coords[0].X(), b.Coords[0].Y())
	fmt.Fprint(builder, "]]},\"properties\":{\"value\":")
	fmt.Fprintf(builder, "%.1f", b.Value)
	fmt.Fprint(builder, "}}")
}
