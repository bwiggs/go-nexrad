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

func preview(r io.ReadSeeker, n int) {
	preview := make([]byte, n)
	binary.Read(r, binary.BigEndian, &preview)
	spew.Dump(preview)
	r.Seek(-int64(n), io.SeekCurrent)
}

func decompress(f io.Reader, size int32) *bytes.Reader {
	start := time.Now()
	defer func() {
		logrus.Tracef("ar2: bz2 extracted %d Bytes in %s", size, time.Since(start))
	}()
	compressedData := make([]byte, size)
	binary.Read(f, binary.BigEndian, &compressedData)
	bz2Reader := bzip2.NewReader(bytes.NewReader(compressedData))
	extractedData := bytes.NewBuffer([]byte{})
	io.Copy(extractedData, bz2Reader)
	return bytes.NewReader(extractedData.Bytes())
}
