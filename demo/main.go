package main

import (
	embedweb "github.com/chendefine/embed-web"
	_ "github.com/chendefine/embed-web/demo/web"
)

func main() {
	ew := embedweb.New(nil)
	log := ew.GetLogger()
	err := ew.StartServer()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("embed-web serve at http://localhost:%d", ew.GetPort())
	select {}
}
