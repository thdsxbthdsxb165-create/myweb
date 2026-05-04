package handlers

import (
	"database/sql"
	"myweb/database"
	"net/http"
	"strconv"
)

func RentPCV(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session")

	email, ok := session.Values["email"].(string)
	if !ok || email == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	hour, err := strconv.Atoi(r.URL.Query().Get("hour"))
	if err != nil || hour <= 0 {
		hour = 1
	}
	if hour > 5 {
		hour = 5
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "tx error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var price sql.NullInt64
	err = tx.QueryRow("SELECT price FROM pcs WHERE id=@p1", id).Scan(&price)
	if err != nil || !price.Valid {
		http.Error(w, "PC not found", http.StatusNotFound)
		return
	}

	total := int(price.Int64) * hour

	var balance int
	err = tx.QueryRow(`
		SELECT COALESCE(balance, 0) FROM users WITH (UPDLOCK)
		WHERE email=@p1
	`, email).Scan(&balance)
	if err != nil {
		http.Error(w, "user error", http.StatusInternalServerError)
		return
	}

	if balance < total {
		http.Error(w, "Not enough balance", http.StatusBadRequest)
		return
	}

	_, err = tx.Exec(`
		UPDATE users
		SET balance = COALESCE(balance, 0) - @p1
		WHERE email = @p2
	`, total, email)
	if err != nil {
		http.Error(w, "update balance error", http.StatusInternalServerError)
		return
	}

	result, err := tx.Exec(`
		UPDATE pcs
		SET end_time = DATEADD(hour, @p3, GETUTCDATE()),
		    user_email = @p2
		WHERE id = @p1
		AND (end_time IS NULL OR end_time < GETUTCDATE())
	`, id, email, hour)
	if err != nil {
		http.Error(w, "rent error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "PC is busy", http.StatusConflict)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "commit error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("ok"))
}
