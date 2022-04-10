package service

import (
	"github.com/wuyangdut/godetector/common"
	"github.com/wuyangdut/godetector/nsq"
	"log"
	"sync"
)

func (sender *ServiceHandler) NsqImageSend(image []byte) ServiceResponse {
	// ###################### 初始化 ######################
	mutex := sync.Mutex{}
	var wg sync.WaitGroup
	result := ServiceResponse{IsIllegal: false}
	result.Contents = make(map[string]TsResponse)
	imageStr := common.EncodeBase64(image) //image二进制需要Base64编码

	// ###################### NSFW DETECT ######################
	wg.Add(1)
	go func() {
		defer wg.Done()
		//构造请求体
		req := nsq.NsqRequest{}
		req.Topic = "nsfw"
		req.Data = imageStr
		key := common.Sha256(req.Topic + imageStr)
		rawContent := sender.EtcdHandler.Get(key) //先读Etcd防止重复提交
		if rawContent == "" {
			//发送到Nsq队列
			sender.NsqHandler.Publish(req)
			//轮询Etcd，注意超时
			rawContent = sender.EtcdHandler.GetWithLoop(key)
		}
		if rawContent != "" {
			resTemp := TsResponse{IsIllegal: false}
			response := common.Json2Map(rawContent)
			prob := sender.MathHandler.ASTParseNsfw(response)
			if prob == 1 {
				resTemp.IsIllegal = true
			}
			resTemp.RawContent = common.Json2Map(rawContent)
			mutex.Lock()
			result.IsIllegal = result.IsIllegal || resTemp.IsIllegal
			result.Contents[req.Topic] = resTemp
			if resTemp.IsIllegal {
				result.Reason = result.Reason + "It's a pornographic image. "
			}
			mutex.Unlock()
		}
	}()

	// ###################### PROTEST DETECT ######################
	wg.Add(1)
	go func() {
		defer wg.Done()
		//构造请求体
		req := nsq.NsqRequest{}
		req.Topic = "protest"
		req.Data = imageStr
		key := common.Sha256(req.Topic + imageStr)
		rawContent := sender.EtcdHandler.Get(key) //先读Etcd防止重复提交
		if rawContent == "" {
			//发送到Nsq队列
			sender.NsqHandler.Publish(req)
			//轮询Etcd，注意超时
			rawContent = sender.EtcdHandler.GetWithLoop(key)
		}
		if rawContent != "" {
			resTemp := TsResponse{IsIllegal: false}
			response := common.Json2Map(rawContent)
			prob := sender.MathHandler.ASTParseProtest(response)
			if prob == 1 {
				resTemp.IsIllegal = true
			}
			resTemp.RawContent = common.Json2Map(rawContent)
			mutex.Lock()
			result.IsIllegal = result.IsIllegal || resTemp.IsIllegal
			result.Contents[req.Topic] = resTemp
			if resTemp.IsIllegal {
				result.Reason = result.Reason + "It's a protest or violent image. "
			}
			mutex.Unlock()
		}
	}()

	wg.Wait()
	return result
}

//NsfwImageDetect addr为请求的TorchServe地址，req为请求的Data
func NsfwImageDetect(addr string, data []byte) TsResponse {
	result := TsResponse{IsIllegal: false}
	TSNsfwModel := addr + "/predictions/nsfw_model"
	request := common.NewTSRequest(TSNsfwModel, nil, data)
	rawContent := common.GetHttpResponse(request)
	response := common.Json2Map(rawContent)
	log.Println("response of nsfw model: \n", rawContent)
	maxKey, maxVal := "", 0.0
	for key, val := range response {
		if val > maxVal {
			maxVal = val
			maxKey = key
		}
	}
	if maxKey == "porn" || maxKey == "hentai" {
		result.IsIllegal = true
	}
	result.RawContent = common.Json2Map(rawContent)
	return result
}

//NsfwLocalImageDetect addr为请求的TorchServe地址，path为文件地址
func NsfwLocalImageDetect(addr string, path string) TsResponse {
	data := common.ReadFile(path)
	return NsfwImageDetect(addr, data)
}

//ProtestImageDetect addr为请求的TorchServe地址，req为请求的Data
func ProtestImageDetect(addr string, data []byte) TsResponse {
	result := TsResponse{IsIllegal: false}
	TSProtestModel := addr + "/predictions/protest_model"
	request := common.NewTSRequest(TSProtestModel, nil, data)
	rawContent := common.GetHttpResponse(request)
	response := common.Json2Map(rawContent)
	log.Println("response of protest model: \n", rawContent)
	if response["protest"] >= 0.12 {
		for key, val := range response {
			if val >= 0.12 && key != "protest" {
				result.IsIllegal = true
			}
		}
	}
	result.RawContent = common.Json2Map(rawContent)
	return result
}

//ProtestLocalImageDetect addr为请求的TorchServe地址，path为文件地址
func ProtestLocalImageDetect(addr string, path string) TsResponse {
	data := common.ReadFile(path)
	return ProtestImageDetect(addr, data)
}
