package log

import (
	"testing"
)

func TestLogger(t *testing.T) {
	Debug("log init++")
	Info("log init++")
	Warn("log init++")
	//logger.Error("log init++")
	//logger.Fatal("log init++")
	//logger.Panic("log init++")
}
