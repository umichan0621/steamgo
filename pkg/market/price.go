package market

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

type OrderInfo struct {
	Price    float64
	Quantity int32
}

type OrderGraph struct {
	BuyOrderGraph  []OrderInfo
	SellOrderGraph []OrderInfo
}

func (core *Core) getItemNameID(appID uint64, hashName string) (string, error) {
	reqUrl := fmt.Sprintf("https://steamcommunity.com/market/listings/%d/%s", appID, url.PathEscape(hashName))
	res, err := core.authCore.HttpClient().Get(reqUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fail to get price, hash name: %s, code: %d", hashName, res.StatusCode)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	htmlString := string(data)
	index := strings.Index(htmlString, "Market_LoadOrderSpread")
	if index >= 0 {
		htmlString = htmlString[index:]
		front := strings.Index(htmlString, "(")
		rear := strings.Index(htmlString, ")")
		if front >= 0 && rear >= 0 {
			itemNameID := strings.ReplaceAll(htmlString[front+1:rear], " ", "")
			return itemNameID, nil
		}
	}
	return "", fmt.Errorf("fail to get item name ID")
}

func (core *Core) ItemOrderGraph(appID uint64, hashName string) (*OrderGraph, error) {
	itemNameID, err := core.getItemNameID(appID, hashName)
	if err != nil {
		return nil, err
	}
	reqBody := url.Values{
		"item_nameid": {itemNameID},
		"language":    {core.language},
		"country":     {core.country},
		"currency":    {core.currency},
	}
	reqUrl := "https://steamcommunity.com/market/itemordershistogram?" + reqBody.Encode()
	res, err := core.authCore.HttpClient().Get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	jsonData := string(data)
	success := gjson.Get(jsonData, "success").Int()
	if success != 1 {
		return nil, fmt.Errorf("fail to get order graph")
	}
	orderGraph := &OrderGraph{}
	for _, buyOrders := range gjson.Get(jsonData, "buy_order_graph").Array() {
		buyOrdersInfo := buyOrders.Array()
		if len(buyOrdersInfo) == 3 {
			orderGraph.BuyOrderGraph = append(orderGraph.BuyOrderGraph,
				OrderInfo{
					Price:    buyOrdersInfo[0].Float(),
					Quantity: int32(buyOrdersInfo[1].Int()),
				})
		}
	}

	for _, sellOrders := range gjson.Get(jsonData, "sell_order_graph").Array() {
		sellOrdersInfo := sellOrders.Array()
		if len(sellOrdersInfo) == 3 {
			orderGraph.SellOrderGraph = append(orderGraph.SellOrderGraph,
				OrderInfo{
					Price:    sellOrdersInfo[0].Float(),
					Quantity: int32(sellOrdersInfo[1].Int()),
				})
		}
	}
	return orderGraph, nil
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
