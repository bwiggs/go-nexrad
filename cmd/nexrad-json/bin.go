package main

import (
	geojson "github.com/paulmach/go.geojson"
	"github.com/twpayne/go-proj/v10"
)

type Bin struct {
	A     proj.Coord
	B     proj.Coord
	C     proj.Coord
	D     proj.Coord
	Value float32
}

func (bin *Bin) ToPoly() *geojson.Feature {
	p := geojson.NewPolygonFeature(
		[][][]float64{
			{
				[]float64{bin.A.X(), bin.A.Y()},
				[]float64{bin.B.X(), bin.B.Y()},
				[]float64{bin.D.X(), bin.D.Y()},
				[]float64{bin.C.X(), bin.C.Y()},
				[]float64{bin.A.X(), bin.A.Y()},
			},
		},
	)

	p.SetProperty("value", bin.Value)

	return p
}

func (bin *Bin) Forward(pj *proj.PJ) (*Bin, error) {
	a, err := pj.Forward(bin.A)

	if err != nil {
		return nil, err
	}

	b, err := pj.Forward(bin.B)

	if err != nil {
		return nil, err
	}

	c, err := pj.Forward(bin.C)

	if err != nil {
		return nil, err
	}

	d, err := pj.Forward(bin.D)

	if err != nil {
		return nil, err
	}

	return &Bin{
		A:     a,
		B:     b,
		C:     c,
		D:     d,
		Value: bin.Value,
	}, nil
}
