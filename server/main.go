package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/server/core"
	"github.com/hackbeex/configcenter/server/handler"
	"github.com/hackbeex/configcenter/util"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	registerServer()

	go exitServer()

	go reportHeartbeat()

	go checkInstances()

	runServer()
}

func registerServer() {
	conf := local.Conf.Server
	if conf.Env == "" {
		log.Fatal("env can not be empty")
	}

	id, err := util.GetUidFromHardwareAddress(conf.ListenPort)
	if err != nil {
		log.Fatal(err)
	}

	core.InitServer(id, conf.Env, conf.ListenHost, conf.ListenPort)
	data, _ := json.Marshal(map[string]interface{}{
		"id":   id,
		"host": conf.ListenHost,
		"port": conf.ListenPort,
		"env":  conf.Env,
	})
	discover := local.Conf.Discover
	url := fmt.Sprintf("http://%s:%d/api/v1/discover/server/register", discover.ListenHost, discover.ListenPort)
	_, err = util.HttpPostJson(url, data)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("config server register successful:", id)
}

func exitServer() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Info("config server exiting ...")
	if err := heartbeat(false); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func heartbeat(online bool) error {
	discover := local.Conf.Discover
	url := fmt.Sprintf("http://%s:%d/api/v1/discover/server/heartbeat", discover.ListenHost, discover.ListenPort)
	status := com.OnlineStatus
	if !online {
		status = com.OfflineStatus
	}
	server := core.GetServer()
	data, _ := json.Marshal(map[string]string{
		"id":     server.Id,
		"status": string(status),
	})
	_, err := util.HttpPostJson(url, data)
	if err != nil {
		log.Warn(err)
		return err
	}
	return nil
}

func reportHeartbeat() {
	for {
		if err := heartbeat(true); err != nil {
			continue
		}
		time.Sleep(time.Second * 10)
	}
}

func checkInstances() {
	server := core.GetServer()
	instances := server.Instances
	for {
		instances.Range(func(instanceId string, val *core.Instance) bool {
			if val.Life <= 0 {
				if val.Status == com.OnlineStatus {
					val.Status = com.BreakStatus
					instances.Store(instanceId, val)
				}
			} else {
				val.Life--
				instances.Store(instanceId, val)
			}
			return true
		})
		time.Sleep(time.Second)
	}
}

func runServer() {
	r := gin.Default()

	//TODO： portal auth, client token

	//instance api
	r.POST("/api/v1/client/config/list", handler.GetClientConfigList)
	r.POST("/api/v1/client/config/watch", handler.WatchConfig)
	r.POST("/api/v1/client/exit", handler.ExitClient)

	//portal api
	r.POST("/api/v1/app/list", handler.GetAppList)
	r.POST("/api/v1/app/detail", handler.GetAppDetail)
	r.POST("/api/v1/app/create", handler.CreateApp)
	r.POST("/api/v1/cluster/create", handler.CreateCluster)
	r.POST("/api/v1/namespace/create", handler.CreateNamespace)
	r.POST("/api/v1/config/detail", handler.GetConfigDetail)
	r.POST("/api/v1/config/list", handler.GetConfigList)
	r.POST("/api/v1/config/create", handler.CreateConfig)
	r.POST("/api/v1/config/update", handler.UpdateConfig)
	r.POST("/api/v1/config/delete", handler.DeleteConfig)
	r.POST("/api/v1/config/history", handler.GetConfigHistory)
	r.POST("/api/v1/config/release", handler.ReleaseConfig)
	r.POST("/api/v1/config/release/history", handler.GetConfigReleaseHistory)
	r.POST("/api/v1/config/rollback", handler.RollbackConfig)
	r.POST("/api/v1/config/sync", handler.SyncConfig)
	r.POST("/api/v1/instance/list", handler.GetInstanceList)

	conf := local.Conf.Server
	addr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	log.Infof("config server run at: %s", addr)

	if err := r.Run(addr); err != nil {
		log.Panic(err)
	}
}
