package handlers

import (
	"database/sql"
	"html/template"
	"myweb/database"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var topupTmpl = template.Must(template.ParseFiles("templates/topup.html"))

type TopupView struct {
	ID          int
	Email       string
	Amount      int
	Status      string
	Note        string
	CreatedAt   string
	ProcessedAt string
	ProcessedBy string
}

func Topup(w http.ResponseWriter, r *http.Request) {
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

		amount, err := strconv.Atoi(r.FormValue("amount"))
		if err != nil || amount <= 0 {
			renderTopup(w, email, role, "Please enter a valid amount", "")
			return
		}
		if amount > 100000 {
			renderTopup(w, email, role, "Amount is too high", "")
			return
		}

		_, err = database.DB.Exec(`
			INSERT INTO topup_requests (email, amount, note)
			VALUES (@p1, @p2, @p3)
		`, email, amount, strings.TrimSpace(r.FormValue("note")))
		if err != nil {
			http.Error(w, "topup error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/topup?created=1", http.StatusSeeOther)
		return
	}

	renderTopup(w, email, role, "", r.URL.Query().Get("created"))
}

func renderTopup(w http.ResponseWriter, email, role, errorMsg, created string) {
	profile, _ := loadProfile(email)
	history, err := loadTopupsByEmail(email)
	if err != nil {
		http.Error(w, "topup history error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Profile": profile,
		"History": history,
		"IsAdmin": role == "admin",
		"Error":   errorMsg,
		"Created": created == "1",
	}

	if err := topupTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func loadTopupsByEmail(email string) ([]TopupView, error) {
	rows, err := database.DB.Query(`
		SELECT id, email, amount, status, note, created_at, processed_at, processed_by
		FROM topup_requests
		WHERE email=@p1
		ORDER BY id DESC
	`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTopups(rows)
}

func scanTopups(rows *sql.Rows) ([]TopupView, error) {
	var items []TopupView
	for rows.Next() {
		var item TopupView
		var note, processedBy sql.NullString
		var createdAt time.Time
		var processedAt sql.NullTime

		if err := rows.Scan(
			&item.ID,
			&item.Email,
			&item.Amount,
			&item.Status,
			&note,
			&createdAt,
			&processedAt,
			&processedBy,
		); err != nil {
			return nil, err
		}

		item.Note = note.String
		item.ProcessedBy = processedBy.String
		item.CreatedAt = createdAt.Local().Format("2006-01-02 15:04")
		if processedAt.Valid {
			item.ProcessedAt = processedAt.Time.Local().Format("2006-01-02 15:04")
		}

		items = append(items, item)
	}

	return items, rows.Err()
}
