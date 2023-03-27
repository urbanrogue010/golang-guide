package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/resty.v1"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

func main() {
	type videoResp struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		MessageCn string `json:"message_cn"`
		Data      struct {
			VideoId string `json:"video_id"`
		} `json:"data"`
	}

	var (
		openApiUrlPrefix string = "https://ad.oceanengine.com/open_api/2/"
		uri              string = "file/video/ad/"
		// 请求Header
		contentType string = "multipart/form-data"
		accessToken string = "e88f206ab28a97ef494b853982d81739b81a1e37"
		//XDebugMode  int = 1
		// 请求参数
		advertiserId   int64  = 1760312309087432 // 广告主ID
		uploadType     string = "UPLOAD_BY_FILE" // 视频上传方式，可选值:UPLOAD_BY_FILE: 文件上传（默认值），UPLOAD_BY_URL: 网址上传
		videoSignature string = "6b12a8bbbe8e69a2ef5929028b0b50c3"
		filename       string = "auto4_15151.aaaaaaa_test环境slicess__V_ZJR_ZJR_en+de_1X1_31s"                  // 素材的文件名，可自定义素材名，不传择默认取文件名，最长255个字符。UPLOAD_BY_URL必填  注：若同一素材已进行上传，重新上传不会改名。
		videoUrl       string = "https://ark-oss.bettagames.com/2023-03/6b12a8bbbe8e69a2ef5929028b0b50c3.mp4" // 视频url地址
		//
		ttVideoResp videoResp
	)
	url := fmt.Sprintf("%s%s", openApiUrlPrefix, uri)
	fileBytes, err := getFileBytes2(videoUrl)
	if err != nil {
		fmt.Println("getFileBytes err", err)
		return
	}

	httpStatus, resp, err := HttpPostMultipart(url,
		map[string]string{
			"advertiser_id":   fmt.Sprintf("%d", advertiserId),
			"upload_type":     uploadType,
			"video_signature": videoSignature,
			"filename":        filename,
		},
		map[string][]byte{
			"video_file": fileBytes,
		},
		map[string]string{
			"Content-Type": contentType,
			"Access-Token": accessToken,
		})
	if err != nil {
		fmt.Println("Post err", err)
		return
	}
	fmt.Println("code:", httpStatus)

	err = json.Unmarshal(resp, &ttVideoResp)
	if err != nil {
		fmt.Println("Unmarshal err:", err)
		return
	}
	if httpStatus != http.StatusOK {
		fmt.Println("resp.StatusCode() != http.StatusOK")
		return
	}
	fmt.Println("ttVideoResp: ", ttVideoResp)
}

//HttpPostMultipart 通过multipart/form-data请求数据
func HttpPostMultipart(url string, formData map[string]string, fileData map[string][]byte,
	header map[string]string) (httpStatus int, resp []byte, err error) {

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	for k, v := range formData {
		_ = writer.WriteField(k, v)
	}
	for k, v := range fileData {
		fileWriter, err := writer.CreateFormField(k)
		if err != nil {
			return 0, nil, err
		}
		_, err = io.Copy(fileWriter, bytes.NewReader(v))
		if err != nil {
			return 0, nil, err
		}
	}
	err = writer.Close()
	if err != nil {
		return 0, nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return 0, nil, err
	}
	for k, v := range header {
		req.Header.Add(k, v)
	}
	if len(fileData) > 0 {
		req.Header.Set("Content-Type", writer.FormDataContentType())
	}
	response, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	resp, err = ioutil.ReadAll(response.Body)
	return response.StatusCode, resp, err
}

func getFileBytes2(netUrl string) ([]byte, error) {
	resp, err := resty.New().R().Get(netUrl)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}
