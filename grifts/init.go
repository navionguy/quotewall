package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/navionguy/quotewall/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}
