package archive2

import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"io"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)

// Archive2 wrapper for processed archive 2 data files.
type Archive2 struct {
	// ElevationScans contains all the messages for every elevation scan in the volume
	ElevationScans map[int][]*Message31
	VolumeHeader   VolumeHeaderRecord
}

// Extract data from a given archive 2 data file.
func Extract(f io.ReadSeeker) *Archive2 {

	ar2 := Archive2{
		ElevationScans: make(map[int][]*Message31),
		VolumeHeader:   VolumeHeaderRecord{},
	}

	// -------------------------- Volume Header Record -------------------------

	// read in the volume header record
	binary.Read(f, binary.BigEndian, &ar2.VolumeHeader)

	// ------------------------------ LDM Records ------------------------------

	// The first LDMRecord is the Metadata Record, consisting of 134 messages of
	// Metadata message types 15, 13, 18, 3, 5, and 2

	// Following the first LDM Metadata Record is a variable number of compressed
	// records containing 120 radial messages (type 31) plus 0 or more RDA Status
	// messages (type 2).

	skipLDMRecord := true
	for true {
		ldm := LDMRecord{}

		// read in control word (size) of LDM record
		if err := binary.Read(f, binary.BigEndian, &ldm.Size); err != nil {
			if err != io.EOF {
				logrus.Panic(err.Error())
			}
			return &ar2
		}

		// As the control word contains a negative size under some circumstances,
		// the absolute value of the control word must be used for determining
		// the size of the block.
		if ldm.Size < 0 {
			ldm.Size = -ldm.Size
		}

		// logrus.Debug("---------------- LDM Compressed Record ----------------")

		msgBuf := decompress(f, ldm.Size)

		for true {

			if skipLDMRecord {
				skipLDMRecord = false
				break
			}

			// eat 12 bytes due to legacy compliance of CTM Header, these are all set to nil
			msgBuf.Seek(LegacyCTMHeaderLen, io.SeekCurrent)

			header := MessageHeader{}
			if err := binary.Read(msgBuf, binary.BigEndian, &header); err != nil {
				if err != io.EOF {
					logrus.Panic(err.Error())
				}
				break
			}
			// logrus.Debugf("== Message %d (i: %d, size: %d)", header.MessageType, header.IDSequenceNumber, header.MessageSize)

			// spew.Dump(header)
			// time.Sleep(1 * time.Second)

			switch header.MessageType {
			case 0:
				msg := make([]byte, header.MessageSize)
				binary.Read(msgBuf, binary.BigEndian, &msg)
			case 31:
				m31 := msg31(msgBuf)
				logrus.Tracef("%s (%3d) ɑ=%7.3f elv=%2d tilt=%5f status=%d", m31.Header.RadarIdentifier, m31.Header.AzimuthNumber, m31.Header.AzimuthAngle, m31.Header.ElevationNumber, m31.Header.ElevationAngle, m31.Header.RadialStatus)

				// if m31.Header.ElevationNumber > 1 {
				// 	return &ar2
				// }
				ar2.ElevationScans[int(m31.Header.ElevationNumber)] = append(ar2.ElevationScans[int(m31.Header.ElevationNumber)], m31)
				// return &ar2
				if m31.VelocityData != nil {
					// logrus.Warn("VelocityData")
				}
			case 2:
				m2 := Message2{}
				binary.Read(msgBuf, binary.BigEndian, &m2)
				// eat the rest of the record since we know it's 2432 bytes
				msg := make([]byte, 2432-16-54-12)
				binary.Read(msgBuf, binary.BigEndian, &msg)
				// spew.Dump(m2, msg)
			// case 15:
			// 	msg := make([]byte, header.MessageSize)
			// 	binary.Read(msgBuf, binary.BigEndian, &msg)
			// 	spew.Dump(msg)
			default:
				msg := make([]byte, header.MessageSize)
				binary.Read(msgBuf, binary.BigEndian, &msg)
			}
		}
	}
	return &ar2
}

func preview(r io.ReadSeeker, n int) {
	preview := make([]byte, n)
	binary.Read(r, binary.BigEndian, &preview)
	spew.Dump(preview)
	r.Seek(-int64(n), io.SeekCurrent)
}

func decompress(f io.Reader, size int32) *bytes.Reader {
	start := time.Now()
	defer func() {
		logrus.Tracef("bz2 extract: %s", time.Since(start))
	}()
	compressedData := make([]byte, size)
	binary.Read(f, binary.BigEndian, &compressedData)
	bz2Reader := bzip2.NewReader(bytes.NewReader(compressedData))
	extractedData := bytes.NewBuffer([]byte{})
	io.Copy(extractedData, bz2Reader)
	return bytes.NewReader(extractedData.Bytes())
}

func msg31(r *bytes.Reader) *Message31 {
	m31h := Message31Header{}
	binary.Read(r, binary.BigEndian, &m31h)

	m31 := Message31{
		Header: m31h,
	}

	for i := uint16(0); i < m31h.DataBlockCount; i++ {
		d := DataBlock{}
		if err := binary.Read(r, binary.BigEndian, &d); err != nil {
			logrus.Panic(err.Error())
		}

		// rewind
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
			binary.Read(r, binary.BigEndian, &data)

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
			logrus.Panicf("Data Block - unknown type '%s'", blockName)
		}
	}
	return &m31
}
