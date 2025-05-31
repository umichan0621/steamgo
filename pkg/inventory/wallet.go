package inventory

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type WalletInfo struct {
	WalletCurrency       int32   `json:"wallet_currency"`
	WalletCountry        string  `json:"wallet_country"`
	WalletBalance        float32 `json:"wallet_balance,string"`
	WalletDelayedBalance float32 `json:"wallet_delayed_balance,string"`
	Success              int32   `json:"success"`
}

func (core *Core) WalletBalance() (*WalletInfo, error) {
	url := "https://steamcommunity.com/market/"
	res, err := core.authCore.HttpClient().Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fail to get wallet balance, code: %d", res.StatusCode)
	}
	datax, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	data := string(datax)
	index := strings.Index(data, "g_rgWalletInfo")
	info := &WalletInfo{}
	info.Success = 0
	if index >= 0 {
		data = data[index:]
		start := strings.Index(data, "{")
		end := strings.Index(data, "}")
		if start >= 0 && end >= 0 {
			data = data[start : end+1]
			err := json.Unmarshal([]byte(data), info)
			if err != nil {
				return info, fmt.Errorf("fail to parse json: %s, data: %s", err.Error(), data)
			}
		}
	}
	info.WalletBalance /= 100.0
	info.WalletDelayedBalance /= 100.0
	return info, nil
}
