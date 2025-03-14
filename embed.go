package eweb

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	gorm "gorm.io/gorm"
)

const (
	embedDbFile  = "embed.db"
	embedLogFile = "embed.log"
	appLogFile   = "app.log"
)

var (
	baseDirPath = path.Dir(os.Args[0])
)

type EmbedWeb struct {
	embedDB  *gorm.DB
	embedLog *logrus.Logger

	config *config
	engine *gin.Engine
	log    *logrus.Logger

	started uint32
}

func New(engine *gin.Engine) *EmbedWeb {
	innerMutex.Lock()
	defer innerMutex.Unlock()

	if innerEmbedWeb != nil {
		innerEmbedWeb.embedLog.Errorf("embed web already initialized")
		return innerEmbedWeb
	} else if engine == nil {
		engine = gin.Default()
	}

	log := newEmbedLog()
	eweb := &EmbedWeb{embedLog: log, engine: engine}
	eweb.initEmbedDB()
	eweb.initConfig()
	eweb.initAppLog()
	innerEmbedWeb = eweb
	return eweb
}

func (ew *EmbedWeb) Engine() *gin.Engine {
	return ew.engine
}

func (ew *EmbedWeb) DB() *gorm.DB {
	return ew.embedDB
}

func (ew *EmbedWeb) Log() *logrus.Logger {
	return ew.log
}

func (ew *EmbedWeb) InjectWebAssets(route string, assets embed.FS, dirPath string, indexPath string) {
	InjectWebAssets(route, assets, dirPath, indexPath)
}

func (ew *EmbedWeb) Run() error {
	if !atomic.CompareAndSwapUint32(&ew.started, 0, 1) {
		return errors.New("embed web already started")
	}
	defer atomic.StoreUint32(&ew.started, 0)

	ew.injectWebAssets()
	ln := ew.tryListen()
	ew.embedLog.Infof("embed web server listen on %s", ln.Addr().String())
	return http.Serve(ln, ew.engine)
}

func (ew *EmbedWeb) tryListen() net.Listener {
	host, port := "localhost", ew.config.Port
	if ew.config.Public {
		host = ""
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err == nil {
		return ln
	}

	ew.embedLog.Warnf("listen on port %d error: %v, retry listen another port", ew.config.Port, err)

	port = 0
	addr = fmt.Sprintf("%s:%d", host, port)
	ln, err = net.Listen("tcp", addr)
	if err != nil {
		ew.embedLog.Fatalf("listen on port %d error: %v, abort to serve embed web", ew.config.Port, err)
	}

	return ln
}

func (ew *EmbedWeb) injectWebAssets() {
	st, _ := fs.Sub(embedAssets.assets, embedAssets.dirPath)
	index, _ := embedAssets.assets.ReadFile(path.Join(embedAssets.dirPath, embedAssets.indexPath))
	ew.engine.StaticFS(embedAssets.route, http.FS(st))
	ew.engine.GET("/", func(c *gin.Context) { c.Data(http.StatusOK, "text/html; charset=utf-8", index) })
}
