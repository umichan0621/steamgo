package market

import (
	"steam/pkg/auth"
)

type Core struct {
	authCore *auth.Core
}

func (core *Core) Init(authCore *auth.Core) {
	core.authCore = authCore
}
