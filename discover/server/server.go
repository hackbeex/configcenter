package server

import (
	"fmt"
	"github.com/hackbeex/configcenter/discover/com"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/pkg/errors"
)

const (
	KeyServerIdPrefix      = "/config-server/id/"
	KeyServerInstantPrefix = "/config-server/instance/"
	KeyServerAttrEnv       = "env"
	KeyServerAttrHost      = "host"
	KeyServerAttrPost      = "post"
	KeyServerAttrStatus    = "status"
)

type Server struct {
	Id     string
	Host   string
	Port   int
	Env    com.EnvType
	Status com.RunStatus
}

func (s *Server) Register(store *store.Store) error {
	if s.Id == "" {
		err := errors.New("server id require")
		log.Error(err)
		return err
	}

	prefix := KeyServerInstantPrefix + s.Id + "/"
	kvs := map[string]string{
		KeyServerIdPrefix + s.Id:     s.Id,
		prefix + KeyServerAttrHost:   s.Host,
		prefix + KeyServerAttrPost:   fmt.Sprintf("%d", s.Port),
		prefix + KeyServerAttrEnv:    string(s.Env),
		prefix + KeyServerAttrStatus: string(com.OnlineStatus),
	}

	if err := store.PutKeyValues(kvs); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
