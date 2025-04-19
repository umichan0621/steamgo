package market

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Core struct {
	httpClient *http.Client
}

func (core *Core) Init(httpClient *http.Client) {
	core.httpClient = httpClient
}

func (core *Core) GetItemPriceHistory(appID uint64, marketHashName string) error {
	u := url.URL{
		Scheme: "https",
		Host:   "steamcommunity.com",
	}
	fmt.Println(core.httpClient.Jar.Cookies(&u))
	val := url.Values{
		"appid":            {strconv.FormatUint(appID, 10)},
		"market_hash_name": {marketHashName},
	}

	res, err := core.httpClient.Get("https://steamcommunity.com/market/pricehistory/?" + val.Encode())
	if err != nil {
		return err
	}

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Println("header =", res.Header)
	fmt.Println("code =", res.StatusCode)
	fmt.Println(string(data))
	// if resp.StatusCode != http.StatusOK {
	// 	return nil, fmt.Errorf("http error: %d", resp.StatusCode)
	// }

	// response := MarketItemResponse{}
	// if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
	// 	return nil, err
	// }

	// if !response.Success {
	// 	return nil, ErrCannotLoadPrices
	// }

	// var prices []interface{}
	// var ok bool
	// if prices, ok = response.Prices.([]interface{}); !ok {
	// 	return nil, ErrCannotLoadPrices
	// }

	// items := []*MarketItemPrice{}
	// for _, v := range prices {
	// 	if v, ok := v.([]interface{}); ok {
	// 		item := &MarketItemPrice{}
	// 		for _, val := range v {
	// 			switch val := val.(type) {
	// 			case string:
	// 				if len(item.Date) != 0 {
	// 					item.Count = val
	// 				} else {
	// 					item.Date = val
	// 				}
	// 			case float64:
	// 				item.Price = val
	// 			}
	// 		}
	// 		items = append(items, item)
	// 	}
	// }

	return nil
}
