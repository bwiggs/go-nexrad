package archive2

import (
	"encoding/binary"
	"io"
)

// Message15 Clutter Filter Map
// see documentation RDA/RPG 3-61
type Message15 struct {
	Message15Header
	// ElevCuts contains info for each elevation angle
	ElevSegments []Message15ElevSegment
}

type Message15Header struct {
	MapGenDate      uint16
	MapGenTime      uint16
	NumElevSegments uint16
}

type Message15ElevSegment struct {
	AzimuthSegments []Message15AzimuthSegment
}

type Message15AzimuthSegment struct {
	NumRangeZones uint16
	RangeZones    []Message15RangeZones
}

type Message15RangeZones struct {
	OpCode   uint16
	EndRange uint16
}

func (m15 *Message15) Read(r io.ReadSeeker) {
	binary.Read(r, binary.BigEndian, &m15.Message15Header)
	m15.ElevSegments = make([]Message15ElevSegment, m15.NumElevSegments)
	for i := range m15.ElevSegments {
		m15.ElevSegments[i].AzimuthSegments = make([]Message15AzimuthSegment, 360)
		for j := range m15.ElevSegments[i].AzimuthSegments {
			binary.Read(r, binary.BigEndian, &m15.ElevSegments[i].AzimuthSegments[j].NumRangeZones)
			m15.ElevSegments[i].AzimuthSegments[j].RangeZones = make([]Message15RangeZones, m15.ElevSegments[i].AzimuthSegments[j].NumRangeZones)
			for k := range m15.ElevSegments[i].AzimuthSegments[j].RangeZones {
				binary.Read(r, binary.BigEndian, &m15.ElevSegments[i].AzimuthSegments[j].RangeZones[k])
			}
		}
	}
}
