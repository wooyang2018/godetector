package nsq

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nsqio/go-nsq"
	"github.com/wuyangdut/godetector/common"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

type MyTestHandler struct {
	t                *testing.T
	q                *nsq.Consumer
	messagesSent     int
	messagesReceived int
	messagesFailed   int
}

func (h *MyTestHandler) LogFailedMessage(message *nsq.Message) {
	h.messagesFailed++
	h.q.Stop()
}

func (h *MyTestHandler) HandleMessage(message *nsq.Message) error {
	if string(message.Body) == "TOBEFAILED" {
		h.messagesReceived++
		return errors.New("fail this message")
	}

	data := struct {
		Msg string
	}{}

	err := json.Unmarshal(message.Body, &data)
	if err != nil {
		return err
	}

	msg := data.Msg
	if msg != "single" && msg != "double" {
		h.t.Error("message 'action' was not correct: ", msg, data)
	}
	h.messagesReceived++
	return nil
}

//TestConsumer 官方例子
func TestConsumer(t *testing.T) {
	consumerTest(t, nil)
}

func consumerTest(t *testing.T, cb func(c *nsq.Config)) {
	config := nsq.NewConfig()
	laddr := "127.0.0.1"
	// so that the test can simulate binding consumer to specified address
	config.LocalAddr, _ = net.ResolveTCPAddr("tcp", laddr+":0")
	// so that the test can simulate reaching max requeues and a call to LogFailedMessage
	config.DefaultRequeueDelay = 0
	// so that the test wont timeout from backing off
	config.MaxBackoffDuration = time.Millisecond * 50
	if cb != nil {
		cb(config)
	}
	topicName := "rdr_test"
	if config.Deflate {
		topicName = topicName + "_deflate"
	} else if config.Snappy {
		topicName = topicName + "_snappy"
	}
	if config.TlsV1 {
		topicName = topicName + "_tls"
	}
	topicName = topicName + strconv.Itoa(int(time.Now().Unix()))
	q, _ := nsq.NewConsumer(topicName, "ch", config)

	h := &MyTestHandler{
		t: t,
		q: q,
	}
	q.AddHandler(h)

	SendMessage(t, topicName, "pub", []byte(`{"msg":"single"}`))
	SendMessage(t, topicName, "mpub", []byte("{\"msg\":\"double\"}\n{\"msg\":\"double\"}"))
	SendMessage(t, topicName, "pub", []byte("TOBEFAILED"))
	h.messagesSent = 4

	addr := "127.0.0.1:4150"
	err := q.ConnectToNSQD(addr)
	if err != nil {
		t.Fatal(err)
	}

	stats := q.Stats()
	if stats.Connections == 0 {
		t.Fatal("stats report 0 connections (should be > 0)")
	}

	err = q.ConnectToNSQD(addr)
	if err == nil {
		t.Fatal("should not be able to connect to the same NSQ twice")
	}

	err = q.DisconnectFromNSQD("1.2.3.4:4150")
	if err == nil {
		t.Fatal("should not be able to disconnect from an unknown nsqd")
	}

	err = q.ConnectToNSQD("1.2.3.4:4150")
	if err == nil {
		t.Fatal("should not be able to connect to non-existent nsqd")
	}

	err = q.DisconnectFromNSQD("1.2.3.4:4150")
	if err != nil {
		t.Fatal("should be able to disconnect from an nsqd - " + err.Error())
	}

	<-q.StopChan

	stats = q.Stats()
	if stats.Connections != 0 {
		t.Fatalf("stats report %d active connections (should be 0)", stats.Connections)
	}

	stats = q.Stats()
	if stats.MessagesReceived != uint64(h.messagesReceived+h.messagesFailed) {
		t.Fatalf("stats report %d messages received (should be %d)",
			stats.MessagesReceived,
			h.messagesReceived+h.messagesFailed)
	}

	if h.messagesReceived != 8 || h.messagesSent != 4 {
		t.Fatalf("end of test. should have handled a diff number of messages (got %d, sent %d)", h.messagesReceived, h.messagesSent)
	}
	if h.messagesFailed != 1 {
		t.Fatal("failed message not done")
	}
}

func SendMessage(t *testing.T, topic string, method string, body []byte) {
	port := 4151
	httpclient := &http.Client{}
	endpoint := fmt.Sprintf("http://127.0.0.1:%d/%s?topic=%s", port, method, topic)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	resp, err := httpclient.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	if resp.StatusCode != 200 {
		t.Fatalf("%s status code: %d", method, resp.StatusCode)
	}
	resp.Body.Close()
}

func TestConsumerHandler(t *testing.T) {
	path := "H:\\0-毕业设计\\godetector\\server\\conf.yaml"
	handler := NewConsumerHandler(common.GetServerConfig(path))
	handler.Connect()
	//图片Base64编码成文本
	for _, topic := range handler.topics {
		data := common.ReadFile("H:\\0-毕业设计\\godetector\\test\\test_img\\neutral.jpg")
		dataStr := base64.StdEncoding.EncodeToString(data)
		req := common.BytesCombine([]byte("{\"topic\":\""), []byte(topic), []byte("\",\"data\":\""), []byte(dataStr), []byte("\"}"))
		SendMessage(t, topic, "pub", req)
	}
	time.Sleep(time.Second * 5) //等待消息处理完毕
}

func TestConsumerTextDetect(t *testing.T) {
	path := "H:\\0-毕业设计\\godetector\\server\\conf.yaml"
	handler := NewConsumerHandler(common.GetServerConfig(path))
	handler.Connect()
	//文本直接传输
	for _, topic := range handler.topics {
		dataStr := "fuck you"
		req := common.BytesCombine([]byte("{\"topic\":\""), []byte(topic), []byte("\",\"data\":\""), []byte(dataStr), []byte("\"}"))
		SendMessage(t, topic, "pub", req)
	}
	time.Sleep(time.Second * 5) //等待消息处理完毕
}
