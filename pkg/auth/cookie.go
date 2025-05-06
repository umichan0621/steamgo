package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"github.com/tidwall/gjson"
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

func (core *Core) SaveCookie(cookiePath string) error {
	cookieData, _ := json.Marshal(core.cookieData)
	file, err := os.OpenFile(cookiePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(cookieData)
	if err != nil {
		return err
	}

	return nil
}

func (core *Core) LoadCookie(cookiePath string) error {
	file, err := os.Open(cookiePath)
	if err != nil {
		return err
	}

	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	jsonData := string(data)
	cookieData := CookieData{
		SessionID:        gjson.Get(jsonData, "SessionID").String(),
		SteamLoginSecure: gjson.Get(jsonData, "SteamLoginSecure").String(),
		SteamID:          gjson.Get(jsonData, "SteamID").String(),
		Expires:          gjson.Get(jsonData, "Expires").Int(),
		MaxAge:           int(gjson.Get(jsonData, "MaxAge").Int()),
	}
	core.SetCookie(cookieData)
	return nil
}
