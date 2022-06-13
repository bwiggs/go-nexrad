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
	SwData           *DataMoment // SwData (Spectrum Width)
	ZdrData          *DataMoment // ZdrData (Differential Reflectivity) used to help identify hail shafts, detect updrafts, determine rain drop size, and identify aggregation of dry snow.
	PhiData          *DataMoment // PhiData (Differential Phase Shift)
	RhoData          *DataMoment // RhoData (Correlation Coefficient)
	CfpData          *DataMoment // CfpData (Clutter Filter Power Removed)
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

	// VOLDataBlockPtr uint32
	// ELVDataBlockPtr uint32
	// RADDataBlockPtr uint32
	// REFDataBlockPtr uint32
}

// AzimuthResolutionSpacing returns the spacing in degrees according to the AzimuthResolutionSpacingCode
func (h *Message31Header) AzimuthResolutionSpacing() float64 {
	if h.AzimuthResolutionSpacingCode == 1 {
		return 0.5
	}
	return 1
}

func msg31(r io.ReadSeeker) *Message31 {
	m31h := Message31Header{}

	// save the position of the first byte so we can easily process data blocks later.
	startPos, _ := r.Seek(0, io.SeekCurrent)

	binary.Read(r, binary.BigEndian, &m31h)

	m31 := Message31{
		Header: m31h,
	}

	// Process Data Block Pointers
	//
	// At this point we've read the M31 header up through the DataBlockCount. Now we
	// need to process an arbitrary amount of DataBlockPointers. As the RDA/RPG is
	// updated, more of these can show up. Ex: in build 19, the CFP data moment was
	// added, which also added an extra block pointer to the header.

	// we know the minimum number of pointers from the datablock count, read those
	// in, then read in 4 bytes until we see RVOL string.

	blockPointers := make([]uint32, m31h.DataBlockCount)
	if err := binary.Read(r, binary.BigEndian, blockPointers); err != nil {
		logrus.Panic(err.Error())
	}

	// check for more DataBlockPointers
	// keep reading 4 bytes until we see the RVOL moment start

	var lookahead = make([]byte, 4)
	hexRVOL := []byte("RVOL")
	maxLoops := 20
	for i := 0; true; i++ {
		if err := binary.Read(r, binary.BigEndian, &lookahead); err != nil {
			logrus.Panic(err.Error())
		}

		if bytes.Equal(lookahead, hexRVOL) {
			// backup 4 bytes to keep the block intact
			r.Seek(-4, io.SeekCurrent)
			break
		}

		ptr := binary.BigEndian.Uint32(lookahead)
		if ptr != 0 {
			blockPointers = append(blockPointers, ptr)
		}

		// prevent infinite loop
		if i == maxLoops {
			logrus.Fatal("M31 Header: failed to find the end of the datablock pointers.")
		}
		i++
	}

	// start processing DataBlocks
	for _, bptr := range blockPointers {
		// logrus.Tracef("ar2: m31: processing datablock %d", i)

		r.Seek(startPos+int64(bptr), io.SeekStart)

		d := DataBlock{}
		if err := binary.Read(r, binary.BigEndian, &d); err != nil {
			logrus.Panic(err.Error())
		}

		// rewind from reading the datalblock
		r.Seek(-4, io.SeekCurrent)

		blockName := string(d.DataName[:])

		switch blockName {
		case "VOL":
			binary.Read(r, binary.BigEndian, &m31.VolumeData)
		case "ELV":
			binary.Read(r, binary.BigEndian, &m31.ElevationData)
		case "RAD":
			binary.Read(r, binary.BigEndian, &m31.RadialData)
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

			data := make([]byte, ldm)
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
			case "CFP":
				m31.CfpData = d
			}
		default:
			// preview(r, 256)
			logrus.Panicf("Data Block - unknown type '%s'", blockName)
		}
	}
	return &m31
}
