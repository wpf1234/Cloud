package handler

import (
	"SoftwareCloud/conf"
	log "github.com/sirupsen/logrus"
	"time"
)

var(
	class []string   //classify
	de []string      //deliver
	st []string      //service type
	department []string  //dept
)

func sliceClear(s interface{}) bool{
	if ps,ok:=s.(*[]string);ok{
		*ps=(*ps)[0:0]
	}else if ps,ok:=s.(*[]int);ok{
		*ps=(*ps)[0:0]
	}else if ps,ok:=s.(*[]float64);ok{
		*ps=(*ps)[0:0]
	}else{
		log.Info("不支持的数据类型!")
		return false
	}
	return true
}

func GetArray(){
	for {
		sliceClear(&class)
		sliceClear(&de)
		sliceClear(&st)
		sliceClear(&department)

		//service_classify
		//service_page
		db:=conf.DB.Raw("select name from service_classify")
		rows,err:=db.Rows()
		if err!=nil{
			log.Error("获取数据失败: ",err)
			return
		}

		for rows.Next(){
			var c string
			rows.Scan(&c)

			class=append(class,c)
		}


		db=conf.DB.Raw("select deliver from service_deliver")
		rows,err=db.Rows()
		if err!=nil{
			log.Error("获取数据失败: ",err)
			return
		}

		for rows.Next(){
			var d string
			rows.Scan(&d)

			de=append(de,d)
		}

		db=conf.DB.Raw("select type from service_type")
		rows,err=db.Rows()
		if err!=nil{
			log.Error("获取数据失败: ",err)
			return
		}

		for rows.Next(){
			var ty string
			rows.Scan(&ty)

			st=append(st,ty)
		}

		db=conf.DB.Raw("select distinct department from service_app")
		rows,err=db.Rows()
		if err!=nil{
			log.Error("获取数据失败: ",err)
			return
		}

		for rows.Next(){
			var dept string
			rows.Scan(&dept)

			department=append(department,dept)
		}

		rows.Close()

		log.Info("数据获取成功!")

		ticker:=time.NewTicker(12*time.Hour)
		<-ticker.C
	}
}

//func GetPubStatus(token,WFID,business string) {
//	for {
//		var operation, remark interface{}
//		//wf := workflow.NewWorkflowClient()
//		//token, err := wf.GetToken("180435", "wpf!2345")
//		//if err != nil {
//		//	log.Error("获取 token 失败: ", err)
//		//
//		//	return
//		//}
//		//WFID:=strconv.Itoa(conf.Sc.PubWFID)
//		////business:=PubSerial
//		//business:="45733088"
//		url := "http://10.2.44.61/wfapi/api/Business?wfID=" + WFID + "&businessID=" + business
//		request, err := http.NewRequest("GET", url, strings.NewReader(""))
//		if err != nil {
//			log.Error("请求失败!")
//
//			return
//		}
//		tt := time.Now().Unix()
//		vc := security.ShaString(fmt.Sprintf(`%s_%s_%d`,
//			conf.Sc.Appid, conf.Sc.Appkey, tt))
//
//		request.Header.Add("Content-Type", "application/json")
//		request.Header.Add("appID", conf.Sc.Appid)
//		request.Header.Add("appKey", conf.Sc.Appkey)
//		request.Header.Add("Authorization", conf.Sc.Appid+`:`+vc)
//		request.Header.Add("token", token)
//
//		client := &http.Client{}
//		response, err := client.Do(request)
//		if err != nil {
//			log.Error("服务器错误: ", err)
//
//			return
//		}
//
//		data, err := ioutil.ReadAll(response.Body)
//		if err != nil {
//			log.Error("获取响应失败: ", err)
//
//			return
//		}
//		result := make(map[string]interface{})
//
//		json.Unmarshal(data, &result)
//		log.Info("123456: ", result)
//
//		for k, v := range result {
//			if k == "Operations" {
//				data := v.([]interface{})
//				opt:=data[len(data)-1].(map[string]interface{})
//				operation=opt["Operation"].(string)
//				remark=opt["Remark"]
//			}
//		}
//
//
//
//		log.Info("审批结果: ", operation)
//		log.Info("审批意见: ", remark)
//
//		ticker:=time.NewTicker(2*time.Hour)
//		<-ticker.C
//	}
//}
//
//func GetUseStatus(token,WFID,business string) {
//	for {
//		var operation, remark interface{}
//		//wf := workflow.NewWorkflowClient()
//		//token, err := wf.GetToken("180435", "wpf!2345")
//		//if err != nil {
//		//	log.Error("获取 token 失败: ", err)
//		//
//		//	return
//		//}
//		//WFID:=strconv.Itoa(conf.Sc.PubWFID)
//		//business:=UseSerial
//		url := "http://10.2.44.61/wfapi/api/Business?wfID=" + WFID + "&businessID=" + business
//		request, err := http.NewRequest("GET", url, strings.NewReader(""))
//		if err != nil {
//			log.Error("请求失败!")
//
//			return
//		}
//		tt := time.Now().Unix()
//		vc := security.ShaString(fmt.Sprintf(`%s_%s_%d`,
//			conf.Sc.Appid, conf.Sc.Appkey, tt))
//
//		request.Header.Add("Content-Type", "application/json")
//		request.Header.Add("appID", conf.Sc.Appid)
//		request.Header.Add("appKey", conf.Sc.Appkey)
//		request.Header.Add("Authorization", conf.Sc.Appid+`:`+vc)
//		request.Header.Add("token", token)
//
//		client := &http.Client{}
//		response, err := client.Do(request)
//		if err != nil {
//			log.Error("服务器错误: ", err)
//
//			return
//		}
//
//		data, err := ioutil.ReadAll(response.Body)
//		if err != nil {
//			log.Error("获取响应失败: ", err)
//
//			return
//		}
//		result := make(map[string]interface{})
//
//		json.Unmarshal(data, &result)
//		log.Info("123456: ", result)
//
//		for k, v := range result {
//			if k == "Operations" {
//				data := v.([]interface{})
//				opt:=data[len(data)-1].(map[string]interface{})
//				operation=opt["Operation"].(string)
//				remark=opt["Remark"]
//			}
//		}
//
//		log.Info("审批结果: ", operation)
//		log.Info("审批意见: ", remark)
//
//		ticker:=time.NewTicker(2*time.Hour)
//		<-ticker.C
//	}
//}

