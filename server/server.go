package server

import "github.com/coreos/etcd/clientv3"

type EnvType string

const (
	EnvDev  EnvType = "develop"
	EnvProd EnvType = "product"
	EnvTest EnvType = "test"
)

type Server struct {
	Id   string
	Host string
	Port int
	Env  EnvType
}

func (s *Server) Register(etcd *clientv3.Client) error {

	return nil
}
