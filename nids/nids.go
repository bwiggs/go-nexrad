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
	ProductSymbologyBlock *ProductSymbologyBlock
}

// NewRPGProduct returns a new RPGProduct struct
func NewRPGProduct() *RPGProduct {
	return &RPGProduct{
		MessageHeader:         &MessageHeader{},
		GraphicProductMessage: &GraphicProductMessage{},
		ProductSymbologyBlock: &ProductSymbologyBlock{},
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
	_                   int16
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
	ProductDependent    [32]byte
	ProductDependent4   int16
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

// ProductSymbologyBlock of nids file
type ProductSymbologyBlock struct {
	_               int16
	BlockID         int16
	BlockSize       int32
	NumLayers       int16
	_               int16
	DataLayerLength int32
	// DisplayDataPackets int32
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

	buf := bytes.NewReader(b)

	// skip WMO header
	buf.Seek(54, io.SeekCurrent)

	rpg := NewRPGProduct()

	// MessageHeader
	if err := binary.Read(buf, binary.BigEndian, rpg.MessageHeader); err != nil {
		return nil, err
	}

	// GraphicProductMessage
	if err := binary.Read(buf, binary.BigEndian, rpg.GraphicProductMessage); err != nil {
		return nil, err
	}

	// preview(buf, 64)

	// ProductSymbologyBlock
	if err := binary.Read(buf, binary.BigEndian, rpg.ProductSymbologyBlock); err != nil {
		return nil, err
	}

	dataLayer := make([]byte, rpg.ProductSymbologyBlock.DataLayerLength)

	if err := binary.Read(buf, binary.BigEndian, dataLayer); err != nil {
		return nil, err
	}

	spew.Dump(dataLayer)

	preview(buf, 32)

	return rpg, nil
}

func preview(r io.ReadSeeker, n int) {
	preview := make([]byte, n)
	binary.Read(r, binary.BigEndian, &preview)
	spew.Dump(preview)
	r.Seek(-int64(n), io.SeekCurrent)
}
