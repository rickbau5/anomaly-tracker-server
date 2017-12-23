package tracker

import (
	"errors"
	"log"
	"regexp"
	"strings"
)

type Anomaly struct {
	ID     string      `json:"id"`
	System string      `json:"system"`
	Type   AnomalyType `json:"type"`
	Name   string      `json:"name"`
}

// Anomaly Errors
var (
	ErrAnomalyMissingID     = errors.New("anomaly: missing ID")
	ErrAnomalyInvalidID     = errors.New("anomaly: invalid ID")
	ErrAnomalyMissingSystem = errors.New("anomaly: missing System")
	ErrAnomalyMissingType   = errors.New("anomaly: missing Type")
	ErrAnomalyInvalidType   = errors.New("anomaly: invalid Type")
	ErrAnomalyMissingName   = errors.New("anomaly: missing Name")

	idRegex = regexp.MustCompile(`^[A-Z0-9]{3}\-[A-Z0-9]{3}$`)
)

type AnomalyType string

// AnomalyTypes
const (
	CombatAnomaly  AnomalyType = "Combat"
	DataAnomaly    AnomalyType = "Data"
	RelicAnomaly   AnomalyType = "Relic"
	GasAnomaly     AnomalyType = "Gas"
	IceAnomaly     AnomalyType = "Ice"
	UnknownAnomaly AnomalyType = "unknown"
)

func GetAnomalyType(str string) AnomalyType {
	switch AnomalyType(str) {
	case CombatAnomaly:
		return CombatAnomaly
	case DataAnomaly:
		return DataAnomaly
	case RelicAnomaly:
		return RelicAnomaly
	case GasAnomaly:
		return GasAnomaly
	case IceAnomaly:
		return IceAnomaly
	default:
		return UnknownAnomaly
	}
}

func IsErrAnomaly(err error) bool {
	return strings.Contains(err.Error(), "anomaly:")
}

func (a *Anomaly) Validate() error {
	if a.ID == "" {
		return ErrAnomalyMissingID
	}
	if !idRegex.MatchString(a.ID) {
		return ErrAnomalyInvalidID
	}
	if a.System == "" {
		return ErrAnomalyMissingSystem
	}
	if a.Type == "" {
		return ErrAnomalyMissingType
	}
	anomalyType := GetAnomalyType(string(a.Type))
	if anomalyType == UnknownAnomaly {
		return ErrAnomalyInvalidType
	}
	a.Type = anomalyType
	if a.Name == "" {
		return ErrAnomalyMissingName
	}

	return nil
}

func AddAnomaly(anomaly Anomaly, apiKey APIKey) error {
	if err := anomaly.Validate(); err != nil {
		return err
	}

	if err := CommitAnomaly(anomaly, apiKey); err != nil {
		log.Println("Failed commiting anomaly to database:", err)
		if strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			return errors.New("Anomaly already exists")
		}
		return errors.New("Failed saving anomaly, try again later")
	}

	log.Println("Added anomaly:", anomaly)

	return nil
}
