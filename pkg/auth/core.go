package auth

import (
	"net"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	kURI_STEAM_API       = "https://api.steampowered.com"
	kURI_STEAM_STROE     = "https://store.steampowered.com"
	kURI_STEAM_COMMUNITY = "https://steamcommunity.com"
	kURI_STEAM_SETTOKEN  = "https://steamcommunity.com/login/settoken"
)

type LoginInfo struct {
	UserName string
	Password string
}

type Core struct {
	httpClient *http.Client
	loginInfo  LoginInfo
	cookieData CookieData
	profileUrl string
}

func (core *Core) Init(httpClient *http.Client, info LoginInfo) {
	core.loginInfo = info
	core.httpClient = httpClient
	core.profileUrl = ""
}

func (core *Core) HttpClient() *http.Client { return core.httpClient }
func (core *Core) SteamID() string          { return core.cookieData.SteamID }
func (core *Core) SessionID() string        { return core.cookieData.SessionID }

// timeout: millsecond, set only while timeout > 0;
// proxy: if proxyUrl == "", ignore
func (core *Core) SetHttpParam(timeout int, proxy string) error {
	transport := &http.Transport{}
	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return err
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}
	if timeout > 0 {
		timeoutVal := time.Duration(timeout) * time.Millisecond
		dialer := net.Dialer{Timeout: timeoutVal}

		transport.DialContext = dialer.DialContext
		transport.TLSHandshakeTimeout = timeoutVal
		transport.ResponseHeaderTimeout = timeoutVal
		transport.ExpectContinueTimeout = timeoutVal
		core.httpClient.Timeout = timeoutVal
	}
	core.httpClient.Transport = transport
	return nil
}

func (core *Core) ProfileUrl() string {
	if core.profileUrl != "" {
		return core.profileUrl
	}
	res, err := core.getProfileUrl()
	if err != nil {
		log.Errorf("Fail to get profile URL, error: %s", err.Error())
		return ""
	}
	return res
}
