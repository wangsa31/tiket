package model

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	Connection *sql.DB
}

type GetAirline struct {
	Id                       int
	Airline_name             string
	Airline_img              string
	Source_airport_code      string
	Destination_airport_code string
	Depature_time            string
	Arrival_time             string
	Flight_duration          string
	Reschedule               string
	Refund                   string
	Price                    int
}

var instance *Database

// func () *Database {
// 	if instance == nil {
// 		// Konfigurasi koneksi database
// 		db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/test")
// 		if err != nil {
// 			panic(err)
// 		}

// 		instance = &Database{connection: db}
// 	}
// 	return instance
// }

func Connect() *Database {
	if instance == nil {
		db, err := sql.Open("mysql", "root:@/tiket")
		if err != nil {
			log.Fatal(err)
		}
		instance = &Database{Connection: db}
	}

	return instance
}
