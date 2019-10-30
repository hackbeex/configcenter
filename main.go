package main

import (
	"fmt"
	"github.com/hackbeex/configcenter/discover"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/server"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
)

func init() {
	//todo: set run single or together mode
	go discover.Run()

}

func main() {
	var srv *fasthttp.Server

	conf := local.Conf.Server
	srv = &fasthttp.Server{
		Handler: server.MainRequestHandler(),

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
		log.Panic(err)
	}
}
