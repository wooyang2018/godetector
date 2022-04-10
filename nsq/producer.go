package nsq

import (
	"encoding/json"
	"fmt"
	"github.com/nsqio/go-nsq"
)

type NsqRequest struct {
	Topic string `yaml:"topic"`
	Data  string `yaml:"data"`
}

type ProducerHandler struct {
	producer *nsq.Producer
	nsqAddr  string
}

func NewProducerHandler(nsqAddr string) *ProducerHandler {
	handler := ProducerHandler{}
	handler.nsqAddr = nsqAddr
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(handler.nsqAddr, config)
	if err != nil {
		fmt.Printf("Error new producer handler: %s", err)
		return nil
	}
	handler.producer = producer
	return &handler
}

func (p *ProducerHandler) Publish(request NsqRequest) {
	bytes, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("json marshal error: %s", err)
		return
	}
	//检查producer的状态，断线需要重连
	if err := p.producer.Ping(); err != nsq.ErrAlreadyConnected {
		config := nsq.NewConfig()
		p.producer, _ = nsq.NewProducer(p.nsqAddr, config)
	}
	err = p.producer.Publish(request.Topic, bytes)
	if err != nil {
		fmt.Printf("%s topic publish error: %s", request.Topic, err)
		return
	}
}

func (p *ProducerHandler) Stop() {
	err := p.producer.Ping()
	if err != nsq.ErrStopped {
		p.producer.Stop()
	} else {
		fmt.Printf("no need to stop: %s", err)
	}
}

func (p *ProducerHandler) PublishManyAsync(requests []NsqRequest) {
	//检查producer的状态，断线需要重连
	if err := p.producer.Ping(); err != nsq.ErrAlreadyConnected {
		config := nsq.NewConfig()
		p.producer, _ = nsq.NewProducer(p.nsqAddr, config)
	}
	msgCount := len(requests)
	responseChan := make(chan *nsq.ProducerTransaction, msgCount)
	for i := 0; i < msgCount; i++ {
		request := requests[i]
		bytes, err := json.Marshal(request)
		if err != nil {
			fmt.Printf("json marshal error: %s", err)
			return
		}
		err = p.producer.PublishAsync(request.Topic, bytes, responseChan, "godetector")
		if err != nil {
			fmt.Printf("%s topic async publish error: %s", request.Topic, err)
			return
		}
	}

	for i := 0; i < msgCount; i++ {
		trans := <-responseChan
		if trans.Error != nil {
			fmt.Println(trans.Error.Error())
		}
		if trans.Args[0].(string) != "godetector" {
			fmt.Printf(`proxied arg "%s" != "godetector"`, trans.Args[0].(string))
		}
	}
}
