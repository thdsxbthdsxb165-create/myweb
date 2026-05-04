package handlers

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte("secret-key"))

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
