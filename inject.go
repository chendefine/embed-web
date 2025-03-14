package embedweb

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"path"

	"github.com/gin-gonic/gin"
)

var (
	assetsFs       embed.FS
	assetsPrefix   string
	assetsInjected bool
)

func InjectAssets(prefix string, assets embed.FS) {
	if assetsInjected {
		panic("embed-web assets should be injected only once")
	}
	assetsInjected = true
	assetsFs = assets
	assetsPrefix, _ = url.JoinPath("/", prefix)
}

func (svr *server) registAssetsRoutes() error {
	var index []byte
	var dirName string

	fe, err := assetsFs.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read web assets dir failed: %w", err)
	} else if len(fe) > 1 || len(fe) == 1 && !fe[0].IsDir() {
		return fmt.Errorf("web assets dir should only have one directory")
	} else if len(fe) == 1 && fe[0].IsDir() {
		dirName = fe[0].Name()
	}

	if dirName != "" {
		index, _ = assetsFs.ReadFile(path.Join(dirName, "index.html"))
	}
	indexFn := func(c *gin.Context) { c.Data(http.StatusOK, "text/html; charset=utf-8", index) }

	registed := false
	assetsPath, _ := url.JoinPath(assetsPrefix, "*filepath")
	for _, route := range svr.engine.Routes() {
		if ((route.Path == assetsPrefix) || (route.Path == assetsPath)) && route.Method == "GET" {
			registed = true
			break
		}
	}
	if !registed && dirName != "" {
		st, _ := fs.Sub(assetsFs, dirName)
		svr.engine.StaticFS(assetsPrefix, http.FS(st))
	}

	registed = false
	for _, route := range svr.engine.Routes() {
		if ((route.Path == "/") || (route.Path == "/*filepath")) && route.Method == "GET" {
			registed = true
			break
		}
	}
	if !registed {
		svr.engine.GET("/", indexFn)
	}
	return nil
}
