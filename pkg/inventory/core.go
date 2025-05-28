package inventory

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

type Core struct {
	httpClient *http.Client
}

func (core *Core) Init(httpClient *http.Client) {
	core.httpClient = httpClient
}

// type InventoryAppStats struct {
// 	AppID            uint64                       `json:"appid"`
// 	Name             string                       `json:"name"`
// 	AssetCount       uint32                       `json:"asset_count"`
// 	Icon             string                       `json:"icon"`
// 	Link             string                       `json:"link"`
// 	InventoryLogo    string                       `json:"inventory_logo"`
// 	TradePermissions string                       `json:"trade_permissions"`
// 	Contexts         map[string]*InventoryContext `json:"rgContexts"`
// }

var inventoryContextRegexp = regexp.MustCompile("var g_rgAppContextData = (.*?);")

func (core *Core) GetInventoryAppStats() error {
	//(map[string]InventoryAppStats, error) {
	resp, err := core.httpClient.Get("https://steamcommunity.com/profiles/" + "76561198276202537" + "/inventory")
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	m := inventoryContextRegexp.FindSubmatch(body)
	if m == nil || len(m) != 2 {
		return err
	}
	fmt.Println(string(m[1]))
	// inven := map[string]InventoryAppStats{}
	// if err = json.Unmarshal(m[1], &inven); err != nil {
	// 	return nil, err
	// }

	return nil
}

type EconDesc struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Color string `json:"color"`
}

type EconTag struct {
	InternalName          string `json:"internal_name"`
	Category              string `json:"category"`
	LocalizedCategoryName string `json:"localized_category_name"`
	LocalizedTagName      string `json:"localized_tag_name"`
}

type EconAction struct {
	Link string `json:"link"`
	Name string `json:"name"`
}

type Asset struct {
	AppID      uint32 `json:"appid"`
	ContextID  uint64 `json:"contextid,string"`
	AssetID    uint64 `json:"assetid,string"`
	ClassID    uint64 `json:"classid,string"`
	InstanceID uint64 `json:"instanceid,string"`
	Amount     uint64 `json:"amount,string"`
}

type EconItemDesc struct {
	ClassID         uint64        `json:"classid,string"`    // for matching with EconItem
	InstanceID      uint64        `json:"instanceid,string"` // for matching with EconItem
	Tradable        int           `json:"tradable"`
	BackgroundColor string        `json:"background_color"`
	IconURL         string        `json:"icon_url"`
	IconLargeURL    string        `json:"icon_url_large"`
	IconDragURL     string        `json:"icon_drag_url"`
	Name            string        `json:"name"`
	NameColor       string        `json:"name_color"`
	MarketName      string        `json:"market_name"`
	MarketHashName  string        `json:"market_hash_name"`
	MarketFeeApp    uint32        `json:"market_fee_app"`
	Comodity        bool          `json:"comodity"`
	Actions         []*EconAction `json:"actions"`
	Tags            []*EconTag    `json:"tags"`
	Descriptions    []*EconDesc   `json:"descriptions"`
}

type Response struct {
	Assets              []Asset         `json:"assets"`
	Descriptions        []*EconItemDesc `json:"descriptions"`
	Success             int             `json:"success"`
	HasMore             int             `json:"more_items"`
	LastAssetID         string          `json:"last_assetid"`
	TotalInventoryCount int             `json:"total_inventory_count"`
	ErrorMsg            string          `json:"error"`
}

type InventoryItem struct {
	AppID      uint32        `json:"appid"`
	ContextID  uint64        `json:"contextid"`
	AssetID    uint64        `json:"id,string,omitempty"`
	ClassID    uint64        `json:"classid,string,omitempty"`
	InstanceID uint64        `json:"instanceid,string,omitempty"`
	Amount     uint64        `json:"amount,string"`
	Desc       *EconItemDesc `json:"-"` /* May be nil  */
}

func (core *Core) GetAll(steamID string, appID, contextID, startAssetID uint64, items *[]InventoryItem) (hasMore bool, lastAssetID uint64, err error) {
	params := url.Values{
		"l": {"schinese"},
	}
	if startAssetID != 0 {
		params.Set("start_assetid", strconv.FormatUint(startAssetID, 10))
		params.Set("count", "50")
	} else {
		params.Set("count", "50")
	}
	url := fmt.Sprintf("http://steamcommunity.com/inventory/%s/%d/%d?", steamID, appID, contextID) + params.Encode()
	res, err := core.httpClient.Get(url)
	if err != nil {
		return false, 0, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return false, 0, err
	}
	resp := Response{}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return false, 0, err
	}

	descriptions := make(map[string]int)
	for i, desc := range resp.Descriptions {
		key := fmt.Sprintf("%d_%d", desc.ClassID, desc.InstanceID)
		descriptions[key] = i
	}

	for _, asset := range resp.Assets {
		var desc *EconItemDesc

		key := fmt.Sprintf("%d_%d", asset.ClassID, asset.InstanceID)
		if d, ok := descriptions[key]; ok {
			desc = resp.Descriptions[d]
		}

		item := InventoryItem{
			AppID:      asset.AppID,
			ContextID:  asset.ContextID,
			AssetID:    asset.AssetID,
			ClassID:    asset.ClassID,
			InstanceID: asset.InstanceID,
			Amount:     asset.Amount,
			Desc:       desc,
		}

		if item.Desc.Tradable != 0 {
			*items = append(*items, item)
		}
	}
	hasMore = resp.HasMore != 0
	if !hasMore {
		return hasMore, 0, nil
	}
	lastAssetID, err = strconv.ParseUint(resp.LastAssetID, 10, 64)
	if err != nil {
		return hasMore, 0, err
	}

	return hasMore, lastAssetID, nil
}
