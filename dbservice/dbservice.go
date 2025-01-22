package dbservice

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)


var Db *sql.DB


func SetupSqlDbconnection(properties map[string]string) {
	cfg := mysql.Config{
		User: properties["username"],
		Passwd: properties["password"],
		Net: "tcp",
		Addr: properties["hostname"]+":"+properties["port"],
		DBName: properties["dbname"],
		AllowNativePasswords: true,
	}
	var err error
	Db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil && Db.Ping() != nil {
		log.Fatal(err)
	} else{
		Db.SetConnMaxIdleTime(time.Minute * 5)
		Db.SetMaxOpenConns(20)
		Db.SetMaxIdleConns(20)
	}


	
}

func GetSqlDb() *sql.DB {
	if Db == nil {
		fmt.Println("db in dbservice is nil")
		return nil
	}

	return Db
}