package main

import (
	"log"
	"net/http"

	"github.com/GoWeb/tiket/controller"
	"github.com/GoWeb/tiket/middelware"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := http.NewServeMux()
	// handle public assets
	r.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./public/assets/"))))

	// route
	r.HandleFunc("/tiket", controller.Dashboard)
	r.HandleFunc("/tiket/booking/order", middelware.AuthLogin(controller.BookingList))
	r.HandleFunc("/search-tiket", controller.SearchTiket)
	r.HandleFunc("/list-tiket", controller.ListTiket)
	r.HandleFunc("/login", controller.Login)
	r.HandleFunc("/login_with_google", controller.LoginWithGoogle)
	r.HandleFunc("/auth/google/callback", controller.Callback)
	r.HandleFunc("/register", controller.Register)
	r.HandleFunc("/booking-step-1", middelware.AuthLogin(controller.Step_one_book))
	r.HandleFunc("/payment-methode", controller.Payment_methode)
	r.HandleFunc("/alfamart-pay", controller.Alfamart_pay)
	r.HandleFunc("/pay/finish", middelware.AuthPay(controller.Finish))
	// r.HandleFunc("/createdb", controller.Create_db)
	// r.HandleFunc("/createdb", utils.LoginWithGoogle)

	x := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}

	log.Fatal(x.ListenAndServe())
}
