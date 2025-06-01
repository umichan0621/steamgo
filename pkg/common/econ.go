package common

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
