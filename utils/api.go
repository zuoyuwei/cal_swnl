package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func api() {
	// url := "http://118.190.202.4:8878/bioindex/findBioIndexWithSurvey" //查询生物年龄指标详情及调查信息
	// url := "http://118.190.202.4:8878/bioindex/getBioIndexNameList" //获取生物年龄指标名称列表
	// url := "http://118.190.202.4:8878/user/findUserByToken"	// 获取用户信息
	url := "http://118.190.202.4:8878/biosurvey/getBioSurveyDateListByToken" //获取有生物年龄调查记录的日期列表
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("x-token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MCwiVXNlcklkIjoiNGJiMWEzMDUtYmQ2OS00ZmE3LWFmNWEtOWYyOTk4MjViOWNkIiwiVXNlcm5hbWUiOiIxNTA5ODkzMzYyMSIsIk5pY2tuYW1lIjoi5b6u5L-h55So5oi3IiwiQnVmZmVyVGltZSI6ODY0MDAsImV4cCI6MTY1MTU1OTgzMSwiaXNzIjoicW1QbHVzIiwibmJmIjoxNjUwOTU0MDMxfQ.l3vtnNvnLgJ6REeQEWkkjCrIw6fL75NeSAn9WTcF_24")

	client := &http.Client{}
	response, err := client.Do(request)
	jsondata, err := ioutil.ReadAll(response.Body)
	fmt.Println(string(jsondata))
	fmt.Println(err)
}
