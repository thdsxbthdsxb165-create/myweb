package handlers

import (
	"database/sql"
	"html/template"
	"myweb/database"
	"net/http"
	"strings"
)

var profileTmpl = template.Must(template.ParseFiles("templates/profile.html"))

type UserProfile struct {
	Email   string
	Phone   string
	NameTH  string
	NameEN  string
	Contact string
	Role    string
	Balance int
}

func Profile(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session")
	email, _ := session.Values["email"].(string)
	role, _ := session.Values["role"].(string)

	if email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		_, err := database.DB.Exec(`
			UPDATE users
			SET phone=@p1, name_th=@p2, name_en=@p3, contact=@p4
			WHERE email=@p5
		`,
			strings.TrimSpace(r.FormValue("phone")),
			strings.TrimSpace(r.FormValue("name_th")),
			strings.TrimSpace(r.FormValue("name_en")),
			strings.TrimSpace(r.FormValue("contact")),
			email,
		)
		if err != nil {
			http.Error(w, "profile update error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/profile?saved=1", http.StatusSeeOther)
		return
	}

	profile, err := loadProfile(email)
	if err != nil {
		http.Error(w, "profile error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Profile": profile,
		"IsAdmin": role == "admin",
		"Saved":   r.URL.Query().Get("saved") == "1",
	}

	if err := profileTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func loadProfile(email string) (UserProfile, error) {
	var p UserProfile
	var phone, nameTH, nameEN, contact sql.NullString

	err := database.DB.QueryRow(`
		SELECT email, COALESCE(balance, 0), role, phone, name_th, name_en, contact
		FROM users
		WHERE email=@p1
	`, email).Scan(&p.Email, &p.Balance, &p.Role, &phone, &nameTH, &nameEN, &contact)
	if err != nil {
		return p, err
	}

	p.Phone = phone.String
	p.NameTH = nameTH.String
	p.NameEN = nameEN.String
	p.Contact = contact.String

	return p, nil
}
