package utils

import (
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
)

var sessionStore = sessions.NewCookieStore([]byte("GOTSessionStore"))
var sessionOptionsSet = false
var globalSession *sessions.Session

func Session(r *http.Request) *sessions.Session {
	if globalSession != nil {
		return globalSession
	}

	if session, err := sessionStore.Get(r, "Session"); err != nil {
		fmt.Printf("Error occured %v\n", err)
		panic(err)
	} else {
		if !sessionOptionsSet {
			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400 * 7,
				HttpOnly: true,
			}
			sessionOptionsSet = true
		}

		globalSession = session // set to global.
		return session
	}
}
