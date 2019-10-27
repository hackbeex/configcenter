package discover

import (
	"fmt"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
	"os"
)

func Run() {
	table := initTable()
	go runServer(table)
}

func runServer(table *Table) {
	var srv *fasthttp.Server

	conf := local.Conf.Discover
	srv = &fasthttp.Server{
		Handler:            serverHandler(table),
		Name:               conf.Name,
		Concurrency:        conf.Concurrency,
		ReadBufferSize:     conf.ReadBufferSize,
		WriteBufferSize:    conf.WriteBufferSize,
		DisableKeepalive:   conf.DisabledKeepAlive,
		ReduceMemoryUsage:  conf.ReduceMemoryUsage,
		MaxRequestBodySize: conf.MaxRequestBodySize,
	}

	host := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	log.Infof("config server start at: %s", host)
	listener, err := reuseport.Listen("tcp4", host)
	err = srv.Serve(listener)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

func serverHandler(table *Table) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

	}
}
