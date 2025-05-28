package market

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type Core struct {
	httpClient *http.Client
}

func (core *Core) Init(httpClient *http.Client) {
	core.httpClient = httpClient
}

func findTitle(n *html.Node) {
	// if n.Type == html.ElementNode && n.Data == "title" {
	// 	for c := n.FirstChild; c != nil; c = c.NextSibling {
	// 		fmt.Println(c.Data) // 这里c.Data将是标题文本内容
	// 	}
	// }
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findTitle(c) // 递归查找其他可能的title元素
	}
}

func (core *Core) WalletBalance() (float64, error) {
	url := "https://store.steampowered.com/account/history/"
	res, err := core.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
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
