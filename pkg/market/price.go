package market

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
	"github.com/umichan0621/steam/pkg/utils"
)

type PriceInfo struct {
	Time  time.Time
	Price float64
	Count int
}

type PriceOverviewInfo struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	MedianPrice string `json:"median_price"`
	Volume      string `json:"volume"`
}

func (core *Core) PriceHistory(appID uint64, hashName string, lastNDays int) ([]*PriceInfo, error) {
	reqBody := url.Values{
		"appid":            {strconv.FormatUint(appID, 10)},
		"market_hash_name": {hashName},
	}
	reqUrl := fmt.Sprintf("https://steamcommunity.com/market/pricehistory/?%s", reqBody.Encode())
	res, err := core.authCore.HttpClient().Get(reqUrl)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fail to get item [%s]'s price history, appID: %d, code: %d", hashName, appID, res.StatusCode)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	jsonData := string(data)

	success := gjson.Get(jsonData, "success").Bool()
	if !success {
		return nil, fmt.Errorf("fail to get item [%s]'s price history, appID: %d", hashName, appID)
	}

	priceInfoList := []*PriceInfo{}
	now := time.Now()
	for _, priceData := range gjson.Get(jsonData, "prices").Array() {
		list := priceData.Array()
		tm, err := utils.ParseSteamTimestamp(list[0].String())
		if err != nil {
			return nil, err
		}
		deltaDay := utils.DeltaDay(tm, now)
		if deltaDay > float64(lastNDays) {
			continue
		}
		count, err := strconv.Atoi(list[2].String())
		if err != nil {
			return nil, err
		}
		price := list[1].Float()
		priceInfoList = append(priceInfoList,
			&PriceInfo{
				Time:  tm,
				Price: price,
				Count: count,
			})
	}
	return priceInfoList, nil
}

func (core *Core) PriceOverview(appID uint64, country, currencyID, marketHashName string) (*PriceOverviewInfo, error) {
	reqBody := url.Values{
		"appid":            {strconv.FormatUint(appID, 10)},
		"country":          {country},
		"currencyID":       {currencyID},
		"market_hash_name": {marketHashName},
	}
	reqUrl := fmt.Sprintf("https://steamcommunity.com/market/priceoverview/?%s", reqBody.Encode())
	res, err := core.authCore.HttpClient().Get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fail to get item [%s]'s price overview, appID: %d, code: %d", marketHashName, appID, res.StatusCode)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	response := &PriceOverviewInfo{}
	err = json.Unmarshal(data, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
