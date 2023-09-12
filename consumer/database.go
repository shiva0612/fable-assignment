package main

import (
	"context"
	"fmt"
	"strings"

	"fable/utils"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	db *sqlx.DB
)

func configureDatabase() error {

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Dbhost, cfg.DbPort, cfg.DbUser, cfg.DbPsw, cfg.DbName)

	var err error
	db, err = sqlx.Connect("postgres", connectionString)
	if err != nil {
		Log.Error("Failed to connect to the database: " + err.Error())
		return err
	}
	db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(100)

	err = db.Ping()
	if err != nil {
		Log.Error("Failed to ping the database: " + err.Error())
		return err
	}

	Log.Info("Connected to the PostgreSQL database!")
	return nil
}

func writeToDB(ctx context.Context, lips []utils.LogInput) {
	var valueStrings []string
	var valueArgs []interface{}

	toBeInserted += int64(len(lips))
	for i, logInput := range lips {
		argNames := []string{
			fmt.Sprintf("$%d", i*4+1),
			fmt.Sprintf("$%d", i*4+2),
			fmt.Sprintf("$%d", i*4+3),
			fmt.Sprintf("$%d", i*4+4),
		}
		valueStrings = append(valueStrings, fmt.Sprintf("(%s)", strings.Join(argNames, ", ")))
		var event int
		if logInput.Event == "logout" {
			event = 1
		}
		valueArgs = append(valueArgs, logInput.Id, logInput.TimeStamp, logInput.UserId, event)
	}

	stmt := fmt.Sprintf("INSERT INTO logs (id, unix_ts, user_id, event_name) VALUES %s", strings.Join(valueStrings, ", "))

	// Execute the multi-row insert statement
	result, err := db.Exec(stmt, valueArgs...)
	if err != nil {
		Log.Debug("Failed to insert data:" + err.Error())
		failed += int64(len(lips))
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		Log.Debug("rows effected error : ", err.Error())
		return
	}
	inserted += int64(n)
}
