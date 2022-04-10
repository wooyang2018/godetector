package service

import "testing"

func TestChineseTextDetect(t *testing.T) {
	res := ChineseTextDetect("http://127.0.0.1:8080", "舔的淫水直流")
	t.Logf("%+v", res)
}

func TestEnglishTextDetect(t *testing.T) {
	res := EnglishTextDetect("http://127.0.0.1:8080", "fuck you tip")
	t.Logf("%+v", res)
}
