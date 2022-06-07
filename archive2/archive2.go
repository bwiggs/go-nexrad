package archive2

import (
	"encoding/binary"
	"fmt"
	"io"
	"sort"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)

// Archive2 wrapper for processed archive 2 data files.
type Archive2 struct {
	// ElevationScans contains all the messages for every elevation scan in the volume
	ElevationScans   map[int][]*Message31
	VolumeHeader     VolumeHeaderRecord
	RadarStatus      *Message2
	RadarPerformance *Message3
}

// Extract data from a given archive 2 data file.
func Extract(f io.ReadSeeker) *Archive2 {

	spew.Config.DisableMethods = true

	ar2 := Archive2{
		ElevationScans: make(map[int][]*Message31),
		VolumeHeader:   VolumeHeaderRecord{},
	}

	// -------------------------- Volume Header Record -------------------------
	// At the start of every volume is a 24-byte record describing certain attributes
	// of the radar data. The first 9 bytes is a character constant of which the
	// last 2 characters identify the version. The next 3 bytes is a numeric string
	// field starting with the value 001 and increasing by one for each volume of
	// radar data in the queue to a maximum value of 999. Once the maximum value is
	// reached the value will be rolled over. The combined 12 bytes are called the
	// Archive II filename.

	// read in the volume header record
	binary.Read(f, binary.BigEndian, &ar2.VolumeHeader)

	logrus.Debug(ar2.VolumeHeader)

	// ------------------------------ LDM Records ------------------------------

	// The first LDMRecord is the Metadata Record, consisting of 134 messages of
	// Metadata message types 15, 13, 18, 3, 5, and 2

	// Following the first LDM Metadata Record is a variable number of compressed
	// records containing 120 radial messages (type 31) plus 0 or more RDA Status
	// messages (type 2).

	for {
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

		logrus.Debugf("---------------- LDM Compressed Record (%d bytes)----------------", ldm.Size)

		msgBuf := decompress(f, ldm.Size)
		numMessages := 0
		messageCounts := map[uint8]int{}
		for {

			numMessages += 1

			// eat 12 bytes due to legacy compliance of CTM Header, these are all set to nil
			msgBuf.Seek(LegacyCTMHeaderLen, io.SeekCurrent)

			header := MessageHeader{}
			if err := binary.Read(msgBuf, binary.BigEndian, &header); err != nil {
				if err != io.EOF {
					logrus.Infof("processed %d messages", numMessages)
					logrus.Panic(err.Error())
				}
				break
			}

			logrus.WithFields(logrus.Fields{
				"type": header.MessageType,
				"seq":  header.IDSequenceNumber,
				"size": header.MessageSize,
			}).Tracef("== Message %d", header.MessageType)

			switch header.MessageType {
			case 2:
				m2 := Message2{}
				binary.Read(msgBuf, binary.BigEndian, &m2)
				// 68 is the size of a Message2 record
				msgBuf.Seek(MessageBodySize-68, io.SeekCurrent)
				// save a ref to most recent message 2
				ar2.RadarStatus = &m2
			case 3:
				m3 := Message3{}
				binary.Read(msgBuf, binary.BigEndian, &m3)
				msgBuf.Seek(MessageBodySize-960, io.SeekCurrent)
				ar2.RadarPerformance = &m3
			case 31:
				m31 := msg31(msgBuf)
				// logrus.Trace(m31.Header.String())
				ar2.ElevationScans[int(m31.Header.ElevationNumber)] = append(ar2.ElevationScans[int(m31.Header.ElevationNumber)], m31)
			default:
				_, err := msgBuf.Seek(MessageBodySize, io.SeekCurrent)
				if err != nil {
					logrus.Panic("failed to seek forward header message size")
				}
			}

			messageCounts[header.MessageType]++
		}
		logrus.Debugf("ar2: message types received: %v", messageCounts)
	}
	return &ar2
}

func (ar2 *Archive2) String() string {
	return fmt.Sprintf("%s\n%s", ar2.VolumeHeader, ar2.RadarStatus)
}

func (ar2 *Archive2) Elevations() []int {
	elevs := make([]int, len(ar2.ElevationScans))
	i := 0
	for k := range ar2.ElevationScans {
		elevs[i] = k
		i++
	}

	sort.Ints(elevs)
	return elevs
}
