package market

import (
<<<<<<< HEAD
	"fmt"
	"io"
	"net/http"
	"steam/pkg/auth"
	"strings"
=======
	"steam/pkg/auth"
>>>>>>> main
)

type Core struct {
	authCore *auth.Core
}

func (core *Core) Init(authCore *auth.Core) {
	core.authCore = authCore
}

func (core *Core) WalletBalance() (float64, error) {
	url := "https://store.steampowered.com/account/history/"
	res, err := core.authCore.HttpClient().Get(url)
	if err != nil {
		return 0, err
	}
	fmt.Println(core.authCore.HttpClient().Jar)

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fail to get wallet balance, code: %d", res.StatusCode)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	x := string(data)

	// q, err := html.Parse(strings.NewReader(x))
	// if err != nil {
	// 	return 0, err
	// }
	// fmt.Println(x)
	// findTitle(q)

	index := strings.Index(x, "wallet")
	fmt.Println(index)
	if index >= 0 {
		fmt.Println(data[index:])

	} else {
		fmt.Println("not found")
	}
	// fmt.Println(x)
	// url = SteamUrl.URL_STEAM_STORE + '/account/history/'
	// response = self._session.get(url)
	// response_soup = bs4.BeautifulSoup(response.text, "html.parser")
	// balance = response_soup.find(id='header_wallet_balance').string
	// if convert_to_decimal:
	// 	return parse_price(balance)
	// else:
	// 	return balance
	return 0, nil
}
