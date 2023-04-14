package main

import "fmt"

type DataPoint struct {
	U     float64 // Longitude or East
	V     float64 // Latitude or North
	W     float64 // Vertical
	Value float32
}

func (p *DataPoint) ToRow() string {
	return fmt.Sprintf("%v,%v,%v\n", p.V, p.U, p.Value)
}
