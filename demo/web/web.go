package web

import (
	"embed"

	embedweb "github.com/chendefine/embed-web"
)

//go:embed dist
var assets embed.FS

func init() {
	embedweb.InjectAssets("web", assets)
}
