package geojson

import (
	"fmt"
	"strings"

	"github.com/bwiggs/go-nexrad/internal/geo"
)

func BinsToString(bins []*geo.Bin) *strings.Builder {
	var b strings.Builder

	fmt.Fprintf(&b, "{\"type\":\"FeatureCollection\",\"features\":[")

	stop := len(bins) - 1

	for i, bin := range bins {
		bin.AppendFeature(&b)

		if i != stop {
			fmt.Fprint(&b, ",")
		}
	}

	fmt.Fprintf(&b, "]}")

	return &b
}
