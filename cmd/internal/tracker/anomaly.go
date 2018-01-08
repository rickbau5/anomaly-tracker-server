package tracker

import (
	"database/sql"
	"errors"
	"log"
	"regexp"
	"strings"

	atp "github.com/rickbau5/anomaly-tracker-proto"
)

// Anomaly Errors
var (
	ErrAnomalyMissingID     = errors.New("anomaly: missing ID")
	ErrAnomalyInvalidID     = errors.New("anomaly: invalid ID")
	ErrAnomalyMissingSystem = errors.New("anomaly: missing System")
	ErrAnomalyMissingType   = errors.New("anomaly: missing Type")
	ErrAnomalyInvalidType   = errors.New("anomaly: invalid Type")
	ErrAnomalyMissingName   = errors.New("anomaly: missing Name")
	ErrAnomalyNotFound      = errors.New("anomaly: anomaly not found")

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

func IsErrAnomaly(err error) bool {
	return strings.Contains(err.Error(), "anomaly:")
}

func GetErrAnomalyMessage(err error) string {
	str := strings.Replace(err.Error(), "anomaly: ", "", 1)
	if len(str) == len(err.Error()) {
		// Not an anomaly error
		return ""
	}
	return str
}

func AddAnomaly(anomaly atp.Anomaly, apiKey APIKey) error {
	if !anomaly.Publishable() {
		log.Println("Cannot commit incomplete anomaly.")
		return errors.New("anomaly: incomplete details")
	}
	if err := CommitAnomaly(anomaly, apiKey); err != nil {
		log.Println("Failed commiting anomaly to database:", err)
		if strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			return errors.New("anomaly: Anomaly already exists")
		}
		return errors.New("Failed saving anomaly, try again later")
	}

	log.Println("Added anomaly:", anomaly.Id)

	return nil
}

func ModifyAnomaly(anomaly atp.Anomaly, apiKey APIKey) (*atp.Anomaly, error) {
	if !idRegex.MatchString(anomaly.GetId()) {
		return nil, ErrAnomalyInvalidID
	}
	if anomaly.GetSystem() != "" {
		return nil, errors.New("anomaly: cannot update System")
	}
	if anomaly.GetType() == "" && anomaly.GetName() == "" {
		return nil, errors.New("anomaly: must specify fields to update")
	}

	updatedAnomaly, err := UpdateAnomaly(anomaly, apiKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnomalyNotFound
		}
		return nil, err
	}

	log.Printf("Updated anomaly for key '%s': ", apiKey.Key)

	return updatedAnomaly, nil
}
