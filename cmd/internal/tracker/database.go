package tracker

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

var (
	appDB             *sql.DB
	insertAnomalyStmt *sql.Stmt
)

const layout = "2006-01-02 15:04:05"

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
	err := row.Scan(&anomaly.InternalID, &anomaly.ID, &anomaly.System, &str, &anomaly.Name)
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
	_, err = appDB.Exec(sql, anom.InternalID)
	if err != nil {
		log.Printf("Failed updating anomaly '%s' for key '%s'.\n", anom.ID, apiKey.Key)
		return nil, err
	}

	return getAnomalyByAnomalyID(anomaly.ID, apiKey.UserID)
}

func GetAnomaliesByAPIKey(apiKey APIKey) ([]Anomaly, error) {
	return GetAnomaliesInGroup(apiKey.GroupID)
}

func GetAnomaliesInGroup(groupID int) ([]Anomaly, error) {
	rows, err := appDB.Query(`
		SELECT
			id, anom_id, anom_system, anom_type, anom_name, user_id, group_id, created_dttm
		FROM
			anomaly_tracker.anomalies
		WHERE
			group_id = ?`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ScanAllAnomalies(rows)
}

func ScanAllAnomalies(rows *sql.Rows) ([]Anomaly, error) {
	var (
		anomalies []Anomaly
		err       error
	)
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, err
		}
		var (
			anomaly        Anomaly
			createdDttmStr string
		)
		rows.Scan(&anomaly.InternalID, &anomaly.ID, &anomaly.System, &anomaly.Type, &anomaly.Name, &anomaly.UserID, &anomaly.GroupID, &createdDttmStr)
		anomaly.Created, err = time.Parse(layout, createdDttmStr)
		if err != nil {
			log.Printf("Failed parsing anomaly time '%s': '%s'", createdDttmStr, err)
			continue
		}
		anomalies = append(anomalies, anomaly)
	}
	return anomalies, nil
}

func CheckAPIKey(apiKey string) (*APIKey, error) {
	row := appDB.QueryRow("SELECT id, `key`, type, user_id, group_id, created_by, created_dttm FROM anomaly_tracker.api_keys WHERE `key` = ?", apiKey)

	var (
		key            APIKey
		createdDttmStr string
	)
	err := row.Scan(&key.ID, &key.Key, &key.Type, &key.UserID, &key.GroupID, &key.CreatedBy, &createdDttmStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	key.CreatedDttm, err = time.Parse(layout, createdDttmStr)
	if err != nil {
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
	ID          int
	Key         string
	Type        string
	UserID      int
	GroupID     int
	CreatedBy   int
	CreatedDttm time.Time
}
