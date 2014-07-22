package sessions

import (
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/utils"
	"github.com/gorilla/sessions"
	"net/http"
)

// The global session store
var (
	// used as session store. never stored.
	longliveCookieStore *sessions.CookieStore
	ephemeron           *sessions.CookieStore
	// globalSession     *sessions.Session // TODO: change this to application scope.
)

func init() {
	// Init session store
	longliveCookieStore = sessions.NewCookieStore([]byte("GOTLongLiveSessionStore"))
	longliveCookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	ephemeron = sessions.NewCookieStore([]byte("GOTShortLiveSessionStore"))
	ephemeron.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   0, // 0 means delete when browser is closed.
		HttpOnly: true,
	}
}

// LongSession returns Long-lived Cookie based Session
func LongCookieSession(r *http.Request) *sessions.Session {
	if session, err := longliveCookieStore.Get(r, "session"); err != nil {
		fmt.Printf("Error occured %v\n", err)
		panic(err)
	} else {
		return session
	}
}

// ShortSession returns Short-lived Cookie based Session
func ShortCookieSession(r *http.Request) *sessions.Session {
	if session, err := ephemeron.Get(r, "ephemeron"); err != nil {
		fmt.Printf("Error occured %v\n", err)
		panic(err)
	} else {
		return session
	}
}

func SessionId(r *http.Request, w http.ResponseWriter) string {
	// TODO: Should I cache session id in request?
	session := ShortCookieSession(r)
	if jsessionId, ok := session.Values[config.SESSIONID_KEY]; !ok || jsessionId == "" {
		// create session id
		jsessionId = utils.GenGuid()
		session.Values[config.SESSIONID_KEY] = jsessionId
		session.Save(r, w)
	}
	return ""
}

func GetSessionId(r *http.Request) string {
	session := ShortCookieSession(r)
	if jsessionId, ok := session.Values[config.SESSIONID_KEY]; ok {
		return jsessionId.(string)
	}
	return ""
}
