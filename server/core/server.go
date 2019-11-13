package core

import "github.com/hackbeex/configcenter/util/com"

type Server struct {
	Id   string
	Env  com.EnvType
	Host string
	Port int

	Instances *InstanceTable
}

var server *Server

func InitServer(id string, env string, host string, port int) {
	server = &Server{
		Id:        id,
		Env:       com.EnvType(env),
		Host:      host,
		Port:      port,
		Instances: NewInstanceTable(),
	}
}

func GetServer() *Server {
	return server
}
