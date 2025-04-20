package market

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"steam/pkg/utils"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
)

type ItemPriceInfo struct {
	Time  time.Time
	Price float64
	Count int
}

func (core *Core) GetItemPriceHistory(appID uint64, marketHashName string, lastNDays int) ([]*ItemPriceInfo, error) {
	val := url.Values{
		"appid":            {strconv.FormatUint(appID, 10)},
		"market_hash_name": {marketHashName},
	}

	res, err := core.httpClient.Get("https://steamcommunity.com/market/pricehistory/?" + val.Encode())
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fail to get item [%s]'s price history, appID: %d, code: %d", marketHashName, appID, res.StatusCode)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	jsonData := string(data)
	success := gjson.Get(jsonData, "success").Bool()
	if !success {
		return nil, fmt.Errorf("fail to get item [%s]'s price history, appID: %d", marketHashName, appID)
	}
	// price_prefix := gjson.Get(jsonData, "price_prefix").String()
	// price_suffix := gjson.Get(jsonData, "price_suffix").String()
	// fmt.Println("price_prefix", price_prefix)
	// fmt.Println("price_suffix", price_suffix)

	resList := []*ItemPriceInfo{}
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
		resList = append(resList,
			&ItemPriceInfo{
				Time:  tm,
				Price: price,
				Count: count,
			})
	}
	return resList, nil
}

func (core *Core) GetItemPriceOverview(appID uint64, country, currencyID, marketHashName string) ([]*ItemPriceInfo, error) {
	val := url.Values{
		"appid":            {strconv.FormatUint(appID, 10)},
		"country":          {country},
		"currencyID":       {currencyID},
		"market_hash_name": {marketHashName},
	}

	res, err := core.httpClient.Get("https://steamcommunity.com/market/priceoverview/?" + val.Encode())
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
	fmt.Println(string(data))
	return nil, nil
}
