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

// Success while Code == 1
type BuyOrderResponse struct {
	Code    int    `json:"success"`
	Msg     string `json:"message"`
	OrderID uint64 `json:"buy_orderid,string"`
}

func (core *Core) PlaceBuyOrder(appID uint64, paymentPrice float64, quantity uint64, currencyID, hashName string) (*BuyOrderResponse, error) {
	reqUrl := "https://steamcommunity.com/market/createbuyorder/"
	reqHeader := http.Header{}
	referer := strings.Replace(hashName, " ", "%20", -1)
	referer = strings.Replace(referer, "#", "%23", -1)
	referer = fmt.Sprintf("https://steamcommunity.com/market/listings/%d/%s", appID, referer)
	reqHeader.Add("Referer", referer)
	reqHeader.Add("Content-Type", "application/x-www-form-urlencoded")
	reqBody := url.Values{
		"appid":            {strconv.FormatUint(appID, 10)},
		"currency":         {currencyID},
		"market_hash_name": {hashName},
		"price_total":      {strconv.FormatUint(uint64(paymentPrice*100), 10)},
		"quantity":         {strconv.FormatUint(quantity, 10)},
		"sessionid":        {core.authCore.SessionID()},
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
	response := &BuyOrderResponse{}
	err = json.Unmarshal(data, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (core *Core) CancelBuyOrder(orderID uint64) error {
	reqUrl := "https://steamcommunity.com/market/cancelbuyorder/"
	reqHeader := http.Header{}
	reqHeader.Add("Referer", "https://steamcommunity.com/market")
	reqHeader.Add("Content-Type", "application/x-www-form-urlencoded")
	reqBody := url.Values{
		"sessionid":   {core.authCore.SessionID()},
		"buy_orderid": {strconv.FormatUint(orderID, 10)},
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(reqBody.Encode()))
	if err != nil {
		return err
	}

	req.Header = reqHeader

	res, err := core.authCore.HttpClient().Do(req)
	if res != nil {
		res.Body.Close()
	}
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("cannot cancel %d: %d", orderID, res.StatusCode)
	}
	return nil
}
