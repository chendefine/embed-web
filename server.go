package embedweb

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func (ew *EmbedWeb) initServer(engine *gin.Engine) {
	ew.server = new(server)
	if engine == nil {
		engine = defaultEngine()
	}
	ew.server.engine = engine
}

func (ew *EmbedWeb) GetEngine() *gin.Engine {
	return ew.server.engine
}

func (ew *EmbedWeb) SetEngine(engine *gin.Engine) {
	ew.server.engine = engine
}

func (ew *EmbedWeb) StartServer() error {
	err := ew.server.start(ew.GetPort(), ew.GetPublic())
	if err != nil {
		return err
	}
	if port := ew.server.port(); port != ew.GetPort() {
		ew.SetPort(port)
	}
	return nil
}

func (ew *EmbedWeb) StopServer() error {
	err := ew.server.stop()
	return err
}

func (ew *EmbedWeb) RestartServer() error {
	ew.StopServer()
	return ew.StartServer()
}

type stat = uint32

const (
	statIdle stat = iota
	statPreparing
	statRunning
)

var (
	errServerNotIdle    = errors.New("server is not idle")
	errServerNotRunning = errors.New("server is not running")
	errServerRunTimeout = errors.New("server run timeout")
)

type server struct {
	lock     sync.Mutex
	stat     uint32
	engine   *gin.Engine
	listener net.Listener
}

func (svr *server) start(port int, public bool) error {
	svr.lock.Lock()
	defer svr.lock.Unlock()

	if svr.stat != statIdle {
		return errServerNotIdle
	}
	svr.stat = statPreparing

	var addr string
	if p := getPortFromCmdArgs(); p != nil {
		port = *p
	}

	if public {
		addr = fmt.Sprintf(":%d", port)
	} else {
		addr = fmt.Sprintf("localhost:%d", port)
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		svr.stat = statIdle
		return err
	}
	svr.listener = ln

	if svr.engine == nil {
		svr.engine = gin.Default()
	}

	err = svr.registAssetsRoutes()
	if err != nil {
		_ = svr.close()
		return err
	}

	go func() {
		err = http.Serve(svr.listener, svr.engine)
	}()
	time.Sleep(100 * time.Millisecond)

	health := false
	for i := 0; i < 10; i++ {
		if svr.isIndexStatusOK() {
			health = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !health {
		_ = svr.close()
		return errServerRunTimeout
	}

	svr.stat = statRunning
	return nil
}

func (svr *server) stop() error {
	svr.lock.Lock()
	defer svr.lock.Unlock()

	if svr.stat != statRunning {
		return errServerNotRunning
	}

	return svr.close()
}

func (svr *server) close() error {
	ln := svr.listener
	svr.stat, svr.listener = statIdle, nil
	if err := ln.Close(); err != nil {
		return err
	}
	return nil
}

func (svr *server) port() int {
	if svr.listener == nil {
		return 0
	}
	return svr.listener.Addr().(*net.TCPAddr).Port
}

func (svr *server) isIndexStatusOK() bool {
	port := svr.port()
	if port == 0 {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	resp, err := resty.New().R().SetContext(ctx).Get(fmt.Sprintf("http://localhost:%d", port))
	if err != nil || resp == nil {
		return false
	}
	return resp.StatusCode() == http.StatusOK
}

func defaultEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	return engine
}
