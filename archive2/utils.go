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

func decompressBZ2(f io.Reader, size int32) *bytes.Reader {
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

// isCompressed return true if the file is compressed and string indicating the compression algorithm.
func isCompressed(f io.ReadSeeker) (bool, string) {
	header := make([]byte, 2)
	if _, err := f.Read(header); err != nil {
		logrus.Fatalf("isCompressed: failed to peek header: %s\n", err)
	}
	f.Seek(-2, io.SeekCurrent)
	headerString := string(header)
	switch headerString {
	case "BZ":
		return true, "bz2"
	case "\x1f\x8b":
		return true, "gz"
	}
	return false, ""
}
