package tracker

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

var (
	appDB             *sql.DB
	insertAnomalyStmt *sql.Stmt
)

func InitializeAppDB(conf AppConfig) error {
	mysqlConfig := conf.BuildMySQLConfig()
	dsn := mysqlConfig.FormatDSN()
	log.Println("Connecting to", mysqlConfig.Addr)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("Failed opening mysql connection:", mysqlConfig.Addr, mysqlConfig.User)
		return err
	}
	log.Println("Connected to app db.")
	appDB = db

	insertAnomalyStmt, err = appDB.Prepare(`
		INSERT anomaly_tracker.anomalies 
			(anom_id, anom_system, anom_type, anom_name, user_id) 
		VALUES 
			( ?, ?, ?, ?, ? )`)
	if err != nil {
		log.Println("Failed prepareing insert statement:", err.Error())
		return err
	}

	log.Println("AppDB initialized.")

	return nil
}

func CommitAnomaly(anomaly Anomaly, apiKey APIKey) error {
	_, err := insertAnomalyStmt.Exec(anomaly.ID, anomaly.System, string(anomaly.Type), anomaly.Name, apiKey.UserID)
	return err
}

func getAnomalyByAnomalyID(anomalyID string, userID int) (*Anomaly, error) {
	row := appDB.QueryRow(`SELECT id, anom_id, anom_system, anom_type, anom_name FROM anomaly_tracker.anomalies where anom_id = ? and user_id = ?`,
		anomalyID, userID)
	anomaly := Anomaly{}
	var str string
	err := row.Scan(&anomaly.id, &anomaly.ID, &anomaly.System, &str, &anomaly.Name)
	anomaly.Type = GetAnomalyType(str)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAnomalyNotFound
		}
		return nil, err
	}
	return &anomaly, nil
}

func DeleteAnomaly(anomaly Anomaly, apiKey APIKey) error {
	row := appDB.QueryRow(
		"SELECT id FROM anomaly_tracker.anomalies where anom_id = ? and user_id = ?",
		anomaly.ID, apiKey.UserID,
	)
	var anomalyDBID int
	if err := row.Scan(&anomalyDBID); err != nil {
		if err == sql.ErrNoRows {
			return ErrAnomalyNotFound
		}
		return err
	}
	log.Printf("Deleting anomaly id '%s' (%d) for API key '%s'.\n", anomaly.ID, anomalyDBID, apiKey.Key)
	res, err := appDB.Exec("DELETE FROM anomaly_tracker.anomalies where id = ?", anomalyDBID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	log.Printf("Deleted %d anomaly.\n", affected)
	return nil
}

func UpdateAnomaly(anomaly Anomaly, apiKey APIKey) (*Anomaly, error) {
	anom, err := getAnomalyByAnomalyID(anomaly.ID, apiKey.UserID)
	if err != nil {
		return nil, err
	}
	sql := `UPDATE anomaly_tracker.anomalies SET `
	toUpdate := make(map[string]string)
	if anomaly.Name != "" {
		toUpdate["anom_name"] = anomaly.Name
	}
	if anomaly.Type != "" {
		toUpdate["anom_type"] = string(anomaly.Type)
	}
	var updates []string
	for k, v := range toUpdate {
		updates = append(updates, fmt.Sprintf("%s = '%s'", k, v))
	}
	sql += strings.Join(updates, ", ") + " WHERE id = ?"
	log.Println("Update query:", sql)
	_, err = appDB.Exec(sql, anom.id)
	if err != nil {
		log.Printf("Failed updating anomaly '%s' for key '%s'.\n", anom.ID, apiKey.Key)
		return nil, err
	}

	return getAnomalyByAnomalyID(anomaly.ID, apiKey.UserID)
}

func CheckAPIKey(apiKey string) (*APIKey, error) {
	row := appDB.QueryRow("SELECT * FROM anomaly_tracker.api_keys where `key` = ?;", apiKey)
	var key APIKey
	err := row.Scan(&key.ID, &key.Key, &key.Type, &key.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &key, nil
}

func CleanupAppDB() {
	if appDB != nil {
		appDB.Close()
	}
	if insertAnomalyStmt != nil {
		insertAnomalyStmt.Close()
	}
}

type APIKey struct {
	ID     int
	Key    string
	Type   string
	UserID int
}
