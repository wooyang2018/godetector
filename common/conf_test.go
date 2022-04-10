package common

import (
	"testing"
)

func TestGetServerConfig(t *testing.T) {
	sconfig := GetServerConfig("H:\\0-毕业设计\\godetector\\conf.yaml")
	t.Logf("%+v\n", sconfig)
}

func TestGetWebConfig(t *testing.T) {
	wconfig := GetWebConfig("H:\\0-毕业设计\\godetector\\conf.yaml")
	t.Logf("%+v\n", wconfig)
}
