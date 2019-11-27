package client

import (
	"encoding/json"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util/log"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
)

type Cache struct {
	sync.RWMutex
	version  string
	dir      string
	filename string
	filepath string
}

func NewCache(filename string) *Cache {
	cache := &Cache{
		version: "1.0.0",
	}

	dir := local.Conf.Server.CacheDir
	if dir == "" {
		if runtime.GOOS == "windows" {
			dir = `C:\opt\data\configcenter\cache\`
		} else {
			dir = `/opt/data/configcenter/cache/cache/`
		}
	}
	cache.dir = dir
	cache.filename = filename
	cache.filepath = dir + filename

	fh, err := os.OpenFile(cache.filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("open file fail: ", err)
	}
	_ = fh.Close()

	return cache
}

type cacheData struct {
	Version string            `json:"version"`
	Data    map[string]string `json:"data"`
}

func (c *Cache) Store(data map[string]string) error {
	writeData := cacheData{
		Version: c.version,
		Data:    data,
	}
	d, err := json.Marshal(writeData)
	if err != nil {
		log.Error(err)
		return err
	}

	c.Lock()
	err = ioutil.WriteFile(c.filepath, d, 0666)
	if err != nil {
		log.Error(err)
		return err
	}
	c.Unlock()

	//todo: backup old cache file

	return nil
}

func (c *Cache) Load() (map[string]string, error) {
	var readData = cacheData{}

	c.Lock()
	d, err := ioutil.ReadFile(c.filepath)
	if err != nil {
		log.Error(err)
		return readData.Data, err
	}
	c.Unlock()

	if err := json.Unmarshal(d, &readData); err != nil {
		log.Error(err)
		return readData.Data, err
	}
	return readData.Data, nil
}
