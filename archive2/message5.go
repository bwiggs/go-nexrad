package archive2

const Message5Length = 960

// Message5 Volume Coverage Pattern Data
// see documentation RDA/RPG 3-54
type Message5 struct {
	Message5Header
	// ElevCuts contains info for each elevation angle
	ElevCuts []Message5ElevCut
}

type Message5Header struct {
	MessageSize         uint16
	PatternType         uint16
	PatternNumber       uint16
	NumElevCuts         uint16
	Version             uint8
	ClutterMapGroup     uint8
	DopplerVelocityRes  uint8
	PulseWidth          uint8
	_                   uint32
	VCPSequencing       uint16
	VCPSupplementalData uint16
	_                   uint16
}

type Message5ElevCut struct {
	ElevationAngle                    uint16
	ChannelConfiguration              uint8
	WaveformType                      uint8
	SuperResControl                   uint8
	SurveillancePRFNumber             uint8
	SurveillancePRFPulseCountRadial   uint16
	AzimuthRate                       uint16
	ReflectivityThreshold             uint16
	VelocityThreshold                 uint16
	SpectrumWidthThreshold            uint16
	DifferentialReflectivityThreshold uint16
	DifferentialPhaseThreshold        uint16
	CorrelationCoefficientThreshold   uint16
	EdgeAngle                         uint16
	DopplerPRFNumber                  uint16
	DopplerPRFPulseCountRadial        uint16
	SupplementalData                  uint16
	_                                 uint16
	_                                 uint16
	_                                 uint16
	EBCAngle                          uint16
	_                                 uint16
	_                                 uint16
	_                                 uint16
	_                                 uint16
}
