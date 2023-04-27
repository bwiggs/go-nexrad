package archive2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	radialStatusStartOfElevationScan   = 0
	radialStatusIntermediateRadialData = 1
	radialStatusEndOfElevation         = 2
	radialStatusBeginningOfVolumeScan  = 3
	radialStatusEndOfVolumeScan        = 4
	radialStatusStartNewElevation      = 5

	LegacyCTMHeaderLen = 12
	MessageHeaderSize  = 16
	DefaultMessageSize = 2432
	// MessageBodySize is the size of the message without the legacy CTM header and Message Header. Note: this does not work for Message 31 types.
	MessageBodySize = DefaultMessageSize - LegacyCTMHeaderLen - MessageHeaderSize
)

// VolumeHeaderRecord for NEXRAD Archive II Data Streams
//
// Description:
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
	X_FileName [12]byte
	// ModifiedJulianDate NEXRAD date since 1970/1/1 = 1
	X_ModifiedJulianDate int32
	// X_ModifiedTime  ms since midnight
	X_ModifiedTime int32
	// ICAO Radar identifier in ASCII. The four uppercase character International Civil Aviation Organization identifier of the radar producing the data.
	ICAO [4]byte
}

// Date returns a time type representing the date of the scan capture
func (vh VolumeHeaderRecord) String() string {
	return fmt.Sprintf("File:%s ICAO:%s Date:%s", vh.FileName(), string(vh.ICAO[:]), vh.Date())
}

// Date returns a time type representing the date of the scan capture
func (vh VolumeHeaderRecord) Date() time.Time {
	return timeFromModifiedJulian(int(vh.X_ModifiedJulianDate), int(vh.X_ModifiedTime))
}

// FileName returns the name of the File
func (vh VolumeHeaderRecord) FileName() string {
	return string(vh.X_FileName[:])
}

func timeFromModifiedJulian(days, ms int) time.Time {
	return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).
		AddDate(0, 0, int(days-1)).
		Add(time.Duration(ms) * time.Millisecond)
}

// LDMRecord (Local Data Manager) contains NEXRAD message data.
// Following the Volume Header Record are variable-length records containing the
// Archive II data messages. These records are referred to as LDM Compressed Record(s).
type LDMRecord struct {
	Size           int32
	MetaDataRecord []byte
}

// MessageHeader wrapper for archive2 Message Headers
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

func (vh MessageHeader) Date() time.Time {
	return timeFromModifiedJulian(int(vh.JulianDate), int(vh.MillisOfDay))
}

// DataBlock wraps Data Block information
type DataBlock struct {
	DataBlockType [1]byte
	DataName      [3]byte
}

// VolumeData wraps information about the Volume being extracted
type VolumeData struct {
	DataBlock
	// LRTUP Size of data block in bytes
	LRTUP                          uint16
	VersionMajor                   uint8
	VersionMinor                   uint8
	Lat                            float32
	Lon                            float32
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

// ElevationData wraps Message 31 elevation data
type ElevationData struct {
	DataBlock
	// LRTUP Size of data block in bytes
	LRTUP uint16
	// ATMOS Atmospheric Attenuation Factor
	ATMOS [2]byte
	// CalibConst Scaling constant used by the Signal Processor for this elevation to calculate reflectivity
	CalibConst float32
}

// RadialData wraps Message 31 radial data
type RadialData struct {
	DataBlock
	// LRTUP Size of data block in bytes
	LRTUP uint16
	// UnambiguousRange, Interval Size
	UnambiguousRange   uint16
	NoiseLevelHorz     float32
	NoiseLevelVert     float32
	NyquistVelocity    uint16
	RadialFlags        uint16
	CalibConstHorzChan float32
	CalibConstVertChan float32
}

func (r RadialData) String() string {
	return fmt.Sprintf("[%s] %s LRTUP:%d NOISE:[%f %f]", r.DataBlockType, r.DataName, r.LRTUP, r.NoiseLevelHorz, r.NoiseLevelVert)
}

// GenericDataMoment is a generic data wrapper for momentary data. ex: REF, VEL, SW data
type GenericDataMoment struct {
	DataBlock
	Reserved uint32
	// NumberDataMomentGates Number of data moment gates for current radial
	NumberDataMomentGates uint16
	// DataMomentRange Range to center of first range gate
	DataMomentRange uint16
	// DataMomentRangeSampleInterval Size of data moment sample interval
	DataMomentRangeSampleInterval uint16
	// TOVER Threshold parameter which specifies the minimum difference in echo power between two resolution gates for them not to be labeled "overlayed"
	TOVER uint16
	// SNRThreshold SNR threshold for valid data
	SNRThreshold uint16
	// ControlFlags Indicates special control features
	ControlFlags uint8
	// DataWordSize Number of bits (DWS) used for storing data for each Data Moment gate
	DataWordSize uint8
	// Scale value used to convert Data Moments from integer to floating point data
	Scale float32
	// Offset value used to convert Data Moments from integer to floating point data
	Offset float32
}

// DataMoment wraps all Momentary data records. ex: REF, VEL, SW data
type DataMoment struct {
	GenericDataMoment
	Data []byte
}

const MomentDataBelowThreshold = 999
const MomentDataFolded = 998

// ScaledData automatically scales the nexrad moment values to their actual values.
// For all data moment integer values N = 0 indicates received signal is below
// threshold and N = 1 indicates range folded data. Actual data range is N = 2
// through 255, or 1023 for data resolution size 8, and 10 bits respectively.
func (d *DataMoment) ScaledData() []float32 {

	gates := make([]uint16, d.NumberDataMomentGates)

	if d.DataWordSize == 8 {
		for i, v := range d.Data {
			gates[i] = uint16(v)
		}
	} else if d.DataWordSize == 16 {
		r := bytes.NewReader(d.Data)
		binary.Read(r, binary.BigEndian, gates)
	}

	scaledData := []float32{}
	for _, v := range gates {
		if v == 0 {
			// below threshold
			scaledData = append(scaledData, MomentDataBelowThreshold)
		} else if v == 1 {
			// range folded
			scaledData = append(scaledData, MomentDataFolded)
		} else {
			scaledData = append(scaledData, scaleUint(uint16(v), d.GenericDataMoment.Offset, d.GenericDataMoment.Scale))
		}
	}

	return scaledData
}

// scaleUint converts unsigned integer data that can be converted to floating point
// data using the Scale and Offset fields, i.e., F = (N - OFFSET) / SCALE where
// N is the integer data value and F is the resulting floating point value. A
// scale value of 0 indicates floating point moment data for each range gate.
func scaleUint(n uint16, offset, scale float32) float32 {
	if scale == 0 {
		return float32(n)
	}
	return (float32(n) - offset) / scale
}
