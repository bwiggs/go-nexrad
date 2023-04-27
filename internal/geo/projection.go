package geo

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/twpayne/go-proj/v10"
)

func createTransforms(radarLatitude float32, radarLongitude float32) []*proj.PJ {
	ltp := fmt.Sprintf("+proj=ortho +lat_0=%v +lon_0=%v +x_0=0 +y_0=0 +ellps=WGS84 +units=m +no_defs", radarLatitude, radarLongitude)

	geographic := "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs"

	ecef := "+proj=geocent +datum=WGS84 +units=m +no_defs +type=crs"

	ltpToEcef, err := proj.NewCRSToCRS(ltp, ecef, nil)

	if err != nil {
		logrus.Fatalln(err)
	}

	ecefToGeographic, err := proj.NewCRSToCRS(ecef, geographic, nil)

	if err != nil {
		logrus.Fatalln(err)
	}

	return []*proj.PJ{ltpToEcef, ecefToGeographic}
}
