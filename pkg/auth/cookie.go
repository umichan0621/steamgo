package auth

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type CookieData struct {
	SessionID        string
	SteamLoginSecure string
	SteamID          string
	Expires          int64
	MaxAge           int
}

func (core *Core) SetCookie(cookieData CookieData) { core.cookieData = cookieData }

func (core *Core) ApplyCookie() {
	cookieList := []*http.Cookie{}
	cookie1 := http.Cookie{
		Name:     "sessionid",
		Value:    core.cookieData.SessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	cookie2 := http.Cookie{
		Name:     "steamLoginSecure",
		Value:    core.cookieData.SteamLoginSecure,
		Path:     "/",
		Expires:  time.Unix(core.cookieData.Expires, 0),
		MaxAge:   core.cookieData.MaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	cookieList = append(cookieList, &cookie1)
	cookieList = append(cookieList, &cookie2)
	cookieList = append(cookieList, &http.Cookie{Name: "mobileClientVersion", Value: "0 (2.1.3)"})
	cookieList = append(cookieList, &http.Cookie{Name: "mobileClient", Value: "android"})
	cookieList = append(cookieList, &http.Cookie{Name: "steamid", Value: core.cookieData.SteamID})
	cookieList = append(cookieList, &http.Cookie{Name: "Steam_Language", Value: "english"})
	cookieList = append(cookieList, &http.Cookie{Name: "dob", Value: ""})
	jar, _ := cookiejar.New(nil)

	jar.SetCookies(
		&url.URL{
			Scheme: "https",
			Host:   "steamcommunity.com",
		},
		cookieList,
	)
	core.httpClient.Jar = jar
}
