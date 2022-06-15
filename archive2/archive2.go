package archive2

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)

// Archive2MetadataRecordLen "The first LDM Compressed Record contains the
// Archive II messages comprising the Archive II metadata. The size of the
// uncompressed metadata is fixed at 134 messages, ie. 325888 bytes."
const Archive2MetadataRecordLen = 325888

// Archive2 wrapper for processed archive 2 data files.
type Archive2 struct {
	// ElevationScans contains all the messages for every elevation scan in the volume
	ElevationScans map[int][]*Message31
	VolumeHeader   VolumeHeaderRecord
	//RadarStatus is a container for the Message2 record from the LDM metadata
	RadarStatus *Message2
	//RadarPerformance is a container for the Message3 record from the LDM metadata
	RadarPerformance *Message3
	VCP              *Message5
}

// Extract data from a given archive 2 data file.
func Extract(f io.ReadSeeker) *Archive2 {
	ar2ExtractTimeStart := time.Now()
	defer func() {
		logrus.Debugf("ar2: done %s", time.Since(ar2ExtractTimeStart))
	}()
	spew.Config.DisableMethods = true

	ar2 := Archive2{
		ElevationScans: make(map[int][]*Message31),
		VolumeHeader:   VolumeHeaderRecord{},
	}

	// older archive2 files are gzipped, check for those and decompress if found
	if yes, ctype := isCompressed(f); yes {
		if ctype != "gz" {
			logrus.Fatalf("unsupported compression %s", ctype)
		}
		var gzd *gzip.Reader
		var err error
		if gzd, err = gzip.NewReader(f); err != nil {
			logrus.Fatalf("failed to open gzip file: %s", err)
		}
		gzb, err := ioutil.ReadAll(gzd)
		if err != nil {
			logrus.Fatalf("failed to read gzip file: %s", err)
		}
		f = bytes.NewReader(gzb)
	}

	// -------------------------- Volume Header Record -------------------------
	// At the start of every volume is a 24-byte record describing certain attributes
	// of the radar data. The first 9 bytes is a character constant of which the
	// last 2 characters identify the version. The next 3 bytes is a numeric string
	// field starting with the value 001 and increasing by one for each volume of
	// radar data in the queue to a maximum value of 999. Once the maximum value is
	// reached the value will be rolled over. The combined 12 bytes are called the
	// Archive II filename.

	// read in the 24 byte volume header record
	binary.Read(f, binary.BigEndian, &ar2.VolumeHeader)

	logrus.Debug(ar2.VolumeHeader)

	for {

		// ------------------------------ LDM Records ------------------------------
		// The first LDMRecord is the Metadata Record, consisting of 134 messages of
		// Metadata message types 15, 13, 18, 3, 5, and 2
		//
		// Following the first LDM Metadata Record is a variable number of compressed
		// records containing 120 radial messages (type 31) plus 0 or more RDA Status
		// messages (type 2).

		ldmExtractTimeStart := time.Now()

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
		} else if ldm.Size == 0 {
			// older files don't have LDM records? Backup 4 bytes (int32 for size)
			f.Seek(-4, io.SeekCurrent)
		}

		logrus.WithFields(logrus.Fields{
			"size": ldm.Size,
		}).Tracef("ar2: ldm: new LDM record")

		var msgBuf io.ReadSeeker
		if c, _ := isCompressed(f); c {
			logrus.Tracef("ar2: ldm: decompressing %d bytes", ldm.Size)
			msgBuf = decompressBZ2(f, ldm.Size)
		} else {
			msgBuf = f
		}

		numMessages := 0
		messageCounts := map[uint8]int{}

		for {

			numMessages += 1

			// CTM Header
			// "The Archive II raw data format contains a 28-byte header. The
			// first 12 bytes are empty, which means the "Message Size" does not
			// begin until byte 13 (halfword 7 or full word 4). This 12 byte
			// offset is due to legacy compliance (previously known as the "CTM
			//header"). See the RDA/RPG ICD for more details (Message Header Data)
			msgBuf.Seek(LegacyCTMHeaderLen, io.SeekCurrent)

			msgHeader := MessageHeader{}
			if err := binary.Read(msgBuf, binary.BigEndian, &msgHeader); err != nil {
				if err != io.EOF {
					logrus.Debugf("processed %d messages", numMessages)
					logrus.Panic(err.Error())
				}
				break
			}

			logrus.WithFields(logrus.Fields{
				"type":          msgHeader.MessageType,
				"seq":           msgHeader.IDSequenceNumber,
				"size":          msgHeader.MessageSize,
				"segments":      msgHeader.NumMessageSegments,
				"segmentNumber": msgHeader.MessageSegmentNum,
				"date":          msgHeader.Date(),
			}).Tracef("ar2: ldm: processing message %d", msgHeader.MessageType)

			switch msgHeader.MessageType {
			case 2:
				m2 := Message2{}
				binary.Read(msgBuf, binary.BigEndian, &m2)
				ar2.RadarStatus = &m2

				// move to the end of the message
				msgBuf.Seek(MessageBodySize-Message2Length, io.SeekCurrent)
			case 3:
				m3 := Message3{}
				binary.Read(msgBuf, binary.BigEndian, &m3)
				ar2.RadarPerformance = &m3

				// move to the end of the message
				msgBuf.Seek(MessageBodySize-Message3Length, io.SeekCurrent)
			// case 5:
			// 	m5 := Message5{}
			// 	binary.Read(msgBuf, binary.BigEndian, &m5.Message5Header)
			// 	ar2.VCP = &m5

			// 	m5.ElevCuts = make([]Message5ElevCut, m5.NumElevCuts)
			// 	binary.Read(msgBuf, binary.BigEndian, &m5.ElevCuts)

			// 	// move to the end of the message
			// 	msgBuf.Seek(MessageBodySize-int64(msgHeader.MessageSize), io.SeekCurrent)
			// case 15:
			// 	m15 := Message15{}
			// 	m15.Read(msgBuf)

			// 	// move to the end of the message
			// 	msgBuf.Seek(MessageBodySize-int64(msgHeader.MessageSize), io.SeekCurrent)
			case 31:
				m31 := msg31(msgBuf)
				// logrus.Trace(m31.Header.String())
				ar2.ElevationScans[int(m31.Header.ElevationNumber)] = append(ar2.ElevationScans[int(m31.Header.ElevationNumber)], m31)
			default:
				if msgHeader.MessageType != 0 {
					logrus.Debugf("ar2: unhandled message: %d", msgHeader.MessageType)
				}
				_, err := msgBuf.Seek(MessageBodySize, io.SeekCurrent)
				if err != nil {
					logrus.Panic("failed to seek forward header message size")
				}
			}

			messageCounts[msgHeader.MessageType]++
		}
		logrus.Tracef("ar2: ldm: done: %s messages:%v", time.Since(ldmExtractTimeStart), messageCounts)
	}
	return &ar2
}

func (ar2 *Archive2) String() string {
	return fmt.Sprintf("-- %s\n-- %s", ar2.VolumeHeader, ar2.RadarStatus)
}

func (ar2 *Archive2) Lon() string {
	return fmt.Sprintf("-- %s\n-- %s", ar2.VolumeHeader, ar2.RadarStatus)
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
