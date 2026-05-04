package handlers

import (
	"database/sql"
	"html/template"
	"myweb/database"
	"net/http"
	"strings"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.ExecuteTemplate(w, "login.html", nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	var dbEmail, dbPass, role string
	err := database.DB.QueryRow(
		"SELECT email, password, role FROM users WHERE email=@p1",
		email,
	).Scan(&dbEmail, &dbPass, &role)

	if err == sql.ErrNoRows {
		tmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "User not found",
		})
		return
	}

	if err != nil {
		tmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Login failed",
		})
		return
	}

	if dbPass != password {
		tmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Incorrect password",
		})
		return
	}

	role = strings.TrimSpace(strings.ToLower(role))

	session, _ := Store.Get(r, "session")
	session.Values["role"] = role
	session.Values["email"] = dbEmail
	session.Save(r, w)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.ExecuteTemplate(w, "register.html", nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")

	phone := strings.TrimSpace(r.FormValue("phone"))
	nameTH := strings.TrimSpace(r.FormValue("name_th"))
	nameEN := strings.TrimSpace(r.FormValue("name_en"))
	contact := strings.TrimSpace(r.FormValue("contact"))

	if email == "" || password == "" {
		tmpl.ExecuteTemplate(w, "register.html", map[string]string{
			"Error": "Email and password are required",
		})
		return
	}

	if password != confirm {
		tmpl.ExecuteTemplate(w, "register.html", map[string]string{
			"Error": "Passwords do not match",
		})
		return
	}

	if len(password) < 4 {
		tmpl.ExecuteTemplate(w, "register.html", map[string]string{
			"Error": "Password must be at least 4 characters",
		})
		return
	}

	var exists string
	err := database.DB.QueryRow(
		"SELECT email FROM users WHERE email=@p1",
		email,
	).Scan(&exists)

	if err == nil {
		tmpl.ExecuteTemplate(w, "register.html", map[string]string{
			"Error": "Email is already used",
		})
		return
	}

	if err != sql.ErrNoRows {
		tmpl.ExecuteTemplate(w, "register.html", map[string]string{
			"Error": "Could not check email",
		})
		return
	}

	_, err = database.DB.Exec(`
		INSERT INTO users (email, password, phone, name_th, name_en, contact, role, balance)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, 'user', 0)
	`, email, password, phone, nameTH, nameEN, contact)

	if err != nil {
		tmpl.ExecuteTemplate(w, "register.html", map[string]string{
			"Error": "Register failed",
		})
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
