package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/GoWeb/tiket/model"
	"github.com/GoWeb/tiket/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("thetiketdotcom"))

type Googleuser struct {
	Fullname string `json:"name"`
}

func Login(w http.ResponseWriter, req *http.Request) {
	html, err := template.ParseFiles("view/login.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(w, req)
}

func LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	config := utils.GetGoogleAuthConfig()
	state := "randomstate"

	url := utils.GetGoogleLoginURL(config, state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func Callback(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	config := utils.GetGoogleAuthConfig()

	userInfo, err := utils.GetGoogleUserInfo(config, code)

	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	json_data := []byte(userInfo)

	var data utils.GoogleInfo

	err = json.Unmarshal(json_data, &data)

	if err != nil {
		log.Fatal(err)
	}

	db := model.Connect()

	// chcek if email not availabel
	var count int8
	err = db.Connection.QueryRow("SELECT COUNT(*) FROM user WHERE email = ?", data.Email).Scan(&count)

	if err != nil {
		log.Fatal(err)
	}

	if count > 0 {

		session, _ := store.Get(req, "users")
		session.Values["Islogin"] = true
		session.Values["name"] = data.Fullname
		session.Values["img"] = data.Picture
		// Save it before we write to the response/return from the handler.
		err := sessions.Save(req, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusTemporaryRedirect)
			return
		}
		http.Redirect(w, req, "/tiket", http.StatusPermanentRedirect)

	} else {
		stmt, err := db.Connection.Prepare("INSERT INTO user (oauth_id, name, email, oauth_provide, image) VALUES(?,?,?,'google',?)")
		if err != nil {
			log.Fatal(err)
		}

		defer stmt.Close()

		_, err = stmt.Exec(data.Provide_id, data.Fullname, data.Email, data.Picture)
		if err != nil {
			log.Fatal(err)
		}

		session, _ := store.Get(req, "users")
		session.Values["Islogin"] = true
		session.Values["name"] = "ok"
		session.Values["img"] = data.Picture
		// Save it before we write to the response/return from the handler.
		err = sessions.Save(req, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, req, "/tiket", http.StatusTemporaryRedirect)
	}

}

func Register(w http.ResponseWriter, req *http.Request) {
	html, err := template.ParseFiles("view/register.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(w, req)
}

func Dashboard(w http.ResponseWriter, req *http.Request) {

	// code := r.URL.Query().Get("code")
	// config := utils.GetGoogleAuthConfig()

	// userInfo, err := utils.GetGoogleUserInfo(config, code)
	// if err != nil {
	// 	http.Error(w, "Failed to get user info", http.StatusInternalServerError)
	// 	return
	// }

	session, _ := store.Get(req, "users")

	login := session.Values["Islogin"]

	user := session.Values["name"]
	img := session.Values["img"]

	session.Save(req, w)

	html, err := template.ParseFiles("view/index.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(w, map[string]interface{}{
		"is_login": login,
		"user":     user,
		"img":      img,
	})

}

func SearchTiket(w http.ResponseWriter, req *http.Request) {

	session, _ := store.Get(req, "users")

	login := session.Values["Islogin"]

	user := session.Values["name"]
	img := session.Values["img"]

	session.Save(req, w)

	html, err := template.ParseFiles("view/search_tiket.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(w, map[string]interface{}{
		"is_login": login,
		"user":     user,
		"img":      img,
	})
}

func Step_one_book(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	fmt.Println(id)
	html, err := template.ParseFiles("view/booking_plane.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(w, req)
}

func ListTiket(w http.ResponseWriter, req *http.Request) {

	session, _ := store.Get(req, "users")

	login := session.Values["Islogin"]

	user := session.Values["name"]
	img := session.Values["img"]

	session.Save(req, w)

	db := model.Connect()
	// dt := time.Now()
	// date_time := dt.Format("01-02-2006 15:04:05")
	var shedule model.GetAirline
	var datas []model.GetAirline
	rows, err := db.Connection.Query("SELECT flights.id, airlines.airline_name, airlines.airline_image, airports_sources.airport_code , airports_destinations.airport_code ,flights.departure_time, flights.arrival_time, flights.flight_duration,flights.refund,flights.reschedule,flights.price  FROM flights INNER JOIN airlines ON flights.airline_id = airlines.id INNER JOIN airports AS airports_sources ON flights.source_airport_id = airports_sources.id INNER JOIN airports AS airports_destinations ON flights.destination_airport_id = airports_destinations.id WHERE departure_time > NOW();")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err := rows.Scan(&shedule.Id, &shedule.Airline_name, &shedule.Airline_img, &shedule.Source_airport_code, &shedule.Destination_airport_code, &shedule.Depature_time, &shedule.Arrival_time, &shedule.Flight_duration, &shedule.Refund, &shedule.Reschedule, &shedule.Price)
		if err != nil {
			log.Fatal(err)
		}
		departureTime, err := time.Parse("2006-01-02 15:04:05", shedule.Depature_time)
		if err != nil {
			log.Fatal(err)
		}
		departureFormatted := departureTime.Format("15:04")
		shedule.Depature_time = departureFormatted

		// Parsing waktu Arrival
		arrivalTime, err := time.Parse("2006-01-02 15:04:05", shedule.Arrival_time)
		if err != nil {
			log.Fatal(err)
		}
		arrivalFormatted := arrivalTime.Format("15:04")
		shedule.Arrival_time = arrivalFormatted

		datas = append(datas, shedule)
	}

	html, err := template.ParseFiles("view/list_tiket.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(w, map[string]interface{}{
		"datas":    datas,
		"is_login": login,
		"user":     user,
		"img":      img,
	})
}

/*
func Create_db(w http.ResponseWriter, req *http.Request) {
	conn := model.Connect()

	stmt, err := conn.Connection.Prepare("CREATE TABLE Flights (id int NOT NULL AUTO_INCREMENT, airline_id int NOT NULL, source_airport_id int NOT NULL, destination_airport_id int NOT NULL, depature_time DATETIME, arrival_time DATETIME, flight_duration TIME, refund ENUM('yes', 'no'), reschedule ENUM('yes', 'no'), price int(255) NOT NULL, FOREIGN KEY (airline_id) REFERENCES airlines(id) ON DELETE CASCADE,  FOREIGN KEY (source_airport_id) REFERENCES airports(id) ON DELETE CASCADE,  FOREIGN KEY (destination_airport_id) REFERENCES airports(id) ON DELETE CASCADE),PRIMARY KEY(id))")

	if err != nil {
		// Penanganan kesalahan
		fmt.Println(err)
		return
	}

	_, err = stmt.Exec()
	if err != nil {
		// Penanganan kesalahan
		fmt.Println(err)
		return
	}

	if err != nil {
		log.Fatal(err)
	}

	res, err := stmt.Exec()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)

}

func EditCol(w http.ResponseWriter, req *http.Request) {
	conn := model.Connect()
	stmt, err := conn.Connection.Prepare("ALTER TABLE user MODIFY COLUMN id int NOT NULL AUTO_INCREMENT")
	if err != nil {
		log.Fatal()
	}

	res, err := stmt.Exec()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
}
*/

// CREATE TABLE airports (id int NOT NULL AUTO_INCREMENT, airline_id varchar(20) NOT NULL, airport_code varchar(20) NOT NULL, city VARCHAR (220) NOT NULL, country VARCHAR (220) NOT NULL NOT NULL, PRIMARY KEY (id))
