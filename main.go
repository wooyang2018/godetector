package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/wuyangdut/godetector/common"
	"github.com/wuyangdut/godetector/service"
)

var config *common.WebConfig
var parsedTemplate *template.Template
var adminTemplate *template.Template
var filterHandler *service.FilterHandler
var serviceHandler *service.ServiceHandler
var conf = flag.String("f", "./conf.yaml", "Input Your Config Yaml Path")
var testTsAddr = flag.String("addr", "http://127.0.0.1:8080", "Input The TorchServe Address For Test")

func TestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	result := service.TsResponse{IsIllegal: false}
	err := parsedTemplate.Execute(w, result)
	if err != nil {
		log.Println("Error executing template :", err)
	}
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	err := adminTemplate.Execute(w, nil)
	if err != nil {
		log.Println("Error executing template :", err)
	}
}

// 翻译：curl http://127.0.0.1:8080/predictions/nsfw_model -T ./test/test_img/porn.jpg
func NsfwImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	data := common.ParseFileBytes(r)
	result := service.NsfwImageDetect(*testTsAddr, data)
	err := parsedTemplate.Execute(w, result.MarshalIndent())
	if err != nil {
		log.Println("Error executing template :", err)
	}
}

//翻译; curl http://127.0.0.1:8080/predictions/protest_model -T ./test/test_img/protest_sign.jpg
func ProtestImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	data := common.ParseFileBytes(r)
	result := service.ProtestImageDetect(*testTsAddr, data)
	err := parsedTemplate.Execute(w, result.MarshalIndent())
	if err != nil {
		log.Println("Error executing template :", err)
	}
}

func FilterTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	result := filterHandler.FilterTextCheat(r.PostFormValue("text"))
	err := parsedTemplate.Execute(w, result.MarshalIndent())
	if err != nil {
		log.Println("Error executing template :", err)
	}
}

// curl -X POST http://127.0.0.1:8080/predictions/cn_text_model -T ./test/cn_text_test.txt
func ChineseTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	result := service.ChineseTextDetect(*testTsAddr, r.PostFormValue("text"))
	err := parsedTemplate.Execute(w, result.MarshalIndent())
	if err != nil {
		log.Println("Error executing template :", err)
	}
}

// curl -X POST http://127.0.0.1:8080/predictions/en_text_model -T ./test/en_text_test.txt
func EnglishTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	result := service.EnglishTextDetect(*testTsAddr, r.PostFormValue("text"))
	err := parsedTemplate.Execute(w, result.MarshalIndent())
	if err != nil {
		log.Println("Error executing template :", err)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	result := service.CheckHealth(*testTsAddr)
	_, err := fmt.Fprint(w, result)
	if err != nil {
		log.Println("Error checking health :", err)
	}
}

func NsqImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	image := common.ParseFileBytes(r)
	serviceResponse := serviceHandler.NsqImageSend(image)
	resBody := serviceResponse.MarshalIndent()
	fmt.Fprint(w, resBody)
}

func NsqTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	text := r.PostFormValue("text")
	serviceResponse := serviceHandler.NsqTextSend(text)
	resBody := serviceResponse.MarshalIndent()
	fmt.Fprint(w, resBody)
}

func StrategyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "405", http.StatusMethodNotAllowed)
	}
	raw := make(map[string]string)
	raw["nsfw"] = r.PostFormValue("nsfw")
	raw["protest"] = r.PostFormValue("protest")
	raw["cntext"] = r.PostFormValue("cntext")
	raw["entext"] = r.PostFormValue("entext")
	raw["numtext"] = r.PostFormValue("numtext")
	serviceHandler.MathHandler = service.NewMathAnalyzer(raw)
	_, err := fmt.Fprint(w, "Updating strategy is finished.")
	if err != nil {
		log.Println("Error updating strategy :", err)
	}
}

func main() {
	flag.Parse()
	fmt.Println("hello, godetector web server!")
	config = common.GetWebConfig(*conf)

	//监听Http请求，用于管理员测试本地推理端后台
	http.HandleFunc("/index/test", TestHandler) //测试主页
	http.HandleFunc("/image/nsfw", NsfwImageHandler)
	http.HandleFunc("/image/protest", ProtestImageHandler)
	http.HandleFunc("/text/filter", FilterTextHandler)
	http.HandleFunc("/text/cn", ChineseTextHandler)
	http.HandleFunc("/text/en", EnglishTextHandler)
	http.HandleFunc("/test/health", HealthCheckHandler)

	//监听Http请求，用于管理员测试违规图片检测、违规文本检测、自定义阻止策略
	http.HandleFunc("/index/admin", AdminHandler) //Admin主页
	http.HandleFunc("/nsq/image", NsqImageHandler)
	http.HandleFunc("/nsq/text", NsqTextHandler)
	http.HandleFunc("/admin/strategy", StrategyHandler)

	parsedTemplate, _ = template.ParseFiles("./template/test.html")
	adminTemplate, _ = template.ParseFiles("./template/admin.html")
	filterHandler = service.NewFilterHandler()
	serviceHandler = service.NewServiceHandler(config.NsqdAddr, config.EtcdAddr, config.Strategy)
	defer serviceHandler.Stop()

	fmt.Printf("admin index serve on: http://%s/index/admin\n", config.WebAddr)
	fmt.Printf("test index serve on: http://%s/index/test\n", config.WebAddr)
	err := http.ListenAndServe(config.WebAddr, nil)
	if err != nil {
		log.Fatal("Error Starting the HTTP Server :", err)
		return
	}
}
