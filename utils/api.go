package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func api() {
	address := "http://118.190.202.4:8878/bioindex/findBioIndexWithSurvey" //查询生物年龄指标详情及调查信息
	// url := "http://118.190.202.4:8878/bioindex/getBioIndexNameList" //获取生物年龄指标名称列表
	// url := "http://118.190.202.4:8878/user/findUserByToken"	// 获取用户信息
	// url := "http://118.190.202.4:8878/biosurvey/getBioSurveyDateListByToken" //获取有生物年龄调查记录的日期列表

	formValues := url.Values{}
	formValues.Set("bioIndexId", "36ce6ed6-2087-4f55-a42c-fd9728f4d04d")
	formDataStr := formValues.Encode()
	formDataBytes := []byte(formDataStr)
	formBytesReader := bytes.NewReader(formDataBytes)

	request, _ := http.NewRequest("POST", address, formBytesReader)
	request.Header.Add("x-token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MCwiVXNlcklkIjoiMjFiMTc2MjUtZDBiNi00MzBmLWIxNDItZmQ2NjA1MGYyMjQ1IiwiVXNlcm5hbWUiOiJzdXBlcmFkbWluIiwiTmlja25hbWUiOiLotoXnuqfnrqHnkIblkZgiLCJCdWZmZXJUaW1lIjo4NjQwMCwiZXhwIjoxNjUxNzEzOTQzLCJpc3MiOiJxbVBsdXMiLCJuYmYiOjE2NTExMDgxNDN9.wcsfYT6idF8zHqxTNJuo5fQOT1FeZ1CsklSXrwTmehY")

	client := &http.Client{}
	response, err := client.Do(request)
	jsondata, err := ioutil.ReadAll(response.Body)
	fmt.Println(string(jsondata))
	fmt.Println(err)
}
