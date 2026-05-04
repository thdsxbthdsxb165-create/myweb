package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"myweb/database"
	"myweb/models"
	"net/http"
	"time"
)

var dashboardTmpl = template.Must(template.ParseFiles("templates/dashboard.html"))

type PCView struct {
	ID        int
	Name      string
	Spec      string
	Price     int
	End       int64
	UserEmail string
	IsBusy    bool
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session")

	role, ok := session.Values["role"].(string)
	email, _ := session.Values["email"].(string)

	if !ok || email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, name, spec, price, end_time, user_email
		FROM pcs
		ORDER BY id
	`)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pcs []PCView
	for rows.Next() {
		var pc models.PC
		var userEmail sql.NullString
		if err := rows.Scan(&pc.ID, &pc.Name, &pc.Spec, &pc.Price, &pc.EndTime, &userEmail); err != nil {
			http.Error(w, "PC data error", http.StatusInternalServerError)
			return
		}

		var end int64
		if pc.EndTime != nil {
			end = pc.EndTime.UTC().Unix()
		}

		now := time.Now().UTC().Unix()
		pcs = append(pcs, PCView{
			ID:        pc.ID,
			Name:      pc.Name,
			Spec:      pc.Spec,
			Price:     pc.Price,
			End:       end,
			UserEmail: userEmail.String,
			IsBusy:    end > now,
		})
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "PC data error", http.StatusInternalServerError)
		return
	}

	var balance int
	if err := database.DB.QueryRow(
		"SELECT COALESCE(balance, 0) FROM users WHERE email=@p1",
		email,
	).Scan(&balance); err != nil {
		log.Println("balance query error:", err)
		http.Error(w, "Balance error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Role":    role,
		"Email":   email,
		"IsAdmin": role == "admin",
		"PCs":     pcs,
		"Balance": balance,
		"Stats":   buildDashboardStats(pcs),
	}

	if err := dashboardTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func buildDashboardStats(pcs []PCView) map[string]int {
	stats := map[string]int{
		"Total":     len(pcs),
		"Available": 0,
		"Busy":      0,
	}

	for _, pc := range pcs {
		if pc.IsBusy {
			stats["Busy"]++
		} else {
			stats["Available"]++
		}
	}

	return stats
}

func ExtendPC(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session")
	email, _ := session.Values["email"].(string)

	if email == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	result, err := database.DB.Exec(`
		UPDATE pcs
		SET end_time = DATEADD(hour, 1, end_time)
		WHERE id = @p1 AND user_email = @p2
	`, id, email)

	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "not owner", http.StatusForbidden)
		return
	}

	w.Write([]byte("ok"))
}
