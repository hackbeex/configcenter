package main

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/server"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
	"os"
	"time"
)

func ConnectToEtcd() *clientv3.Client {
	config := local.Conf.Etcd

	cli, err := clientv3.New(
		clientv3.Config{
			Endpoints:            config.Endpoints,
			AutoSyncInterval:     time.Duration(config.AutoSyncInterval) * time.Second,
			DialTimeout:          time.Duration(config.DialTimeout) * time.Second,
			DialKeepAliveTime:    time.Duration(config.DialKeepAliveTime) * time.Second,
			DialKeepAliveTimeout: time.Duration(config.DialKeepAliveTimeout) * time.Second,
			Username:             config.Username,
			Password:             config.Password,
		},
	)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	return cli
}

func main() {
	//etcdClient := ConnectToEtcd()

	var serv *fasthttp.Server

	serverConf := local.Conf.Server
	serv = &fasthttp.Server{
		Handler: server.MainRequestHandler,

		Name:               serverConf.Name,
		Concurrency:        serverConf.Concurrency,
		ReadBufferSize:     serverConf.ReadBufferSize,
		WriteBufferSize:    serverConf.WriteBufferSize,
		DisableKeepalive:   serverConf.DisabledKeepAlive,
		ReduceMemoryUsage:  serverConf.ReduceMemoryUsage,
		MaxRequestBodySize: serverConf.MaxRequestBodySize,
	}

	host := fmt.Sprintf("%s:%d", serverConf.ListenHost, serverConf.ListenPort)
	log.Infof("config server start at: %s", host)
	listener, err := reuseport.Listen("tcp4", host)
	err = serv.Serve(listener)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}
