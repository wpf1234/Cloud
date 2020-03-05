package util

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

func AnalysisToken(token string) (map[string]interface{},error) {
	//解析token
	var myMap=make(map[string]interface{})
	// "http://10.2.15.92:9000/v2/sso/token"
	// http://172.16.17.205:8888/v2/sso/token
	url:="http://10.2.24.169:9000/v2/sso/token"
	//req,err:=http.NewRequest("POST","http://10.2.24.169:8080/v2/sso/token",
	//	strings.NewReader(""))
	req,err:=http.NewRequest("POST",url, strings.NewReader(""))
	if err!=nil{
		log.WithFields(log.Fields{
			"err": err,
		}).Error("请求失败!")

		return nil, err
	}
	req.Header.Add("authorization",token)
	cli:=&http.Client{}
	res,err:=cli.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("服务器错误!")

		return nil, err
	}

	msg, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取响应失败!")

		return nil, err
	}
	json.Unmarshal(msg,&myMap)

	return myMap,nil
}
