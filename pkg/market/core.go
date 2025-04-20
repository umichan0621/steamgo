package market

import (
	"net/http"
)

type Core struct {
	httpClient *http.Client
}

func (core *Core) Init(httpClient *http.Client) {
	core.httpClient = httpClient
}
