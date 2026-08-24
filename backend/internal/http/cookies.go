package http

import (
	"net/http"
	"os"
	"time"
)

// cookieSecure reports whether session cookies must carry the Secure flag.
// Enabled in any deployment served over HTTPS (set COOKIE_SECURE=true); left
// off for plain-HTTP local development, where a Secure cookie would be dropped.
func cookieSecure() bool {
	return os.Getenv("COOKIE_SECURE") == "true"
}

// sessionCookie builds a session cookie. The frontend proxies /api to this
// service, so browser and API share an origin and SameSite=Lax suffices.
func sessionCookie(name, value string, exp time.Time) *http.Cookie {
	return &http.Cookie{
		Name: name, Value: value, Path: "/", HttpOnly: true,
		Secure: cookieSecure(), SameSite: http.SameSiteLaxMode, Expires: exp,
	}
}

// clearCookie expires a session cookie. Its attributes must match the ones used
// when setting it, or the browser keeps the original cookie.
func clearCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name: name, Value: "", Path: "/", HttpOnly: true,
		Secure: cookieSecure(), SameSite: http.SameSiteLaxMode, MaxAge: -1,
	}
}
