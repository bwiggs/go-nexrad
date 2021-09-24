package archive2

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Message2 RDA Status Data (User 3.2.4.6)
type Message2 struct {
	RDAStatus                       uint16
	OperabilityStatus               uint16
	ControlStatus                   uint16
	AuxPowerGeneratorState          uint16
	AvgTxPower                      uint16
	HorizRefCalibCorr               uint16
	DataTxEnabled                   uint16
	VolumeCoveragePatternNum        uint16
	RDAControlAuth                  uint16
	RDABuild                        uint16
	OperationalMode                 uint16
	SuperResStatus                  uint16
	ClutterMitigationDecisionStatus uint16
	AvsetStatus                     uint16
	RDAAlarmSummary                 uint16
	CommandAck                      uint16
	ChannelControlStatus            uint16
	SpotBlankingStatus              uint16
	BypassMapGenDate                uint16
	BypassMapGenTime                uint16
	ClutterFilterMapGenDate         uint16
	ClutterFilterMapGenTime         uint16
	VertRefCalibCorr                uint16
	TransitionPwrSourceStatus       uint16
	RMSControlStatus                uint16
	PerformanceCheckStatus          uint16
	AlarmCodes                      uint16
	Spares                          [14]byte
}

func (m2 Message2) String() string {
	return fmt.Sprintf("Message 2 - %s and %s. VCP %d build %.2f",
		m2.GetRDAStatus(),
		m2.GetOperabilityStatus(),
		m2.VolumeCoveragePatternNum,
		m2.GetBuildNumber(),
	)
}

// GetRDAStatus returns a human friendly status
func (m2 Message2) GetRDAStatus() string {
	mapping := map[uint16]string{
		2:  "start-up",
		4:  "standby",
		8:  "restart",
		16: "operating",
		32: "spare",
		64: "spare",
	}

	if val, ok := mapping[m2.RDAStatus]; ok {
		return val
	}

	logrus.Warnf("unknown RDA status code %d", m2.RDAStatus)
	return "UNKNOWN"
}

// GetOperabilityStatus returns a human friendly status
func (m2 Message2) GetOperabilityStatus() string {
	mapping := map[uint16]string{
		2:  "online",
		4:  "maintenance required",
		8:  "maintenance mandatory",
		16: "commanded shut down",
		32: "inoperable",
	}

	if val, ok := mapping[m2.OperabilityStatus]; ok {
		return val
	}

	logrus.Warnf("unknown operability status code %d", m2.RDAStatus)
	return "UNKNOWN"
}

// GetBuildNumber as a more recognizable float
func (m2 Message2) GetBuildNumber() float32 {
	return float32(m2.RDABuild / 100)
}
