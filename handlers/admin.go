package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"myweb/database"
	"myweb/models"
	"net/http"
	"strconv"
	"strings"
)

var adminTmpl = template.Must(template.ParseFiles("templates/admin.html"))

type AdminUser struct {
	Email   string
	Phone   string
	NameTH  string
	NameEN  string
	Contact string
	Role    string
	Balance int
}

type AdminPC struct {
	ID        int
	Name      string
	Spec      string
	Price     int
	UserEmail string
	IsBusy    bool
}

func Admin(w http.ResponseWriter, r *http.Request) {
	email, role, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if role != "admin" {
		http.Error(w, "admin only", http.StatusForbidden)
		return
	}

	users, err := loadAdminUsers()
	if err != nil {
		log.Println("admin users error:", err)
		http.Error(w, "users error", http.StatusInternalServerError)
		return
	}

	pcs, err := loadAdminPCs()
	if err != nil {
		log.Println("admin pcs error:", err)
		http.Error(w, "pcs error", http.StatusInternalServerError)
		return
	}

	topups, err := loadAllTopups()
	if err != nil {
		log.Println("admin topups error:", err)
		http.Error(w, "topups error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Email":   email,
		"Users":   users,
		"PCs":     pcs,
		"Topups":  topups,
		"Message": r.URL.Query().Get("msg"),
	}

	if err := adminTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	_, role, ok := currentUser(r)
	if !ok || role != "admin" {
		http.Error(w, "admin only", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	balance, err := strconv.Atoi(r.FormValue("balance"))
	if err != nil || balance < 0 {
		http.Error(w, "invalid balance", http.StatusBadRequest)
		return
	}

	userRole := strings.TrimSpace(strings.ToLower(r.FormValue("role")))
	if userRole != "admin" {
		userRole = "user"
	}

	_, err = database.DB.Exec(`
		UPDATE users
		SET balance=@p1, role=@p2, phone=@p3, name_th=@p4, name_en=@p5, contact=@p6
		WHERE email=@p7
	`,
		balance,
		userRole,
		strings.TrimSpace(r.FormValue("phone")),
		strings.TrimSpace(r.FormValue("name_th")),
		strings.TrimSpace(r.FormValue("name_en")),
		strings.TrimSpace(r.FormValue("contact")),
		strings.TrimSpace(r.FormValue("email")),
	)
	if err != nil {
		http.Error(w, "update user error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin?msg=user-updated", http.StatusSeeOther)
}

func AdminSavePC(w http.ResponseWriter, r *http.Request) {
	_, role, ok := currentUser(r)
	if !ok || role != "admin" {
		http.Error(w, "admin only", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	price, err := strconv.Atoi(r.FormValue("price"))
	if err != nil || price < 0 {
		http.Error(w, "invalid price", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	spec := strings.TrimSpace(r.FormValue("spec"))
	if name == "" {
		http.Error(w, "missing pc name", http.StatusBadRequest)
		return
	}

	if id > 0 {
		_, err = database.DB.Exec(`
			UPDATE pcs
			SET name=@p1, spec=@p2, price=@p3
			WHERE id=@p4
		`, name, spec, price, id)
	} else {
		_, err = database.DB.Exec(`
			INSERT INTO pcs (id, name, spec, price, end_time, user_email)
			SELECT COALESCE(MAX(id), 0) + 1, @p1, @p2, @p3, NULL, NULL
			FROM pcs WITH (TABLOCKX)
		`, name, spec, price)
	}
	if err != nil {
		log.Println("admin save pc error:", err)
		http.Error(w, "save pc error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin?msg=pc-saved", http.StatusSeeOther)
}

func AdminPCAction(w http.ResponseWriter, r *http.Request) {
	_, role, ok := currentUser(r)
	if !ok || role != "admin" {
		http.Error(w, "admin only", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	switch r.FormValue("action") {
	case "reset":
		_, err := database.DB.Exec(`
			UPDATE pcs
			SET end_time=NULL, user_email=NULL
			WHERE id=@p1
		`, id)
		if err != nil {
			http.Error(w, "reset pc error", http.StatusInternalServerError)
			return
		}
	case "delete":
		_, err := database.DB.Exec("DELETE FROM pcs WHERE id=@p1", id)
		if err != nil {
			http.Error(w, "delete pc error", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin?msg=pc-updated", http.StatusSeeOther)
}

func AdminTopupAction(w http.ResponseWriter, r *http.Request) {
	adminEmail, role, ok := currentUser(r)
	if !ok || role != "admin" {
		http.Error(w, "admin only", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	action := r.FormValue("action")
	if action != "approve" && action != "reject" {
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "tx error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var userEmail string
	var amount int
	var status string
	err = tx.QueryRow(`
		SELECT email, amount, status
		FROM topup_requests WITH (UPDLOCK)
		WHERE id=@p1
	`, id).Scan(&userEmail, &amount, &status)
	if err != nil {
		http.Error(w, "topup not found", http.StatusNotFound)
		return
	}
	if status != "pending" {
		http.Error(w, "topup already processed", http.StatusConflict)
		return
	}

	nextStatus := "rejected"
	if action == "approve" {
		nextStatus = "approved"
		if _, err := tx.Exec(`
			UPDATE users
			SET balance = COALESCE(balance, 0) + @p1
			WHERE email=@p2
		`, amount, userEmail); err != nil {
			http.Error(w, "update balance error", http.StatusInternalServerError)
			return
		}
	}

	_, err = tx.Exec(`
		UPDATE topup_requests
		SET status=@p1, processed_at=SYSUTCDATETIME(), processed_by=@p2
		WHERE id=@p3
	`, nextStatus, adminEmail, id)
	if err != nil {
		http.Error(w, "update topup error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "commit error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin?msg=topup-"+nextStatus, http.StatusSeeOther)
}

func currentUser(r *http.Request) (string, string, bool) {
	session, _ := Store.Get(r, "session")
	email, _ := session.Values["email"].(string)
	role, _ := session.Values["role"].(string)

	return email, role, email != ""
}

func loadAdminUsers() ([]AdminUser, error) {
	rows, err := database.DB.Query(`
		SELECT
			COALESCE(email, ''),
			COALESCE(balance, 0),
			COALESCE(role, 'user'),
			phone,
			name_th,
			name_en,
			contact
		FROM users
		ORDER BY email
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		var phone, nameTH, nameEN, contact sql.NullString
		if err := rows.Scan(&u.Email, &u.Balance, &u.Role, &phone, &nameTH, &nameEN, &contact); err != nil {
			return nil, err
		}
		u.Phone = phone.String
		u.NameTH = nameTH.String
		u.NameEN = nameEN.String
		u.Contact = contact.String
		users = append(users, u)
	}

	return users, rows.Err()
}

func loadAdminPCs() ([]AdminPC, error) {
	rows, err := database.DB.Query(`
		SELECT id, name, spec, price, user_email,
		       CASE WHEN end_time IS NOT NULL AND end_time > GETUTCDATE() THEN 1 ELSE 0 END
		FROM pcs
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pcs []AdminPC
	for rows.Next() {
		var pc models.PC
		var userEmail sql.NullString
		var isBusy int
		if err := rows.Scan(&pc.ID, &pc.Name, &pc.Spec, &pc.Price, &userEmail, &isBusy); err != nil {
			return nil, err
		}
		pcs = append(pcs, AdminPC{
			ID:        pc.ID,
			Name:      pc.Name,
			Spec:      pc.Spec,
			Price:     pc.Price,
			UserEmail: userEmail.String,
			IsBusy:    isBusy == 1,
		})
	}

	return pcs, rows.Err()
}

func loadAllTopups() ([]TopupView, error) {
	rows, err := database.DB.Query(`
		SELECT id, email, amount, status, note, created_at, processed_at, processed_by
		FROM topup_requests
		ORDER BY
			CASE WHEN status='pending' THEN 0 ELSE 1 END,
			id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTopups(rows)
}
