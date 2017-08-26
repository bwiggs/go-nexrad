package main

import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"io"
	"os"

	"github.com/davecgh/go-spew/spew"
)

// VolumeHeaderRecord for NEXRAD Archive II Data Streams
//
// Description:
// The Volume Header Record
// The Volume Header Record is fixed length and contains information uniquely
// identifying the format and the data that follows. Sits at the beginning of the
// Archive II data stream.
//
// Volume Header Record Data Format:
// The first 9 bytes is a character constant of which the last 2 characters
// identify the version. The next 3 bytes is a numeric string field starting
// with the value 001 and increasing by one for each volume of radar data in the
// queue to a maximum value of 999. Once the maximum value is reached the value
// will be rolled over. The combined 12 bytes are called the Archive II filename.
// The next 4 bytes contain the NEXRAD-modified Julian date the volume was
// produced at the RDA followed by 4 bytes containing the time the volume was
// recorded. The date and time integer values are big Endian. The last 4 bytes
// contain a 4-letter radar identifier assigned by ICAO.
//
// Version Number Reference:
// Version 02: Super Resolution disabled at the RDA (pre RDA Build 12.0)
// Version 03: Super Resolution (pre RDA Build 12.0)
// Version 04: Recombined Super Resolution
// Version 05: Super Resolution disabled at the RDA (RDA Build 12.0 and later)
// Version 06: Super Resolution (RDA Build 12.0 and later)
// Version 07: Recombined Super Resolution (RDA Build 12.0 and later)
type VolumeHeaderRecord struct {
	Tape      [7]byte
	Version   [2]byte
	Extension [3]byte
	// Date NEXRAD- modified Julian
	Date [4]byte
	// Time ms since midnight
	Time [4]byte
	// ICAO Radar identifier in ASCII. The four uppercase character International Civil Aviation Organization identifier of the radar producing the data.
	ICAO [4]byte
}

// LDMRecord (Local Data Manager) contains NEXRAD message data.
// Following the Volume Header Record are variable-length records containing the
// Archive II data messages. These records are referred to as LDM Compressed Record(s).
type LDMRecord struct {
	Size           uint32
	MetaDataRecord []byte
}

type MessageHeader struct {
	MessageSize         uint16
	RDARedundantChannel uint8
	MessageType         uint8
	IDSequenceNumber    uint16
	JulianDate          uint16
	MillisOfDay         uint32
	NumMessageSegments  uint16
	MessageSegmentNum   uint16
}

// Message31 Digital Radar Data Generic Format
//
// Description:
// The message consists of base data information, that is, reflectivity, mean
// radial velocity, spectrum width, differential reflectivity, differential
// phase, correlation coefficient, azimuth angle, elevation angle, cut type,
// scanning strategy and calibration parameters. The frequency and volume of the
// message will be dependent on the scanning strategy and the type of data
// associated with that scanning strategy.
type Message31 struct {
	Header Message31Header
}

type Message31Header struct {
	RadarIdentifier          [4]byte
	CollectionTime           uint32
	ModifiedJulianDate       uint16
	AzimuthNumber            uint16
	AzimuthAngle             float32
	CompressionIndicator     uint8
	Spare                    uint8
	RadialLength             uint16
	AzimuthResolutionSpacing uint8
	RadialStatus             uint8
	ElevationNumber          uint8
	CutSectorNumber          uint8
	ElevationAngle           float32
	RadialSpotBlankingStatus uint8
	AzimuthIndexingMode      uint8
	DataBlockCount           uint16
	DataBlockPointers        [9]uint32
}

type DataBlock struct {
	DataBlockType [1]byte
	DataName      [3]byte
}

type VolumeData struct {
	DataBlock
	LRTUP                          uint16
	VersionMajor                   uint8
	VersionMinor                   uint8
	Lat                            float32
	Long                           float32
	SiteHeight                     uint16
	FeedhornHeight                 uint16
	CalibrationConstant            float32
	SHVTXPowerHor                  float32
	SHVTXPowerVer                  float32
	SystemDifferentialReflectivity float32
	InitialSystemDifferentialPhase float32
	VolumeCoveragePatternNumber    uint16
	ProcessingStatus               uint16
}

type ElevationData struct {
	DataBlock
	LRTUP      uint16
	ATMOS      [2]byte
	CalibConst float32
}

type RadialData struct {
	DataBlock
	LRTUP              uint16
	UnambiguousRange   uint16
	NoiseLevelHorz     float32
	NoiseLevelVert     float32
	NyquistVelocity    uint16
	Spares             [2]byte
	CalibConstHorzChan float32
	CalibConstVertChan float32
}

type GenericDataMoment struct {
	DataBlock
	Reserved                      uint32
	NumberDataMomentGates         uint16
	DataMomentRange               uint16
	DataMomentRangeSampleInterval uint16
	TOVER                         uint16
	SNRThreshold                  uint16
	ControlFlags                  [1]byte
	DataWordSize                  uint8
	Scale                         float32
	Offset                        float32
}

// RefData Reflectivity
type DataMoment struct {
	GenericDataMoment
	Data []byte
}

func main() {

	f, _ := os.Open("KEWX20170826_042554_V06")
	vhr := VolumeHeaderRecord{}

	binary.Read(f, binary.LittleEndian, &vhr)

	// ==============================
	// LDM Compressed Metadata Record
	// ==============================

	ldm := LDMRecord{}
	binary.Read(f, binary.BigEndian, &ldm.Size)

	compressedData := make([]byte, ldm.Size)
	binary.Read(f, binary.BigEndian, &compressedData)

	bz2Reader := bzip2.NewReader(bytes.NewReader(compressedData))

	meta := bytes.NewBuffer([]byte{})
	io.Copy(meta, bz2Reader)
	ldm.MetaDataRecord = meta.Bytes()

	messages := make([][2432]byte, 134)
	binary.Read(meta, binary.BigEndian, &messages)

	// ==============================
	// Message 31 Data
	// ==============================

	var message31Size uint32
	binary.Read(f, binary.BigEndian, &message31Size)
	m31Buf := decomp(f, message31Size)
	m31Reader := bytes.NewReader(m31Buf)
	m31Reader.Seek(12, io.SeekCurrent) // eat 12 empty bytes

	// header

	header := MessageHeader{}
	binary.Read(m31Reader, binary.BigEndian, &header)

	m31h := Message31Header{}
	binary.Read(m31Reader, binary.BigEndian, &m31h)

	for i := uint16(0); i < m31h.DataBlockCount; i++ {
		d := DataBlock{}
		binary.Read(m31Reader, binary.BigEndian, &d)
		m31Reader.Seek(-4, io.SeekCurrent)
		switch string(d.DataName[:]) {
		case "VOL":
			vd := VolumeData{}
			binary.Read(m31Reader, binary.BigEndian, &vd)
			spew.Dump(vd)
		case "ELV":
			ed := ElevationData{}
			binary.Read(m31Reader, binary.BigEndian, &ed)
			spew.Dump(ed)
		case "RAD":
			rad := RadialData{}
			binary.Read(m31Reader, binary.BigEndian, &rad)
			spew.Dump(rad)
		case "REF":
			fallthrough
		case "VEL":
			fallthrough
		case "SW":
			fallthrough
		case "ZDR":
			fallthrough
		case "PHI":
			fallthrough
		case "RHO":
			m := GenericDataMoment{}
			binary.Read(m31Reader, binary.BigEndian, &m)
			//LDM is the amount of space in bytes required for a data moment array and equals
			//((NG * DWS) / 8) where NG is the number of gates at the gate spacing resolution specified and DWS is the number of bits stored for each gate (DWS is always a multiple of 8).
			ldm := m.NumberDataMomentGates * uint16(m.DataWordSize) / 8
			data := make([]uint8, ldm)
			binary.Read(m31Reader, binary.BigEndian, &data)
			spew.Dump(data)
			// dm := DataMoment{
			// 	GenericDataMoment: m,
			// 	Data:              data,
			// }

			// spew.Dump(dm)
		}
	}

	// data blocks

	// volume := VolumeData{}
	// binary.Read(m31Reader, binary.BigEndian, &volume)

	// elev := ElevationData{}
	// binary.Read(m31Reader, binary.BigEndian, &elev)

	// radial := RadialData{}
	// binary.Read(m31Reader, binary.BigEndian, &radial)

	// ref := RefData{
	// 	Data: make([byte], )
	// }
	// binary.Read(m31Reader, binary.BigEndian, &ref)
	// spew.Dump(ref)

	// vel := VelData{}
	// binary.Read(m31Reader, binary.BigEndian, &vel)

	// spew.Dump(vel)
}

func decomp(f *os.File, size uint32) []byte {
	cdata := make([]byte, size)
	binary.Read(f, binary.BigEndian, &cdata)
	bz2Reader := bzip2.NewReader(bytes.NewReader(cdata))
	ddata := bytes.NewBuffer([]byte{})
	io.Copy(ddata, bz2Reader)
	return ddata.Bytes()
}

func message31Header() Message31Header {
	return Message31Header{}
}
