package nsq

/*
消费者需要实现nsqd管道的连接，订阅。主题类型"nsfw,protest,cntext,entext"四类，分别设置不同的Handler。
*/

import (
	"encoding/json"
	"fmt"
	"github.com/nsqio/go-nsq"
	"github.com/wuyangdut/godetector/common"
	"log"
)

type ConsumerHandler struct {
	consumers map[string]*nsq.Consumer
	nsqdAddr  string
	topics    []string
}

type NsfwHandler struct {
	EtcdAddr string
	TsAddr   string
}
type ProtestHandler struct {
	EtcdAddr string
	TsAddr   string
}
type CnTextHandler struct {
	EtcdAddr string
	TsAddr   string
}
type EnTextHandler struct {
	EtcdAddr string
	TsAddr   string
}

func (h *NsfwHandler) NsfwImageDetect(data []byte) string {
	TSNsfwModel := h.TsAddr + "/predictions/nsfw_model"
	request := common.NewTSRequest(TSNsfwModel, nil, data)
	rawContent := common.GetHttpResponse(request)
	return rawContent
}

func (h *NsfwHandler) HandleMessage(message *nsq.Message) error {
	req := NsqRequest{}
	err := json.Unmarshal(message.Body, &req)
	if err != nil {
		return err
	}
	image, err := common.DecodeBase64(req.Data)
	if err != nil {
		return err
	}
	//发起推理端请求
	value := h.NsfwImageDetect(image)
	//连接Etcd
	etcd := NewEtcdHandler(h.EtcdAddr)
	defer etcd.Stop()
	//保存结果
	key := common.Sha256(req.Topic + req.Data)
	err = etcd.PutWithLease(key, value)
	fmt.Println(key, " ==> ", value)
	return err
}

func (h *ProtestHandler) ProtestImageDetect(data []byte) string {
	TSProtestModel := h.TsAddr + "/predictions/protest_model"
	request := common.NewTSRequest(TSProtestModel, nil, data)
	rawContent := common.GetHttpResponse(request)
	return rawContent
}

func (h *ProtestHandler) HandleMessage(message *nsq.Message) error {
	req := NsqRequest{}
	err := json.Unmarshal(message.Body, &req)
	if err != nil {
		return err
	}
	image, err := common.DecodeBase64(req.Data)
	if err != nil {
		return err
	}
	//发起推理端请求
	value := h.ProtestImageDetect(image)
	//连接Etcd
	etcd := NewEtcdHandler(h.EtcdAddr)
	defer etcd.Stop()
	//保存结果
	key := common.Sha256(req.Topic + req.Data)
	err = etcd.PutWithLease(key, value)
	fmt.Println(key, " ==> ", value)
	return err
}

func (h *CnTextHandler) ChineseTextDetect(text string) string {
	TSCnTextModel := h.TsAddr + "/predictions/cn_text_model"
	request := common.NewTSRequest(TSCnTextModel, map[string]string{"data": text}, nil)
	response := common.GetHttpResponse(request)
	return response
}

func (h *CnTextHandler) HandleMessage(message *nsq.Message) error {
	req := NsqRequest{}
	err := json.Unmarshal(message.Body, &req)
	if err != nil {
		return err
	}
	//发起推理端请求
	value := h.ChineseTextDetect(req.Data)
	//连接Etcd
	etcd := NewEtcdHandler(h.EtcdAddr)
	defer etcd.Stop()
	//保存结果
	key := common.Sha256(req.Topic + req.Data)
	err = etcd.PutWithLease(key, value)
	fmt.Println(key, " ==> ", value)
	return err
}

func (h *EnTextHandler) EnglishTextDetect(text string) string {
	TSEnTextModel := h.TsAddr + "/predictions/en_text_model"
	request := common.NewTSRequest(TSEnTextModel, map[string]string{"data": text}, nil)
	rawContent := common.GetHttpResponse(request)
	return rawContent
}

func (h *EnTextHandler) HandleMessage(message *nsq.Message) error {
	req := NsqRequest{}
	err := json.Unmarshal(message.Body, &req)
	if err != nil {
		return err
	}
	//发起推理端请求
	value := h.EnglishTextDetect(req.Data)
	//连接Etcd
	etcd := NewEtcdHandler(h.EtcdAddr)
	defer etcd.Stop()
	//保存结果
	key := common.Sha256(req.Topic + req.Data)
	err = etcd.PutWithLease(key, value)
	fmt.Println(key, " ==> ", value)
	return err
}

//NewConsumerHandler 初始化ConsumerHandler
func NewConsumerHandler(conf *common.ServerConfig) *ConsumerHandler {
	handler := ConsumerHandler{}
	handler.nsqdAddr = conf.NsqdAddr
	handler.topics = conf.SupportTopics
	handler.consumers = make(map[string]*nsq.Consumer)
	config := nsq.NewConfig()
	for _, topic := range handler.topics {
		chname := "channel-" + topic
		q, err := nsq.NewConsumer(topic, chname, config)
		if err != nil {
			log.Fatalf("new consumer failed: %+v\n", err)
		}
		switch topic {
		case "nsfw":
			q.AddHandler(&NsfwHandler{EtcdAddr: conf.EtcdAddr, TsAddr: conf.TorchServeAddr["nsfw"]})
		case "protest":
			q.AddHandler(&ProtestHandler{EtcdAddr: conf.EtcdAddr, TsAddr: conf.TorchServeAddr["protest"]})
		case "cntext":
			q.AddHandler(&CnTextHandler{EtcdAddr: conf.EtcdAddr, TsAddr: conf.TorchServeAddr["cntext"]})
		case "entext":
			q.AddHandler(&EnTextHandler{EtcdAddr: conf.EtcdAddr, TsAddr: conf.TorchServeAddr["entext"]})
		}
		handler.consumers[topic] = q
	}
	return &handler
}

//Connect ConsumerHandler正式连接Nsqd
func (c *ConsumerHandler) Connect() {
	for _, topic := range c.topics {
		q := c.consumers[topic]
		err := q.ConnectToNSQD(c.nsqdAddr)
		if err != nil {
			log.Fatalf("consumer connected to nsqd failed: %+v\n", err)
		}
	}
}

//Connect ConsumerHandler测试连接状态
func (c *ConsumerHandler) Check() bool {
	for _, topic := range c.topics {
		q := c.consumers[topic]
		stats := q.Stats()
		if stats.Connections == 0 {
			return false
		}
	}
	return true
}
