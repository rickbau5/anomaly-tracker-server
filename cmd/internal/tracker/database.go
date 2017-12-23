package tracker

import (
	"database/sql"
	"log"
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
