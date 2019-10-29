package server

import (
	"github.com/hackbeex/configcenter/discover/store"
)

const (
	KeyConfigServerIdPrefix      = "/config-server/id/"
	KeyConfigServerInstantPrefix = "/config-server/instance/"
	KeyConfigServerAttrEnv       = "env"
	KeyConfigServerAttrHost      = "host"
	KeyConfigServerAttrPost      = "post"
)

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

func (s *Server) Register(store *store.Store) error {

	return nil
}
