package nsq

import (
	"testing"
	"time"
)

func TestEtcd(t *testing.T) {
	endpoint := "127.0.0.1:2379"
	handler := NewEtcdHandler(endpoint)
	defer handler.Stop()
	handler.Put("one", "test string")
	res := handler.GetWithLoop("one")
	handler.Delete("one")
	t.Log(res)
}

func TestEtcdPutWithLease(t *testing.T) {
	endpoint := "127.0.0.1:2379"
	handler := NewEtcdHandler(endpoint)
	handler.minLeaseTTL = 5
	defer handler.Stop()
	handler.PutWithLease("lease", "test string")
	time.Sleep(time.Second * 6)
	res := handler.Get("lease")
	t.Log(res == "")
}
