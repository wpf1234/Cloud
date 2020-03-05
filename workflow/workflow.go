package workflow

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gree/common/security"

	"github.com/alecthomas/log4go"
	"github.com/bitly/go-simplejson"
)

type Workflow struct{}

func workflowClient(method, uri, token, body string) (string, error) {
	req, err := http.NewRequest(method, uri, strings.NewReader(body))
	if err != nil {
		_ = log4go.Error(err.Error())
		return "", err
	}

	appid := "0124fdf8-9b7c-49ef-a7c6-e606dba474b9"
	appkey := "0c5e557b-3a5a-49b5-a7dc-fd83df2e639b"
	t := time.Now().Unix()
	vc := security.ShaString(fmt.Sprintf(`%s_%s_%d`, appid, appkey, t))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("appId", appid)
	req.Header.Set("appKey", appkey)
	req.Header.Set("Authorization", appid+`:`+vc)
	req.Header.Set("token", token)

	c := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := c.Do(req)
	if err != nil {
		_ = log4go.Error(err.Error())
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(`exception status code `, resp.StatusCode)
		r, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			_ = log4go.Error(err.Error())
			return "", err
		}

		return "", errors.New(fmt.Sprintf(`{"code":%d, "msg":"%s"}`, resp.StatusCode, string(r)))
	}
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_ = log4go.Error(err.Error())
		return "", err
	}

	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		_ = log4go.Error(err.Error())
		return "", err
	}

	if err := resp.Body.Close(); err != nil {
		_ = log4go.Error(err.Error())
		return "", err
	}

	return string(r), nil
}

func NewWorkflowClient() *Workflow {
	return new(Workflow)
}

func (w *Workflow) GetToken(username, passwd string) (string, error) {
	r, err := workflowClient(http.MethodPost, `http://10.2.44.61/wfApi/api/login`, ``, fmt.Sprintf(`{"UserID":"%s","Password":"%s"}`, username, passwd))
	if err != nil {
		return "", err
	}
	obj, err := simplejson.NewJson([]byte(r))
	if err != nil {
		return "", err
	}

	return obj.Get("Token").MustString(), nil
}

func (w *Workflow) ActiveWorkflow(WFID, WFNodeOperationID, mainData, subDatas, token string) error {
	body := fmt.Sprintf(`{ "WFID":"%s","WFNodeOperationID":"%s","MainData":%s, "SubDatas":%s}`, WFID, WFNodeOperationID, mainData, subDatas)

	//fmt.Println("req:" + body)
	r, err := workflowClient(http.MethodPost, "http://10.2.44.61/wfApi/api/Business", token, body)
	if err != nil {
		_ = log4go.Error(err)
		return err
	}
	//{"JsonCode":0,"JsonMessage":"","Result":"43804778"}
	//fmt.Println(r)
	obj, err := simplejson.NewJson([]byte(r))
	if err != nil {
		_ = log4go.Error(err)
		return err
	}
	jc := obj.Get("JsonCode").MustInt()
	jm := obj.Get("JsonMessage").MustString()

	if jc > 0 && len(jm) > 0 {
		return errors.New(jm)
	}

	return nil
}
