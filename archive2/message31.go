package archive2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

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
	Header           Message31Header
	VolumeData       VolumeData
	ElevationData    ElevationData
	RadialData       RadialData
	ReflectivityData *DataMoment
	VelocityData     *DataMoment
	SwData           *DataMoment
	ZdrData          *DataMoment
	PhiData          *DataMoment
	RhoData          *DataMoment
}

func (h Message31Header) String() string {
	return fmt.Sprintf("Message 31 - %s @ %v deg=%.2f tilt=%.2f",
		string(h.RadarIdentifier[:]),
		h.Date(),
		h.AzimuthAngle,
		h.ElevationAngle,
	)
}

// Date and time this data is valid for
func (h Message31Header) Date() time.Time {
	return time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC).
		Add(time.Duration(h.CollectionDate) * time.Hour * 24).
		Add(time.Duration(h.CollectionTime) * time.Millisecond)
}

// Message31Header contains header information for an Archive 2 Message 31 type
type Message31Header struct {
	RadarIdentifier [4]byte
	// CollectionTime Radial data collection time in milliseconds past midnight GMT
	CollectionTime uint32
	// CollectionDate Current Julian date - 2440586.5
	CollectionDate uint16
	// AzimuthNumber Radial number within elevation scan
	AzimuthNumber uint16
	// AzimuthAngle Azimuth angle at which radial data was collected
	AzimuthAngle float32
	// CompressionIndicator Indicates if message type 31 is compressed and what method of compression is used. The Data Header Block is not compressed.
	CompressionIndicator uint8
	Spare                uint8
	// RadialLength Uncompressed length of the radial in bytes including the Data Header block length
	RadialLength uint16
	// AzimuthResolutionSpacing Code for the Azimuthal spacing between adjacent radials. 1 = .5 degrees, 2 = 1degree
	AzimuthResolutionSpacingCode uint8
	// RadialStatus Radial Status
	RadialStatus uint8
	// ElevationNumber Elevation number within volume scan
	ElevationNumber uint8
	// CutSectorNumber Sector Number within cut
	CutSectorNumber uint8
	// ElevationAngle Elevation angle at which radial radar data was collected
	ElevationAngle float32
	// RadialSpotBlankingStatus Spot blanking status for current radial, elevation scan and volume scan
	RadialSpotBlankingStatus uint8
	// AzimuthIndexingMode Azimuth indexing value (Set if azimuth angle is keyed to constant angles)
	AzimuthIndexingMode uint8
	DataBlockCount      uint16
	// DataBlockPointers   [10]uint32
	VOLDataBlockPtr uint32
	ELVDataBlockPtr uint32
	RADDataBlockPtr uint32
	// REFDataBlockPtr uint32
}

// AzimuthResolutionSpacing returns the spacing in degrees according to the AzimuthResolutionSpacingCode
func (h *Message31Header) AzimuthResolutionSpacing() float64 {
	if h.AzimuthResolutionSpacingCode == 1 {
		return 0.5
	}
	return 1
}

func msg31(r *bytes.Reader) *Message31 {
	m31h := Message31Header{}
	startPos, _ := r.Seek(0, io.SeekCurrent)

	binary.Read(r, binary.BigEndian, &m31h)

	m31 := Message31{
		Header: m31h,
	}

	// logrus.Tracef("ar2: m31: reading %d data blocks", m31h.DataBlockCount)

	// you will always get VOL, ELV and RAD. Then there's a a dynamic set of blocks after that.
	var err error
	_, err = r.Seek(int64(m31.Header.VOLDataBlockPtr)+startPos, io.SeekStart)
	if err != nil {
		logrus.Panicf("failed to seek to VOL pointer offset: %s", err)
	}
	binary.Read(r, binary.BigEndian, &m31.VolumeData)

	_, err = r.Seek(int64(m31.Header.ELVDataBlockPtr)+startPos, io.SeekStart)
	if err != nil {
		logrus.Panicf("failed to seek to ELV pointer offset: %s", err)
	}
	binary.Read(r, binary.BigEndian, &m31.ElevationData)

	_, err = r.Seek(int64(m31.Header.RADDataBlockPtr)+startPos, io.SeekStart)
	if err != nil {
		logrus.Panicf("failed to seek to RAD pointer offset: %s", err)
	}
	binary.Read(r, binary.BigEndian, &m31.RadialData)

	numAdditionalDataBlocks := m31h.DataBlockCount - 3

	for i := uint16(0); i < numAdditionalDataBlocks; i++ {
		// logrus.Tracef("ar2: m31: processing datablock %d", i)

		d := DataBlock{}
		if err := binary.Read(r, binary.BigEndian, &d); err != nil {
			logrus.Panic(err.Error())
		}

		// rewind, so we can reread the blockname in the struct processors
		r.Seek(-4, io.SeekCurrent)

		blockName := string(d.DataName[:])
		switch blockName {
		case "REF":
			fallthrough
		case "VEL":
			fallthrough
		case "CFP":
			fallthrough
		case "SW ":
			fallthrough
		case "ZDR":
			fallthrough
		case "PHI":
			fallthrough
		case "RHO":
			m := GenericDataMoment{}
			binary.Read(r, binary.BigEndian, &m)

			// LDM is the amount of space in bytes required for a data moment
			// array and equals ((NG * DWS) / 8) where NG is the number of gates
			// at the gate spacing resolution specified and DWS is the number of
			// bits stored for each gate (DWS is always a multiple of 8).
			ldm := m.NumberDataMomentGates * uint16(m.DataWordSize) / 8
			data := make([]uint8, ldm)
			binary.Read(r, binary.BigEndian, data)

			d := &DataMoment{
				GenericDataMoment: m,
				Data:              data,
			}

			switch blockName {
			case "REF":
				m31.ReflectivityData = d
			case "VEL":
				m31.VelocityData = d
			case "SW ":
				m31.SwData = d
			case "ZDR":
				m31.ZdrData = d
			case "PHI":
				m31.PhiData = d
			case "RHO":
				m31.RhoData = d
			}
		default:
			// preview(r, 256)
			logrus.Panicf("Data Block - unknown type '%s'", blockName)
		}
	}
	return &m31
}
