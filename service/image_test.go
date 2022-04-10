package service

import (
	"github.com/wuyangdut/godetector/common"
	"testing"
)

func TestNsfwLocalImageDetect(t *testing.T) {
	res := NsfwLocalImageDetect("http://127.0.0.1:8080", "H:\\0-毕业设计\\godetector\\test\\test_img\\porn.jpg")
	t.Logf("%+v", res)
}

func TestNsfwImageDetect(t *testing.T) {
	file := common.ReadFile("H:\\0-毕业设计\\godetector\\test\\test_img\\porn.jpg")
	res := NsfwImageDetect("http://127.0.0.1:8080", file)
	t.Logf("%+v", res)
}

func TestProtestLocalImageDetect(t *testing.T) {
	res := ProtestLocalImageDetect("http://127.0.0.1:8080", "H:\\0-毕业设计\\godetector\\test\\test_img\\protest_fire.png")
	t.Logf("%+v", res)
}
