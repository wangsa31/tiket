package middelware

import (
	"net/http"

	controller "github.com/GoWeb/tiket/controller"
)

func AuthLogin(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		session, _ := controller.Store.Get(req, "users")
		if session.Values["Islogin"] != true {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func AuthPay(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		session, _ := controller.Store.Get(req, "booking")
		if session.Values["IsPay"] != true {
			http.Redirect(w, req, "/alfamart-pay", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, req)
	})
}
