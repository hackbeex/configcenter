package util

import (
	"github.com/hackbeex/configcenter/util/log"
	"github.com/pkg/errors"
	"net"
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
