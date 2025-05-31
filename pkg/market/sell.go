package market

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type MarketSellResponse struct {
	Success                    bool   `json:"success"`
	RequiresConfirmation       uint32 `json:"requires_confirmation"`
	MobileConfirmationRequired bool   `json:"needs_mobile_confirmation"`
	EmailConfirmationRequired  bool   `json:"needs_email_confirmation"`
	EmailDomain                string `json:"email_domain"`
}

func (core *Core) GetProfileURL() (string, error) {
	tmpClient := *core.authCore.HttpClient()
	// tmpClient := http.Client{
	// 	Jar: core.authCore.HttpClient().Jar,
	// }

	/* We do not follow redirect, we want to know where it'd redirect us.  */
	tmpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errors.New("do not redirect")
	}

	/* Query normal, this will redirect us.  */
	resp, err := tmpClient.Get("https://steamcommunity.com/my")
	if resp == nil {
		return "", err
	}

	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("http error: %d", resp.StatusCode)
	}

	/* We now have a few useful variables in header, for now, we will just grap "Location".  */
	return resp.Header.Get("Location"), nil
}

func (core *Core) CreateSellOrder(appID uint32, contextID, assetID, amount, receivedPrice uint64) (*MarketSellResponse, error) {
	body := url.Values{
		"amount":    {strconv.FormatUint(amount, 10)},
		"appid":     {strconv.FormatUint(uint64(appID), 10)},
		"assetid":   {strconv.FormatUint(assetID, 10)},
		"contextid": {strconv.FormatUint(contextID, 10)},
		"price":     {strconv.FormatUint(receivedPrice, 10)},
		"sessionid": {core.authCore.SessionID()},
	}

	req, err := http.NewRequest(http.MethodPost, "https://steamcommunity.com/market/sellitem/", strings.NewReader(body.Encode()))

	fmt.Println(core.authCore.SessionID())
	fmt.Println("-----------------")
	if err != nil {
		return nil, err
	}
	profileURL, err := core.GetProfileURL()
	if err != nil {
		return nil, err
	}
	fmt.Println(profileURL + "inventory/")
	req.Header.Add("Referer", profileURL+"inventory/")

	resp, err := core.authCore.HttpClient().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}
	fmt.Println(core.authCore.SessionID())
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("http error: %d, %s", resp.StatusCode, err.Error())
		}
		return nil, fmt.Errorf("http error: %d, %s", resp.StatusCode, string(data))
	}

	response := &MarketSellResponse{}
	if err = json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}

	return response, nil
}
