package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/GoWeb/tiket/model"
	"github.com/GoWeb/tiket/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/thanhpk/randstr"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

var Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_STROE")))

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

		session, _ := Store.Get(req, "users")
		session.Values["Islogin"] = true
		session.Values["name"] = data.Fullname
		session.Values["img"] = data.Picture
		session.Values["email"] = data.Email
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

		session, _ := Store.Get(req, "users")
		session.Values["Islogin"] = true
		session.Values["name"] = "ok"
		session.Values["img"] = data.Picture
		session.Values["email"] = data.Email
		session.Options.MaxAge = 24 * 60 * 60
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

	session, _ := Store.Get(req, "users")
	login := session.Values["Islogin"]
	user := session.Values["name"]
	img := session.Values["img"]
	email := session.Values["email"]
	session.Save(req, w)

	db := model.Connect()
	var id_user int
	db.Connection.QueryRow("SELECT id FROM user where email= ?", email).Scan(&id_user)
	fmt.Println(email)

	session, _ = Store.Get(req, "users")
	session.Values["id_user"] = id_user
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

func BookingList(w http.ResponseWriter, req *http.Request) {
	session, _ := Store.Get(req, "users")
	login := session.Values["Islogin"]
	// id_user := session.Values["id_user"]
	img := session.Values["img"]
	session.Save(req, w)
	db := model.Connect()

	var BookingList model.GetBooking
	var datas []model.GetBooking

	rows, _ := db.Connection.Query("SELECT booking.id, airlines.airline_image, airlines.airline_name, source_airport.airport_name, destination_airport.airport_name, flights.departure_date, flights.departure_time FROM booking INNER JOIN flights ON booking.flight_id = flights.id INNER JOIN airlines ON flights.airline_id = airlines.id INNER JOIN airports AS source_airport ON flights.source_airport_id = source_airport.id INNER JOIN airports AS destination_airport ON flights.destination_airport_id = destination_airport.id WHERE booking.user_id=11; ")

	for rows.Next() {
		err := rows.Scan(&BookingList.Id, &BookingList.Airline_img, &BookingList.Airline_name, &BookingList.Source_airport_name, &BookingList.Destination_airport_name, &BookingList.Depature_date, &BookingList.Depature_time)
		if err != nil {
			log.Fatal(err)
		}
		departureTime, err := time.Parse("15:04:05", BookingList.Depature_time)
		if err != nil {
			log.Fatal(err)
		}
		departureFormatted := departureTime.Format("15:04")
		BookingList.Depature_time = departureFormatted

		// Parsing waktu Arrival

		datas = append(datas, BookingList)
	}

	html, err := template.ParseFiles("view/book_list.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(w, map[string]interface{}{
		"data":     "ok",
		"is_login": login,
		"img":      img,
		"datas":    datas,
	})
}

func SearchTiket(w http.ResponseWriter, req *http.Request) {

	session, _ := Store.Get(req, "users")
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

	session, _ := Store.Get(req, "users")
	user := session.Values["name"]
	session.Save(req, w)

	id, _ := strconv.Atoi(req.URL.Query().Get("id"))
	db := model.Connect()
	// dt := time.Now()
	// date_time := dt.Format("01-02-2006 15:04:05")
	var shedule model.GetAirline
	// var datas []model.GetAirline

	db.Connection.QueryRow("SELECT flights.id, airlines.airline_name, airlines.airline_image, airports_sources.airport_code,airports_sources.city , airports_destinations.airport_code,airports_destinations.city ,flights.departure_date, flights.arrival_date, flights.departure_time, flights.arrival_time, flights.flight_duration,flights.refund,flights.reschedule,flights.price  FROM flights INNER JOIN airlines ON flights.airline_id = airlines.id INNER JOIN airports AS airports_sources ON flights.source_airport_id = airports_sources.id INNER JOIN airports AS airports_destinations ON flights.destination_airport_id = airports_destinations.id WHERE flights.id = ?;", id).Scan(&shedule.Id, &shedule.Airline_name, &shedule.Airline_img, &shedule.Source_airport_code, &shedule.Source_airport_city, &shedule.Destination_airport_code, &shedule.Destination_airport_city, &shedule.Depature_date, &shedule.Arrival_date, &shedule.Depature_time, &shedule.Arrival_time, &shedule.Flight_duration, &shedule.Refund, &shedule.Reschedule, &shedule.Price)

	// parse date
	// BUG
	// dates := utils.Format_time(shedule.Depature_date)
	// shedule.Depature_date = dates
	// END
	// Parsing time

	timeparse, _ := time.Parse("2006-01-02", shedule.Depature_date)

	date := timeparse.Format("2 jan 06")

	shedule.Depature_date = date

	departureTime, err := time.Parse("15:04:05", shedule.Depature_time)
	if err != nil {
		log.Fatal(err)
	}
	departureFormatted := departureTime.Format("15:04")
	shedule.Depature_time = departureFormatted

	arrivalTime, err := time.Parse("15:04:05", shedule.Arrival_time)
	if err != nil {
		log.Fatal(err)
	}
	arrivalFormatted := arrivalTime.Format("15:04")
	shedule.Arrival_time = arrivalFormatted

	times_convert_duration := utils.Format_time("03:30")
	shedule.Flight_duration = times_convert_duration

	html, err := template.ParseFiles("view/booking_plane.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(w, map[string]interface{}{
		"data": shedule,
		"user": user,
	})

}

func ListTiket(w http.ResponseWriter, req *http.Request) {

	session, _ := Store.Get(req, "users")
	login := session.Values["Islogin"]
	user := session.Values["name"]
	img := session.Values["img"]
	session.Options.MaxAge = 24 * 60 * 60
	session.Save(req, w)

	pram1, _ := strconv.Atoi(req.FormValue("depature"))
	pram2, _ := strconv.Atoi(req.FormValue("arrive"))
	pram3 := req.FormValue("date_depature")

	db := model.Connect()
	// dt := time.Now()
	// date_time := dt.Format("01-02-2006 15:04:05")
	var shedule model.GetAirline
	var datas []model.GetAirline
	rows, err := db.Connection.Query("SELECT flights.id, airlines.airline_name, airlines.airline_image, airports_sources.airport_code ,airports_sources.city, airports_destinations.airport_code ,airports_destinations.city,flights.departure_date, flights.arrival_date, flights.departure_time, flights.arrival_time, flights.flight_duration,flights.refund,flights.reschedule,flights.price FROM flights INNER JOIN airlines ON flights.airline_id = airlines.id INNER JOIN airports AS airports_sources ON flights.source_airport_id = airports_sources.id INNER JOIN airports AS airports_destinations ON flights.destination_airport_id = airports_destinations.id WHERE flights.source_airport_id = ? && flights.destination_airport_id = ? && flights.departure_date = ? || flights.departure_time > NOW();", pram1, pram2, pram3)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err := rows.Scan(&shedule.Id, &shedule.Airline_name, &shedule.Airline_img, &shedule.Source_airport_code, &shedule.Destination_airport_city, &shedule.Destination_airport_city, &shedule.Destination_airport_city, &shedule.Depature_date, &shedule.Arrival_date, &shedule.Depature_time, &shedule.Arrival_time, &shedule.Flight_duration, &shedule.Refund, &shedule.Reschedule, &shedule.Price)
		if err != nil {
			log.Fatal(err)
		}
		departureTime, err := time.Parse("15:04:05", shedule.Depature_time)
		if err != nil {
			log.Fatal(err)
		}
		departureFormatted := departureTime.Format("15:04")
		shedule.Depature_time = departureFormatted

		// Parsing waktu Arrival
		arrivalTime, err := time.Parse("15:04:05", shedule.Arrival_time)
		if err != nil {
			log.Fatal(err)
		}
		arrivalFormatted := arrivalTime.Format("15:04")
		shedule.Arrival_time = arrivalFormatted

		times_convert_duration := utils.Format_time("03:30")
		shedule.Flight_duration = times_convert_duration

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

func Payment_methode(w http.ResponseWriter, req *http.Request) {
	// get form
	flight_id, _ := strconv.Atoi(req.FormValue("flight_id"))
	title_book := req.FormValue("title_book")
	name_book := req.FormValue("input1")
	email_book := req.FormValue("input2")
	code_country := req.FormValue("code_country")
	phone := req.FormValue("input3")
	title_passangers := req.FormValue("title_passengers")
	name_passangers := req.FormValue("input5")

	db := model.Connect()
	// dt := time.Now()
	// date_time := dt.Format("01-02-2006 15:04:05")
	var shedule model.GetAirline

	db.Connection.QueryRow("SELECT flights.id, airlines.airline_name, airlines.airline_image, airports_sources.airport_code,airports_sources.city , airports_destinations.airport_code,airports_destinations.city ,flights.departure_date, flights.arrival_date, flights.departure_time, flights.arrival_time, flights.flight_duration,flights.refund,flights.reschedule,flights.price  FROM flights INNER JOIN airlines ON flights.airline_id = airlines.id INNER JOIN airports AS airports_sources ON flights.source_airport_id = airports_sources.id INNER JOIN airports AS airports_destinations ON flights.destination_airport_id = airports_destinations.id WHERE flights.id = ?;", flight_id).Scan(&shedule.Id, &shedule.Airline_name, &shedule.Airline_img, &shedule.Source_airport_code, &shedule.Source_airport_city, &shedule.Destination_airport_code, &shedule.Destination_airport_city, &shedule.Depature_date, &shedule.Arrival_date, &shedule.Depature_time, &shedule.Arrival_time, &shedule.Flight_duration, &shedule.Refund, &shedule.Reschedule, &shedule.Price)

	timeparse, _ := time.Parse("2006-01-02", shedule.Depature_date)

	date := timeparse.Format("2 jan 06")

	shedule.Depature_date = date

	departureTime, err := time.Parse("15:04:05", shedule.Depature_time)
	if err != nil {
		log.Fatal(err)
	}
	departureFormatted := departureTime.Format("15:04")
	shedule.Depature_time = departureFormatted

	arrivalTime, err := time.Parse("15:04:05", shedule.Arrival_time)
	if err != nil {
		log.Fatal(err)
	}
	arrivalFormatted := arrivalTime.Format("15:04")
	shedule.Arrival_time = arrivalFormatted

	times_convert_duration := utils.Format_time("03:30")
	shedule.Flight_duration = times_convert_duration

	data := struct {
		Flight_id        int
		title_book       string
		name_book        string
		email_book       string
		code_country     string
		phone            string
		title_passangers string
		name_passangers  string
	}{
		Flight_id:        flight_id,
		title_book:       title_book,
		name_book:        name_book,
		email_book:       email_book,
		code_country:     code_country,
		phone:            phone,
		title_passangers: title_passangers,
		name_passangers:  name_passangers,
	}
	session, _ := Store.Get(req, "booking")
	session.Values["id"] = data.Flight_id
	session.Values["title_book"] = data.title_book
	session.Values["name_book"] = data.name_book
	session.Values["email"] = data.email_book
	session.Values["code_country"] = data.code_country
	session.Values["phone"] = data.phone
	session.Values["name_passangers"] = data.name_passangers
	session.Values["title_passangers"] = data.title_passangers
	session.Options.MaxAge = 24 * 60 * 60
	session.Save(req, w)

	fmt.Println(data)
	html, err := template.ParseFiles("view/methode_pay.html")

	if err != nil {
		log.Println(err)
	}

	html.Execute(w, map[string]interface{}{
		"data":     data,
		"shedules": shedule,
		"order":    randstr.String(16),
	})
}

func Alfamart_pay(w http.ResponseWriter, req *http.Request) {

	flight_id, _ := strconv.Atoi(req.URL.Query().Get("flight_id"))
	order := req.URL.Query().Get("order")

	var shedule model.GetAirline
	db := model.Connect()
	db.Connection.QueryRow("SELECT flights.id, airlines.airline_name, airlines.airline_image, airports_sources.airport_code,airports_sources.city , airports_destinations.airport_code,airports_destinations.city ,flights.departure_date, flights.arrival_date, flights.departure_time, flights.arrival_time, flights.flight_duration,flights.refund,flights.reschedule,flights.price  FROM flights INNER JOIN airlines ON flights.airline_id = airlines.id INNER JOIN airports AS airports_sources ON flights.source_airport_id = airports_sources.id INNER JOIN airports AS airports_destinations ON flights.destination_airport_id = airports_destinations.id WHERE flights.id = ?;", flight_id).Scan(&shedule.Id, &shedule.Airline_name, &shedule.Airline_img, &shedule.Source_airport_code, &shedule.Source_airport_city, &shedule.Destination_airport_code, &shedule.Destination_airport_city, &shedule.Depature_date, &shedule.Arrival_date, &shedule.Depature_time, &shedule.Arrival_time, &shedule.Flight_duration, &shedule.Refund, &shedule.Reschedule, &shedule.Price)

	timeparse, _ := time.Parse("2006-01-02", shedule.Depature_date)

	date := timeparse.Format("2 jan 06")

	shedule.Depature_date = date

	departureTime, err := time.Parse("15:04:05", shedule.Depature_time)
	if err != nil {
		log.Fatal(err)
	}
	departureFormatted := departureTime.Format("15:04")
	shedule.Depature_time = departureFormatted

	arrivalTime, err := time.Parse("15:04:05", shedule.Arrival_time)
	if err != nil {
		log.Fatal(err)
	}
	arrivalFormatted := arrivalTime.Format("15:04")
	shedule.Arrival_time = arrivalFormatted

	times_convert_duration := utils.Format_time("03:30")
	shedule.Flight_duration = times_convert_duration

	session, _ := Store.Get(req, "booking")
	session.Values["IsPay"] = true
	name_book := session.Values["name_book"]
	email := session.Values["email"]
	title_book := session.Values["title_book"]
	// code_country := session.Values["code_country"]
	phone := session.Values["phone"]

	session.Save(req, w)

	session, _ = Store.Get(req, "users")
	id_user := session.Values["id_user"]
	session.Save(req, w)

	if err != nil {
		log.Println(err)
	}

	// 1. Initiate coreapi client
	c := coreapi.Client{}
	c.New("SB-Mid-server-_ayUoMyGoUamrWSU_LxwSDyf", midtrans.Sandbox)

	var AlfaRess utils.AlfaRess
	check, _ := c.CheckTransaction(order)

	if check.OrderID != order {
		// 2. Initiate charge request
		chargeReq := &coreapi.ChargeReq{

			PaymentType: coreapi.PaymentTypeConvenienceStore,
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  order,
				GrossAmt: int64(shedule.Price),
			},
			ConvStore: &coreapi.ConvStoreDetails{
				Store: "alfamart",
			},
		}

		// 3. Request to Midtrans
		coreApiRes, _ := c.ChargeTransaction(chargeReq)

		AlfaRess = utils.AlfaRess{
			Status_code:    coreApiRes.StatusCode,
			Transaction_id: coreApiRes.TransactionID,
			Order_id:       coreApiRes.OrderID,
			Gross_amount:   coreApiRes.GrossAmount,
			Payment_code:   coreApiRes.PaymentCode,
		}

		var book_id int
		db := model.Connect()
		db.Connection.Exec("INSERT INTO booking VALUES (?,?,?,?,?,?,?,?)", "", id_user, name_book, flight_id, title_book, email, phone, coreApiRes.TransactionStatus)
		db.Connection.QueryRow("SELECT id FROM booking WHERE user_id = ? ORDER BY id DESC LIMIT 1;", id_user).Scan(&book_id)
		db.Connection.Exec("INSERT INTO transactions VALUES(?,?,?,?,?)", order, book_id, time.Now(), coreApiRes.Store, coreApiRes.TransactionStatus)

		session, _ = Store.Get(req, "booking")
		session.Values["book_id"] = book_id
		session.Save(req, w)
	} else {
		http.Redirect(w, req, "http://localhost:3000/pay/finish", http.StatusTemporaryRedirect)
	}

	html, err := template.ParseFiles("view/alfamart_pay.html")

	if err != nil {
		log.Println(err)
	}

	html.Execute(w, map[string]interface{}{
		"schedules": shedule,
		"order":     order,
		"ress_pay":  AlfaRess,
	})
}

func Finish(w http.ResponseWriter, req *http.Request) {
	session, _ := Store.Get(req, "booking")
	book_id := session.Values["book_id"]
	name_passengers := session.Values["name_passangers"]
	title_passengers := session.Values["title_passangers"]
	session.Save(req, w)

	session, _ = Store.Get(req, "users")
	id_user := session.Values["id_user"]
	session.Save(req, w)

	db := model.Connect()
	fmt.Println(book_id, " ", name_passengers, " ", title_passengers)
	_, err := db.Connection.Exec("INSERT INTO passengers VALUES (?,?,?,?,?)", "", book_id, name_passengers, title_passengers)
	_, err = db.Connection.Exec("UPDATE booking SET status = 'confirmed' WHERE user_id = ? ORDER BY id DESC LIMIT 1", id_user)
	_, err = db.Connection.Exec("UPDATE transactions SET status = 'settlement' WHERE  booking_id= ?", book_id)

	if err != nil {
		log.Println(err)
	}

	if err != nil {
		log.Println(err)
	}
	session, _ = Store.Get(req, "booking") // Ganti "session-name" dengan nama sesi Anda

	// Menghapus session
	session.Options.MaxAge = -1
	session.Save(req, w)
	w.Write([]byte("the tiket has pay"))
}
