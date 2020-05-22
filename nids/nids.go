package nids

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
)

type WMOHeader struct {
	Header [41]byte
}

// NewReader returns a nids reader
func NewReader(fileName string) (io.ReadCloser, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	// wmoh := WMOHeader{}
	// if err := binary.Read(f, binary.LittleEndian, &wmoh); err != nil {
	// 	return nil, err
	// }

	// skip the WMO header
	if _, err := f.Seek(41, io.SeekCurrent); err != nil {
		return nil, err
	}

	// nids data is zlib encoded
	return zlib.NewReader(f)
}

// RPGProduct contains data from a nids file
type RPGProduct struct {
	MessageHeader         *MessageHeader
	GraphicProductMessage *GraphicProductMessage
}

// NewRPGProduct returns a new RPGProduct struct
func NewRPGProduct() *RPGProduct {
	return &RPGProduct{
		MessageHeader:         &MessageHeader{},
		GraphicProductMessage: &GraphicProductMessage{},
	}
}

// MessageHeader of nids file
type MessageHeader struct {
	// Raw [18]byte
	MessageCode        int16
	ModifiedJulianDate int16
	ModifiedJulianTime int32
	MessageLength      int32
	SourceID           int16
	DestID             int16
	NumBlocks          int16
}

// GraphicProductMessage of nids file
type GraphicProductMessage struct {
	Lat                 int32
	Lng                 int32
	Height              int16
	ProductCode         int16
	OperationalMode     int16
	VCP                 int16
	Seq                 int16
	VolumeScanNum       int16
	VolumeScanDate      int16
	VolumeScanStartTime int32
	GenerationDate      int16
	GenerationTime      int32
	ProductDependent1   int16
	ProductDependent2   int16
	ElevationNum        int16
	ProductDependent3   int16
	ProductDependent4   int16
	ProductDependent    [15]byte
	ProductDependent5   int16
	ProductDependent6   int16
	ProductDependent7   int16
	ProductDependent8   int16
	ProductDependent9   int16
	ProductDependent10  int16
	Version             int8
	SpotBank            int8
	OffsetSymbology     int32
	OffsetToGraphic     int32
	OffsetToTabular     int32
}

// Open returns an RPGProduct
func Open(f string) (*RPGProduct, error) {
	r, err := NewReader(f)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	spew.Dump(b[:128])

	buf := bytes.NewReader(b)

	// skip WMO header
	buf.Seek(54, io.SeekCurrent)

	rpg := NewRPGProduct()

	if err := binary.Read(buf, binary.BigEndian, rpg.MessageHeader); err != nil {
		return nil, err
	}

	buf.Seek(2, io.SeekCurrent)

	if err := binary.Read(buf, binary.BigEndian, rpg.GraphicProductMessage); err != nil {
		return nil, err
	}

	return rpg, nil
}
