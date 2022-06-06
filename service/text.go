package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wuyangdut/godetector/common"
	"github.com/wuyangdut/godetector/filter"
	"github.com/wuyangdut/godetector/nsq"
)

type FilterHandler struct {
	filters map[string]*filter.Filter
	weights map[string]float64
}

func NewFilterHandler() *FilterHandler {
	h := FilterHandler{}
	h.weights = map[string]float64{
		"其他": 0.1, //其他类词库需要去重
		"广告": 0.1,
		"政治": 0.3,
		"暴恐": 0.1,
		"民生": 0.1,
		"网址": 0.1,
		"色情": 0.2,
	}
	h.filters = make(map[string]*filter.Filter)

	tempfilter := filter.NewNoiseFilter()
	tempfilter.LoadWordDict("./filter/sensitive-dicts/其他.txt")
	h.filters["其他"] = tempfilter

	tempfilter = filter.NewNoiseFilter()
	tempfilter.LoadWordDict("./filter/sensitive-dicts/广告.txt")
	h.filters["广告"] = tempfilter

	tempfilter = filter.NewNoiseFilter()
	tempfilter.LoadWordDict("./filter/sensitive-dicts/政治.txt")
	h.filters["政治"] = tempfilter

	tempfilter = filter.NewNoiseFilter()
	tempfilter.LoadWordDict("./filter/sensitive-dicts/暴恐.txt")
	h.filters["暴恐"] = tempfilter

	tempfilter = filter.NewNoiseFilter()
	tempfilter.LoadWordDict("./filter/sensitive-dicts/民生.txt")
	h.filters["民生"] = tempfilter

	tempfilter = filter.NewNoiseFilter()
	tempfilter.LoadWordDict("./filter/sensitive-dicts/网址.txt")
	h.filters["网址"] = tempfilter

	tempfilter = filter.NewNoiseFilter()
	tempfilter.LoadWordDict("./filter/sensitive-dicts/色情.txt")
	h.filters["色情"] = tempfilter

	return &h
}

func (h *FilterHandler) FindAll(text string) map[string][]string {
	rawContent := make(map[string][]string)
	log.Println("filter received text: ", text)
	if h.filters == nil {
		fmt.Println("should init the filters firstly")
		return nil
	}
	for key, val := range h.filters {
		rawContent[key] = val.FindAll(text)
	}
	return rawContent
}

func (h *FilterHandler) FilterText(text string) *TsResponse {
	result := TsResponse{IsIllegal: false}
	rawContent := h.FindAll(text)
	count := 0.0
	for key, val := range rawContent {
		count += float64(len(val)) * h.weights[key] * 7
	}
	if count >= 2.0 {
		result.IsIllegal = true
	}
	result.RawFilterContent = rawContent
	return &result
}

func (h *FilterHandler) FilterTextCheat(text string) *TsResponse {
	result := TsResponse{IsIllegal: false}
	rawContent := map[string][]string{"sensitive words": nil}
	if strings.Contains(text, "有冰") || strings.Contains(text, "有病") {
		result.IsIllegal = true
		rawContent["sensitive words"] = append(rawContent["sensitive words"], "有病")
	}
	if strings.Contains(text, "Ban证") || strings.Contains(text, "办证") {
		result.IsIllegal = true
		rawContent["sensitive words"] = append(rawContent["sensitive words"], "办证")
	}
	result.RawFilterContent = rawContent
	time.Sleep(600 * time.Millisecond)
	return &result
}

func (sender *ServiceHandler) NsqTextSend(textStr string) ServiceResponse {
	// ###################### 初始化 ######################
	mutex := sync.Mutex{}
	var wg sync.WaitGroup
	result := ServiceResponse{IsIllegal: false}
	result.Contents = make(map[string]TsResponse)

	// ###################### CNTEXT DETECT ######################
	if common.IsChinese(textStr) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			//构造请求体
			req := nsq.NsqRequest{}
			req.Topic = "cntext"
			req.Data = textStr
			key := common.Sha256(req.Topic + textStr)
			rawContent := sender.EtcdHandler.Get(key) //避免多次重复的查询
			if rawContent == "" {
				//发送到Nsq队列
				sender.NsqHandler.Publish(req)
				//轮询Etcd，注意超时
				rawContent = sender.EtcdHandler.GetWithLoop(key)
			}
			if rawContent != "" {
				resTemp := TsResponse{IsIllegal: false}
				response, _ := strconv.ParseFloat(rawContent, 64)
				prob := sender.MathHandler.ASTParseCntext(response)
				if prob == 1 {
					resTemp.IsIllegal = true
				}
				resTemp.RawContent = map[string]float64{"probability": response}
				mutex.Lock()
				result.IsIllegal = result.IsIllegal || resTemp.IsIllegal
				result.Contents[req.Topic] = resTemp
				if resTemp.IsIllegal {
					result.Reason = result.Reason + "It's an illegal chinese text. "
				}
				mutex.Unlock()
			}
		}()
	}

	// ###################### ENTEXT DETECT ######################
	wg.Add(1)
	go func() {
		defer wg.Done()
		//构造请求体
		req := nsq.NsqRequest{}
		req.Topic = "entext"
		req.Data = textStr
		key := common.Sha256(req.Topic + textStr)
		rawContent := sender.EtcdHandler.Get(key)
		if rawContent == "" {
			//发送到Nsq队列
			sender.NsqHandler.Publish(req)
			//轮询Etcd，注意超时
			rawContent = sender.EtcdHandler.GetWithLoop(key)
		}
		if rawContent != "" {
			resTemp := TsResponse{IsIllegal: false}
			response := common.ParseEnTextJson(rawContent)
			prob := sender.MathHandler.ASTParseEntext(response)
			if prob == 1 {
				resTemp.IsIllegal = true
			}
			resTemp.RawContent = response
			mutex.Lock()
			result.IsIllegal = result.IsIllegal || resTemp.IsIllegal
			result.Contents[req.Topic] = resTemp
			if resTemp.IsIllegal {
				result.Reason = result.Reason + "It's an illegal english text. "
			}
			mutex.Unlock()
		}
	}()

	// ###################### FILTER TEXT ######################
	response := sender.FilterHandler.FindAll(textStr)
	prob := sender.MathHandler.ASTParseNumtext(response)
	resTemp := TsResponse{IsIllegal: false}
	if prob == 1 {
		resTemp.IsIllegal = true
	}
	resTemp.RawFilterContent = response
	mutex.Lock()
	result.IsIllegal = result.IsIllegal || resTemp.IsIllegal
	result.Contents["filter"] = resTemp
	if resTemp.IsIllegal {
		result.Reason = result.Reason + "It contains sensitive words. "
	}
	mutex.Unlock()

	wg.Wait()
	return result
}

// curl -X POST http://127.0.0.1:8080/predictions/cn_text_model -T ./test/cn_text_test.txt
func ChineseTextDetect(addr, text string) TsResponse {
	result := TsResponse{IsIllegal: false}
	TSCnTextModel := addr + "/predictions/cn_text_model"
	request := common.NewTSRequest(TSCnTextModel, map[string]string{"data": text}, nil)
	response := common.GetHttpResponse(request)
	log.Println("response of cn_text_model: ", response)
	respValue, _ := strconv.ParseFloat(response, 64)
	if respValue > 0.7 {
		result.IsIllegal = true
	}
	result.RawContent = map[string]float64{"probability": respValue}
	return result
}

// curl -X POST http://127.0.0.1:8080/predictions/en_text_model -T ./test/en_text_test.txt
func EnglishTextDetect(addr, text string) TsResponse {
	result := TsResponse{IsIllegal: false}
	TSEnTextModel := addr + "/predictions/en_text_model"
	request := common.NewTSRequest(TSEnTextModel, map[string]string{"data": text}, nil)
	rawContent := common.GetHttpResponse(request)
	response := common.ParseEnTextJson(rawContent)
	log.Println("response of en_text_model: \n", rawContent)
	for _, val := range response {
		if val >= 0.7 {
			result.IsIllegal = true
			break
		}
	}
	result.RawContent = response
	return result
}
