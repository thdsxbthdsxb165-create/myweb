package routes

import (
	"myweb/handlers"
	"net/http"
)

func SetupRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	http.HandleFunc("/login", handlers.Login)
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/dashboard", handlers.Dashboard)
	http.HandleFunc("/profile", handlers.Profile)
	http.HandleFunc("/topup", handlers.Topup)
	http.HandleFunc("/rent", handlers.RentPCV)
	http.HandleFunc("/extend", handlers.ExtendPC)
	http.HandleFunc("/admin", handlers.Admin)
	http.HandleFunc("/admin/user/update", handlers.AdminUpdateUser)
	http.HandleFunc("/admin/pc/save", handlers.AdminSavePC)
	http.HandleFunc("/admin/pc/action", handlers.AdminPCAction)
	http.HandleFunc("/admin/topup/action", handlers.AdminTopupAction)
	http.HandleFunc("/logout", handlers.Logout)
}
