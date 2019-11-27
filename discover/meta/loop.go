package meta

import (
	"github.com/hackbeex/configcenter/discover/server"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"time"
)

func runLoop() {
	defer func() {
		if err := recover(); err != nil {
			log.Warn("discover loop recover: ", err)
			runLoop()
		}
	}()

	table := GetTable()
	servers := table.servers

	for {
		servers.Range(func(key server.IdKey, val *server.Server) bool {
			//log.Debug("servers loop:", key, val)
			if val.Life <= 0 {
				if val.Status == com.OnlineStatus {
					_ = servers.UpdateStatus(key, com.BreakStatus)
				}
			} else {
				val.Life--
				servers.Store(key, val)
			}
			return true
		})
		time.Sleep(time.Second * 3)
	}
}
