package embedweb

import (
	"os"
	"path"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	embedDbFile  = "embed.db"
	embedLogFile = "embed.log"
)

var (
	eweb        *EmbedWeb
	lock        sync.Mutex
	baseDirPath = path.Dir(os.Args[0])
)

type EmbedWeb struct {
	cfg    *config
	log    *logger
	db     *gorm.DB
	server *server
}

func New(engine *gin.Engine) *EmbedWeb {
	lock.Lock()
	defer lock.Unlock()

	if eweb != nil {
		return eweb
	}
	eweb = new(EmbedWeb)
	eweb.init(engine)
	return eweb
}

func (ew *EmbedWeb) init(engine *gin.Engine) {
	ew.initLogger()
	ew.initDB()
	ew.initConfig()
	ew.initServer(engine)
	ew.initLogger()
}
