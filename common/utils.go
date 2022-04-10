package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func Json2Map(b string) map[string]float64 {
	res := make(map[string]float64)
	err := json.Unmarshal([]byte(b), &res)
	if err != nil {
		fmt.Printf("Unmarshal with error: %+v\n", err)
	}
	return res
}

func Map2Json(m map[string][]string) string {
	res, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("Marshal with error: %+v\n", err)
	}
	return string(res)
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func ReadFile(path string) []byte {
	if path == "" {
		fmt.Printf("file path cannot be empty")
		return nil
	}
	if !PathExists(path) {
		fmt.Printf("file path is not exist")
		return nil
	}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %+v\n", err)
		return nil
	}
	return file
}

func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func WriteFile(path string, data []byte) {
	if path == "" {
		fmt.Printf("file path cannot be empty")
		return
	}
	err := ioutil.WriteFile(path, data, 0777)
	// handle this error
	if err != nil {
		fmt.Printf("Error writing file: %+v\n", err)
	}
}

func GetHostName() string {
	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error getting host name: %+v\n", err)
		return ""
	}
	return name
}

//https://stackoverflow.com/questions/63593441/how-to-decode-base64-encoded-json-in-go
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeBase64(str string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Printf("Error decoding base64 string: %+v\n", err)
	}
	return data, err
}

func Sha256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

//获取当前路径
func GetCurrentPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func FormReader(reader io.Reader) string {
	// 将io.Reader转换成string
	buf := new(strings.Builder)
	_, copyError := io.Copy(buf, reader)
	if copyError != nil {
		fmt.Printf("Error copying, %+v\n", copyError)
	}
	return buf.String()
}

//IsChinese 判断str是否是中文
func IsChinese(str string) bool {
	var count int
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			count++
			break
		}
	}
	return count > 0
}

//构造Http请求，支持自定义json参数和携带文件，用于文本检测和图片检测
func NewTSRequest(uri string, params map[string]string, data []byte) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	//判断data的类型
	if data != nil && len(data) != 0 {
		part, err := writer.CreateFormField("data")
		if err != nil {
			fmt.Printf("Error creating form file, %+v\n", err)
		}
		_, err = io.Copy(part, bytes.NewReader(data))
		if err != nil {
			fmt.Printf("Error copying, %+v\n", err)
		}
	}
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	writer.Close()
	request, err := http.NewRequest("POST", uri, body)
	if err != nil {
		fmt.Printf("Error new http request, %+v\n", err)
	}
	// request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Connection", "Keep-Alive")
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

//newFileTSRequest 构造Http请求携带文件，用于图片检测
func NewFileTSRequest(uri string, path string) *http.Request {
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)
	if path != "" {
		file, err := os.Open(path)
		defer file.Close()
		if err != nil {
			fmt.Printf("Error open file, %+v\n", err)
		} else {
			part, err := writer.CreateFormFile("data", path)
			if err != nil {
				fmt.Printf("Error creating form file, %+v\n", err)
			} else {
				_, err = io.Copy(part, file)
				if err != nil {
					fmt.Printf("Error copying, %+v\n", err)
				}
			}
		}
	}
	writer.Close() //该语句的位置很关键，否则会造成请求头中的Content-Length
	request, err := http.NewRequest("POST", uri, &body)
	if err != nil {
		fmt.Printf("Error new http request, %+v\n", err)
	}
	// request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Connection", "Keep-Alive")
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func GetHttpResponse(req *http.Request) string {
	// 关闭remote
	trans := http.Transport{
		DisableKeepAlives: true,
	}
	client := http.Client{
		Transport: &trans,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error doing request, %+v\n", err)
		return ""
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return string(b)
}

func ParseEnTextJson(rawContent string) map[string]float64 {
	response := make(map[string]float64)
	rawContent = rawContent[1 : len(rawContent)-1]
	rawContent = strings.Replace(rawContent, "'", "", -1)
	contents := strings.Split(rawContent, ",")
	for _, content := range contents {
		temp := strings.Split(content, ":")
		tempvalue, _ := strconv.ParseFloat(strings.TrimSpace(temp[1]), 64)
		response[strings.TrimSpace(temp[0])] = tempvalue
	}
	return response
}

//SaveFile 上传文件到服务端local，用于图片检测
func SaveFile(r *http.Request) string {
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Printf("Error Getting File, %+v\n", err)
	}
	defer file.Close()

	out, pathError := ioutil.TempFile("upload-images", "upload-*.png")
	if pathError != nil {
		fmt.Printf("Error Creating a file for writing, %+v\n", pathError)
	}
	defer out.Close()

	_, copyError := io.Copy(out, file)
	if copyError != nil {
		fmt.Printf("Error copying, %+v\n", copyError)
	}
	// fmt.Fprintln(w, "File Uploaded Successfully! ")
	// fmt.Fprintln(w, "Name of the File: ", header.Filename)
	// fmt.Fprintln(w, "Size of the File: ", header.Size)
	return out.Name()
}

//ParseFileBytes 将请求中的文件转换成字节流，用于图片检测
func ParseFileBytes(r *http.Request) []byte {
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Printf("Error Getting File, %+v\n", err)
	}
	defer file.Close()
	buf := bytes.Buffer{}
	_, copyError := io.Copy(&buf, file)
	if copyError != nil {
		fmt.Printf("Error copying, %+v\n", copyError)
	}
	return buf.Bytes()
}
