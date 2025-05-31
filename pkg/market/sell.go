package market

import (
	"encoding/json"
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

func (core *Core) CreateSellOrder(appID uint32, contextID, assetID, amount, receivedPrice uint64) (*MarketSellResponse, error) {
	reqUrl := "https://steamcommunity.com/market/sellitem/"
	profileUrl := core.authCore.ProfileUrl()
	if profileUrl == "" {
		return nil, fmt.Errorf("fail to get http header refer")
	}
	reqHeader := http.Header{}
	reqHeader.Add("Content-Type", "application/x-www-form-urlencoded")
	reqHeader.Add("Referer", profileUrl+"inventory/")
	reqBody := url.Values{
		"assetid":   {strconv.FormatUint(assetID, 10)},
		"sessionid": {core.authCore.SessionID()},
		"contextid": {strconv.FormatUint(contextID, 10)},
		"appid":     {strconv.FormatUint(uint64(appID), 10)},
		"amount":    {strconv.FormatUint(amount, 10)},
		"price":     {strconv.FormatUint(receivedPrice, 10)},
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(reqBody.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header = reqHeader

	res, err := core.authCore.HttpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error: %d, %s", res.StatusCode, string(data))
	}

	response := &MarketSellResponse{}
	err = json.Unmarshal(data, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
