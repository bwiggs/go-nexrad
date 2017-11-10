package main

import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	// nrl2 := NEXRADLevel2File{
	// 	VolData: []VolumeData{},
	// 	ElvData: []ElevationData{},
	// 	RadData: []RadialData{},
	// 	Moments: make(map[string][]DataMoment),
	// }

	f, _ := os.Open("data/KCRP20170826_235827_V06")
	// f, _ := os.Open("data/KGRK20170827_234611_V06")

	vhr := VolumeHeaderRecord{}
	binary.Read(f, binary.BigEndian, &vhr)

	// First record consists of Metadata message types 15, 13, 18, 3, 5, and 2
	ldm := LDMRecord{}
	binary.Read(f, binary.BigEndian, &ldm.Size)
	decompress(f, ldm.Size)

	for true {
		logrus.Debugf("-------------------- LDM RECORD START --------------------")
		// preview(f, 32)
		ldm := LDMRecord{}

		if err := binary.Read(f, binary.BigEndian, &ldm.Size); err != nil {
			if err == io.EOF {
				logrus.Info("completed processing")
				break
			}
			logrus.Panic(err.Error())
		}
		logrus.Debugf("decompressing (%12d bytes)", ldm.Size)
		// preview(f, 16)

		msgBuf := decompress(f, ldm.Size)

		for true {
			msgBuf.Seek(12, io.SeekCurrent)
			header := MessageHeader{}
			if err := binary.Read(msgBuf, binary.BigEndian, &header); err != nil {
				if err != io.EOF {
					logrus.Panic(err.Error())
				}
				break
			}
			// if header.MessageType != 31 {
			logrus.Infof("=== Message %d", header.MessageType)
			// }
			// spew.Dump(header)
			switch header.MessageType {
			case 0:
				spew.Dump(header)
				msg := make([]byte, header.MessageSize)
				binary.Read(msgBuf, binary.BigEndian, &msg)
				spew.Dump(msg)
				return
			case 31:
				m31 := msg31(msgBuf)
				if m31.Header.ElevationNumber == 1 {
					logrus.Infof("\tAzimuth Number: %d", m31.Header.AzimuthNumber)
					logrus.Infof("\tAzimuth Angle: %f", m31.Header.AzimuthAngle)
					logrus.Infof("\tAzimuth Res Spacing: %d", m31.Header.AzimuthResolutionSpacing)
					logrus.Infof("\tElevation Angle: %f", m31.Header.ElevationAngle)
					logrus.Infof("\tElevation Number: %d", m31.Header.ElevationNumber)
					logrus.Infof("\tRadialStatus: %d", m31.Header.RadialStatus)
					logrus.Infof("\tRadial Length: %d", m31.Header.RadialLength)
					logrus.Infof("\tCut Sector Num: %d", m31.Header.CutSectorNumber)
				}

				dm := m31.MomentData[3].(DataMoment)
				spew.Dump(dm)
				for _, n := range dm.Data {
					spew.Dump((float32(n) - dm.Offset) / dm.Scale)
				}
				return
			case 2:
				// spew.Dump(header)
				m2 := RDAStatusMessage2{}
				binary.Read(msgBuf, binary.BigEndian, &m2)
				// spew.Dump(m2)
				// eat the rest of the record since we know it's 2432 bytes
				msg := make([]byte, 2432-16-54-12)
				binary.Read(msgBuf, binary.BigEndian, &msg)
				// spew.Dump(msg)
			default:
				spew.Dump(header)
				// eat the rest of the record since we know it's 2432 bytes (2416 - header)
				msg := make([]byte, 2416)
				binary.Read(msgBuf, binary.BigEndian, &msg)
				spew.Dump(msg)
			}
		}
	}
}

func preview(r io.ReadSeeker, n int) {
	preview := make([]byte, n)
	binary.Read(r, binary.BigEndian, &preview)
	spew.Dump(preview)
	r.Seek(-int64(n), io.SeekCurrent)
}

func decompress(f *os.File, size uint32) *bytes.Reader {
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
		Header:     m31h,
		MomentData: []interface{}{},
	}

	for i := uint16(0); i < m31h.DataBlockCount; i++ {
		d := DataBlock{}
		if err := binary.Read(r, binary.BigEndian, &d); err != nil {
			logrus.Panic(err.Error())
		}
		r.Seek(-4, io.SeekCurrent)

		// spew.Dump(d)

		blockName := string(d.DataName[:])
		// fmt.Printf("\t%s\n", blockName)
		switch blockName {
		case "VOL":
			d := VolumeData{}
			binary.Read(r, binary.BigEndian, &d)
			m31.MomentData = append(m31.MomentData, d)
		case "ELV":
			d := ElevationData{}
			binary.Read(r, binary.BigEndian, &d)
			m31.MomentData = append(m31.MomentData, d)
		case "RAD":
			d := RadialData{}
			binary.Read(r, binary.BigEndian, &d)
			m31.MomentData = append(m31.MomentData, d)
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
			//LDM is the amount of space in bytes required for a data moment array and equals
			//((NG * DWS) / 8) where NG is the number of gates at the gate spacing resolution specified and DWS is the number of bits stored for each gate (DWS is always a multiple of 8).
			ldm := m.NumberDataMomentGates * uint16(m.DataWordSize) / 8
			data := make([]uint8, ldm)
			binary.Read(r, binary.BigEndian, &data)
			d := DataMoment{
				GenericDataMoment: m,
				Data:              data,
			}
			m31.MomentData = append(m31.MomentData, d)
		default:
			logrus.Panicf("Data Block - unknown type '%s'", blockName)
		}
	}
	return &m31
}
