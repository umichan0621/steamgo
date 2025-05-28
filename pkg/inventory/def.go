package inventory

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
