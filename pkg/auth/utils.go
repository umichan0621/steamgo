package auth

import (
	"errors"
	"fmt"
	"net/http"
)

func (core *Core) getProfileUrl() (string, error) {
	// We do not follow redirect, we want to know where it'd redirect us.
	core.httpClient.CheckRedirect =
		func(req *http.Request, via []*http.Request) error {
			return errors.New("do not redirect")
		}

	// Query normal, this will redirect us.
	res, err := core.httpClient.Get("https://steamcommunity.com/my")
	if res == nil {
		return "", err
	}

	res.Body.Close()
	if res.StatusCode != http.StatusFound {
		return "", fmt.Errorf("http error: %d", res.StatusCode)
	}

	// We now have a few useful variables in header, for now, we will just grap "Location".
	return res.Header.Get("Location"), nil
}
