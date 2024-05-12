package main

import (
	"github.com/gridsx/micro-conf/config"
	"github.com/gridsx/micro-conf/grace"
	"github.com/gridsx/micro-conf/server"
)

var app = config.App

func main() {
	grace.Prepare()
	server.Serve(app.Server.Port)
}
