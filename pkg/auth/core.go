package auth

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	kURI_STEAM_API       = "https://api.steampowered.com"
	kURI_STEAM_STROE     = "https://store.steampowered.com"
	kURI_STEAM_COMMUNITY = "https://steamcommunity.com"
)

type LoginInfo struct {
	UserName string
	Password string
}

type Core struct {
	loginInfo  LoginInfo
	sessionId  string
	httpClient http.Client
}

func (mgr *Core) Init(info LoginInfo) {
	mgr.loginInfo = info
	mgr.sessionId = ""
	mgr.httpClient = http.Client{}
}

// timeout: millsecond, set only while timeout > 0;
// proxy: if proxyUrl == "", ignore
func (mgr *Core) SetHttpParam(timeout int, proxy string) error {
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
		mgr.httpClient.Timeout = timeoutVal
	}
	mgr.httpClient.Transport = transport
	return nil
}
