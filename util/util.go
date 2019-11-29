package util

import (
	"bytes"
	"encoding/json"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func GetHardwareAddressAsLong() []int64 {
	var long []int64

	inter, _ := net.Interfaces()
	for _, i := range inter {
		macStr := i.HardwareAddr.String()
		if macStr != "" {
			hexStr := strings.Join(strings.Split(macStr, ":"), "")

			integer, err := strconv.ParseInt(hexStr, 16, 64)
			if err != nil {
				log.Error(err)
			} else {
				long = append(long, integer)
			}
		}
	}
	return long
}

func GetUidFromHardwareAddress(port int) (string, error) {
	hardwareAddr := GetHardwareAddressAsLong()
	if len(hardwareAddr) > 0 {
		uid := strconv.FormatInt(hardwareAddr[0], 10) + ":" + strconv.FormatInt(int64(port), 10)
		return uid, nil
	}
	return "", errors.Errorf("can not gain local hardware address")
}

func HttpPostJson(url string, data []byte) (*http.Response, error) {
	log.Debugf("request url[%s], data[%s]", url, string(data))
	return http.Post(url, "application/json", bytes.NewReader(data))
}

func HttpParseResponseToJson(res *http.Response, resp interface{}) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("request body: ", string(body))
	err = json.Unmarshal(body, resp)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
