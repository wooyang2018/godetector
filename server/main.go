//实现推理端的代理程序
package main

import (
	"flag"
	"fmt"
	"github.com/wuyangdut/godetector/common"
	"github.com/wuyangdut/godetector/nsq"
)

func main() {
	fmt.Println("hello, godetector AI server!")
	var conf = flag.String("f", "./conf.yaml", "Input Your Config Yaml Path")
	flag.Parse()
	config := common.GetServerConfig(*conf)
	//尝试连接NSQ
	handler := nsq.NewConsumerHandler(config)
	handler.Connect()
	//成功连上NSQ则阻塞主线程
	if handler.Check() {
		ch := make(chan int, 0)
		<-ch
	}
}
