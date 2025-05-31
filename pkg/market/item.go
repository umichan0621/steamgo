package market

import (
	"fmt"
	"io"
	"os"
)

func (core *Core) GetItemMarketList() (error, error) {
	// https://steamcommunity.com/market/listings/730/Glove%20Case
	res, err := core.authCore.HttpClient().Get("https://steamcommunity.com/market/listings/730/StatTrak%E2%84%A2%20USP-S%20%7C%20Ticket%20to%20Hell%20%28Factory%20New%29")
	if err != nil {
		return nil, err
	}

	// res, err := core.httpClient.Get("https://steamcommunity.com/market/listings/730/P250%20%7C%20Boreal%20Forest%20%28Field-Tested%29/render/?query=&start=0&count=10")
	// if err != nil {
	// 	return nil, err
	// }

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(data))
	file, err := os.OpenFile("1.html", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, nil
	}
	defer file.Close()
	file.WriteString(string(data))
	return nil, nil
}
