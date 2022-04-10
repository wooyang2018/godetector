package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wuyangdut/godetector/common"
	"github.com/wuyangdut/godetector/nsq"
	"net/http"
	"strconv"
	"strings"
)

type TsResponse struct {
	IsIllegal        bool
	RawContent       map[string]float64
	RawFilterContent map[string][]string
}

func (t TsResponse) MarshalIndent() string {
	body, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		fmt.Println("Error Marshal Indent, ", err)
		return ""
	}
	return string(body)
}

type ServiceResponse struct {
	IsIllegal bool
	Contents  map[string]TsResponse
	Reason    string
}

func (s ServiceResponse) MarshalIndent() string {
	body, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		fmt.Println("Error Marshal Indent, ", err)
		return ""
	}
	return string(body)
}

type ServiceHandler struct {
	NsqHandler    *nsq.ProducerHandler
	EtcdHandler   *nsq.EtcdHandler
	MathHandler   *MathAnalyzer
	FilterHandler *FilterHandler
}

func NewServiceHandler(nsqAddr, etcdAddr string, raw map[string]string) *ServiceHandler {
	sender := ServiceHandler{}
	sender.NsqHandler = nsq.NewProducerHandler(nsqAddr)
	sender.EtcdHandler = nsq.NewEtcdHandler(etcdAddr)
	sender.MathHandler = NewMathAnalyzer(raw)
	sender.FilterHandler = NewFilterHandler()
	return &sender
}

func (sender *ServiceHandler) Stop() {
	sender.NsqHandler.Stop()
	sender.EtcdHandler.Stop()
}

func CheckHealth(tsAddr string) string {
	models := []string{"protest_model", "nsfw_model", "en_text_model", "cn_text_model"}
	addr := strings.Split(tsAddr, ":")
	port, _ := strconv.ParseInt(addr[2], 10, 32)
	tsAddr = addr[0] + ":" + addr[1] + ":" + strconv.Itoa(int(port)+1)
	buffer := bytes.Buffer{}
	for _, model := range models {
		tempUri := tsAddr + "/models/" + model + "/all"
		req, _ := http.NewRequest("GET", tempUri, nil)
		response := common.GetHttpResponse(req)
		fmt.Fprintf(&buffer, "%s\n", response)
	}
	return buffer.String()
}
