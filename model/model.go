package model

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Database struct {
	Connection *sql.DB
}

type GetAirline struct {
	Id                       int
	Airline_name             string
	Airline_img              string
	Source_airport_code      string
	Source_airport_city      string
	Destination_airport_code string
	Destination_airport_city string
	Depature_date            string
	Arrival_date             string
	Depature_time            string
	Arrival_time             string
	Flight_duration          string
	Reschedule               string
	Refund                   string
	Price                    int
}

type GetBooking struct {
	Id                       uint
	Airline_img              string
	Airline_name             string
	Source_airport_name      string
	Destination_airport_name string
	Depature_date            string
	Depature_time            string
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
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
		db, err := sql.Open(os.Getenv("DATABASE_PROVIDER"), os.Getenv("DATABASE"))
		if err != nil {
			log.Fatal(err)
		}
		instance = &Database{Connection: db}
	}

	return instance
}
