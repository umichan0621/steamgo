package inventory

import (
	"github.com/umichan0621/steam/pkg/auth"
)

type Core struct {
	authCore *auth.Core
	language string
}

func (core *Core) Init(authCore *auth.Core) {
	core.authCore = authCore
	core.language = "english"
}

func (core *Core) SetLanguage(language string) { core.language = language }
