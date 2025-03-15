package eweb

import (
	"context"
	"embed"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	gorm "gorm.io/gorm"
)

var (
	innerEmbedWeb *EmbedWeb
	innerMutex    sync.Mutex

	embedAssets assetsWrap
)

type assetsWrap struct {
	route     string
	assets    embed.FS
	dirPath   string
	indexPath string
}

func Init(engine *gin.Engine) {
	innerEmbedWeb = New(engine)
}

func Engine() *gin.Engine {
	return innerEmbedWeb.Engine()
}

func DB() *gorm.DB {
	return innerEmbedWeb.DB()
}

func Log() *logrus.Logger {
	return innerEmbedWeb.Log()
}

func GetWebServerPort() int {
	return innerEmbedWeb.GetWebServerPort()
}

func GetWebServerPublic() bool {
	return innerEmbedWeb.GetWebServerPublic()
}

func GetLogLevel() string {
	return innerEmbedWeb.GetLogLevel()
}

func SetWebServerPort(ctx context.Context, port int) error {
	return innerEmbedWeb.SetWebServerPort(ctx, port)
}

func SetWebServerPublic(ctx context.Context, public bool) error {
	return innerEmbedWeb.SetWebServerPublic(ctx, public)
}

func SetLogLevel(ctx context.Context, level string) error {
	return innerEmbedWeb.SetLogLevel(ctx, level)
}

func InjectWebAssets(route string, assets embed.FS, dirPath string, indexPath string) {
	embedAssets = assetsWrap{route: route, assets: assets, dirPath: dirPath, indexPath: indexPath}
}

func Run() {
	innerEmbedWeb.Run()
}
