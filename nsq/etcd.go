package nsq

import (
	"fmt"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"log"
	"time"
)

var requestTimeout = 5 * time.Second
var dialTimeout = 5 * time.Second

type EtcdHandler struct {
	cli         *clientv3.Client
	minLeaseTTL int64 //数据失效时间，单位秒
	getLoopTTL  int64 //循环查询的超时时间，单位秒
}

func NewEtcdHandler(addr string) *EtcdHandler {
	handler := EtcdHandler{}
	var err error
	handler.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{addr},
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatalf("Error new etcd handler: %s\n", err)
		return nil
	}
	handler.minLeaseTTL = 100
	handler.getLoopTTL = 5
	return &handler
}

func (h *EtcdHandler) Stop() {
	err := h.cli.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (h *EtcdHandler) PutWithLease(key string, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := h.cli.Grant(ctx, h.minLeaseTTL)
	if err != nil {
		fmt.Printf("Error etcd grant: %s\n", err)
		return err
	}
	_, err = h.cli.Put(ctx, key, value, clientv3.WithLease(resp.ID))
	cancel()
	if err != nil {
		fmt.Printf("Error etcd put: %s\n", err)
		return err
	}
	return nil
}

func (h *EtcdHandler) Put(key string, value string) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := h.cli.Put(ctx, key, value)
	cancel()
	if err != nil {
		switch err {
		case context.Canceled:
			fmt.Printf("ctx is canceled by another routine: %v\n", err)
		case context.DeadlineExceeded:
			fmt.Printf("ctx is attached with a deadline is exceeded: %v\n", err)
		case rpctypes.ErrEmptyKey:
			fmt.Printf("client-side error: %v\n", err)
		default:
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
	}
}

func (h *EtcdHandler) Get(key string) string {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := h.cli.Get(ctx, key)
	cancel()
	if err != nil || len(resp.Kvs) == 0 {
		//fmt.Printf("Error etcd get: %s\n", key)
		return ""
	}
	return string(resp.Kvs[0].Value)
}

func (h *EtcdHandler) GetWithLoop(key string) (res string) {
	t := time.NewTicker(time.Millisecond * 500)
	start := time.Now()
	//定时尝试获取Etcd的Key
	for range t.C {
		res = h.Get(key)
		if res != "" { //取到Key
			t.Stop()
			break
		}
		if int64(time.Now().Sub(start).Seconds()) > h.getLoopTTL {
			t.Stop()
			break
		}
	}
	return
}

func (h *EtcdHandler) Delete(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	// count keys about to be deleted
	gresp, err := h.cli.Get(ctx, key)
	if len(gresp.Kvs) == 0 || err != nil {
		return false
	}
	// delete the keys
	dresp, err := h.cli.Delete(ctx, key)
	if err != nil {
		fmt.Printf("Error etcd delete: %s\n", err)
		return false
	}
	return int64(len(gresp.Kvs)) == dresp.Deleted
}
