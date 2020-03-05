package handler

import (
	"SoftwareCloud/conf"
	"SoftwareCloud/util"
	"SoftwareCloud/workflow"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/broker"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	h "SoftwareCloud/proto/home"

	"SoftwareCloud/proto/ad"

	a "SoftwareCloud/proto/apply"

	m "SoftwareCloud/proto/manage"

	d "SoftwareCloud/proto/detail"

	lm "SoftwareCloud/proto/message"

	log "github.com/sirupsen/logrus"
)

type Gin struct {
}

type delReq struct {
	Id    int      `json:"id"`
	File  []string `json:"file"`
	Image []string `json:"image"`
	Video []string `json:"video"`
}

type home struct {
	User       string      `json:"user"`
	Classify   []string    `json:"classify"`
	Banner     interface{} `json:"banner"`
	NewService interface{} `json:"newService"`
	Rank       interface{} `json:"rank"`
	Service    interface{} `json:"service"`
}

type adDetail struct {
	Id       int         `json:"id"`
	Position string      `json:"position"`
	Remark   string      `json:"remark"`
	Content  interface{} `json:"content"`
}

type header struct {
	Classify []string `json:"classify"`
	Dept     []string `json:"dept"`
	//Deliver []string `json:"deliver"`
	//ServiceType []string `json:"serviceType"`
}

type BoardHeader struct {
	PubRatio   interface{} `json:"pubRatio"`
	PubTotal   int         `json:"pubTotal"`
	ApplyRatio interface{} `json:"applyRatio"`
	ApplyTotal int         `json:"applyTotal"`
}

type Board struct {
	Apply     interface{} `json:"apply"`
	Publish   interface{} `json:"publish"`
	ApplyRank interface{} `json:"applyRank"`
	PubRank   interface{} `json:"pubRank"`
}

type deptRank struct {
	Rank  []*m.Rank `json:"rank"`
	Total int       `json:"total"`
}

type Number struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type Ratio struct {
	Dept  string  `json:"dept"`
	Ratio float64 `json:"ratio"`
	Count int     `json:"count"`
}

type SystemMsg struct {
	Theme   string   `json:"theme"`
	SendTo  []string `json:"sendTo"`
	Content string   `json:"content"`
	MsgTime int64    `json:"msgTime"`
}
type UpdateMsg struct {
	Content   string   `json:"content"`
	SendTo    []string `json:"sendTo"`
	ServiceId int      `json:"serviceId"`
	MsgTime   string   `json:"msgTime"`
}
type ServiceMsg struct {
	Service   string `json:"service"`   // 服务名
	Status    int    `json:"status"`    // 状态
	Content   string `json:"content"`   //详细信息
	StateTime string `json:"stateTime"` // 成功时间，驳回时间
	Data      string `json:"data"`      // 驳回原因
}

type UpdateInfo struct {
	Id        int    `json:"id"`
	Content   string `json:"content"`
	MsgTime   string `json:"msgTime"`
	Status    int    `json:"status"`
	ServiceId int    `json:"serviceId"`
	Classify  string `json:"classify"`
}

type SysInfo struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
	MsgTime string `json:"msgTime"`
	Status  int    `json:"status"`
}

type msgCenter struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
	MsgTime string `json:"msgTime"`
	Status  int    `json:"status"`
}

type publish struct {
	Mail        string `json:"邮箱号"`
	JobNum      string `json:"工号"`
	Name        string `json:"姓名"`
	Dept        string `json:"部门"`
	ServiceName string `json:"服务名称"`
	DeptLeader  string `json:"科室领导"`
	DeptHeads   string `json:"部门领导"`
	ServiceType string `json:"服务类型"`
	Link        string `json:"服务链接"`
}

type using struct {
	UseName   string `json:"申请人姓名"`
	UseMail   string `json:"申请人邮箱号"`
	UseDept   string `json:"申请人部门"`
	Developer string `json:"开发者姓名"`
	DevMail   string `json:"开发者邮箱号"`
	DevDept   string `json:"开发者部门"`
	DevLeader string `json:"开发者部门领导"`
	Service   string `json:"服务名称"`
	SrvType   string `json:"服务类型"`
	Project   string `json:"申请项目名称"`
	Desc      string `json:"申请服务用途说明"`
}

var (
	msgState = 2
	topic    = "cloud.micro.topic.notice"
	//PubSerial string
	//UseSerial string
)

/*
	A-首页
*/
func (g *Gin) Home(c *gin.Context) {
	var classify []string
	var req h.HomeReq
	var res home
	var display, token string

	err := c.ShouldBind(&req)
	if err != nil {
		log.Error("获取请求体失败: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    err,
			"message": "请求体解析失败!",
		})
		return
	}

	// 解析 token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		log.Info("token is null！")
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			//log.Info("返回值: ", v)
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期 或 token值不正确",
				})
				return
			}
			if k == "data" {
				display = v.(map[string]interface{})["displayName"].(string)
				//token=v.(map[string]interface{})["token"].(string)
				//mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
				token = v.(map[string]interface{})["token"].(string)
			}
		}
	}
	log.Info("单点登录的Token值: ", token)

	db := conf.DB.Raw("select name from service_classify "+
		"where `show`=? order by sort_by desc", 1)
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取分类失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取分类失败!",
		})
		return
	}

	for rows.Next() {
		var name string
		rows.Scan(&name)
		classify = append(classify, name)
	}
	res.Classify = classify

	// 获取广告
	var banner []h.AD
	var str string
	db = conf.DB.Raw("select content from ad where position='banner'")
	db.Row().Scan(&str)

	json.Unmarshal([]byte(str), &banner)
	res.Banner = banner

	var newService []interface{}
	db = conf.DB.Raw("select content from ad where position!='banner'")
	rows, err = db.Rows()
	if err != nil {
		log.Error("获取新服务失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取新服务失败!",
		})
		return
	}

	for rows.Next() {
		var s interface{}
		var cont string
		rows.Scan(&cont)

		err := json.Unmarshal([]byte(cont), &s)
		if err != nil {
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "获取新服务失败!",
			})
			return
		}

		newService = append(newService, s)
	}

	res.NewService = newService

	// 查询排名前N的服务
	var ranks []h.Rank
	db = conf.DB.Raw("select * from (" +
		"select id,service_name,department from service_app " +
		"where status in (6,5)) as service " +
		"left join (" +
		"select service_name," +
		"sum(page_view+apply_number*20+collect_number*5) c " +
		"from service_count " +
		"group by service_name) as count " +
		"on service.service_name=count.service_name " +
		"order by count.c desc limit 5")
	rows, err = db.Rows()
	if err != nil {
		log.Error("获取排名前5的服务数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取排名前5的服务数据失败!",
		})
		return
	}

	for rows.Next() {
		var rank h.Rank
		var str sql.NullString
		rows.Scan(&rank.Id, &rank.ServiceName, &rank.Dept, &str, &rank.Count)

		ranks = append(ranks, rank)
	}

	res.Rank = ranks

	// 按照排序规则查询服务信息
	var col, order, sqlStr string
	var service h.ServiceList
	var sLists = []*h.Service{}
	if req.Sort.Column == "applyNumber" {
		col = "count.apply_number"
	} else if req.Sort.Column == "createTime" {
		col = "service.create_at"
	} else {
		col = "count.s"
	}

	if req.Sort.Sort == 0 {
		order = "desc"
	} else {
		order = "asc"
	}

	db = conf.DB.Raw("select count(*) from service_app where status in (5,6)")
	db.Row().Scan(&service.Total)

	sqlStr = fmt.Sprintf("select * from "+
		"(select id,create_at,service_name,brief,deliver,tag,update_at "+
		"from service_app where status in (5,6)) as service "+
		"left join "+
		"(select service_name,"+
		"sum(page_view+apply_number*20+collect_number*5) s ,"+
		"apply_number from service_count "+
		"group by service_name) as count "+
		"on service.service_name=count.service_name "+
		"order by %s limit %d,%d",
		col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
	db = conf.DB.Raw(sqlStr)
	rows, err = db.Rows()
	if err != nil {
		log.Error("获取服务信息失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取服务信息失败!",
		})
		return
	}

	for rows.Next() {
		var s h.Service
		var tag, deliver string
		var ct, ut time.Time
		var str sql.NullString
		var count, apply sql.NullInt64
		rows.Scan(&s.Id, &ct, &s.ServiceName, &s.Brief,
			&deliver, &tag, &ut, &str, &count, &apply)

		db := conf.DB.Raw("select apply_status,collection_status "+
			"from service_use where service_id=? and applicant=?",
			s.Id, display)
		db.Row().Scan(&s.ApplyStatus, &s.CollectionStatus)

		tags := strings.Split(tag, ",")
		for _, v := range tags {
			s.Tag = append(s.Tag, v)
		}
		ss := strings.Split(deliver, ",")
		if len(ss) == 1 {
			s.Deliver = ss[0]
		} else {
			s.Deliver = ss[0] + "等"
		}

		s.UpdateTime = ut.Format("2006-01-02 15:04:05")

		sLists = append(sLists, &s)
	}
	rows.Close()

	service.List = sLists
	res.Service = service
	res.User = display

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"data":    res,
		"message": "获取数据成功!",
	})

	// 开协程插叙表单签审结果及原因

}

/*
	A1-服务发布申请
*/
func (g *Gin) ServiceApply(c *gin.Context) {
	var req a.ServiceApply
	var res a.PublishSuccess
	var dept, mail, display, jobNum, serial, token string
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("获取数据失败: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    err,
			"message": "解析数据失败!",
		})
		return
	}

	json.Unmarshal(body, &req)

	// 解析 token，获取 mail 和 department
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				display = v.(map[string]interface{})["displayName"].(string)
				mail = v.(map[string]interface{})["email"].(string)
				dept = v.(map[string]interface{})["department"].(string)
				jobNum = v.(map[string]interface{})["employeeId"].(string)
				token = v.(map[string]interface{})["token"].(string)
				//mobile = v.(map[string]interface{})["mobile"].(string)
			}
		}
	}

	// 调用签审平台
	pub := publish{}
	pub.Mail = mail
	pub.Name = display
	pub.JobNum = jobNum
	pub.Dept = dept
	pub.ServiceName = req.Name
	pub.ServiceType = req.ServiceType
	pub.Link = req.Url
	pub.DeptLeader = req.Leader
	pub.DeptHeads = req.Head

	pubData, _ := json.Marshal(pub)

	// `{"邮箱号":"180435","姓名":"test"}`
	reqBody := fmt.Sprintf(`{ "WFID":"%s","WFNodeOperationID":"%s","MainData":%s, "SubDatas":%s}`,
		strconv.Itoa(conf.Sc.PubWFID), strconv.Itoa(conf.Sc.PubId),
		string(pubData), "{}")

	request, err := http.NewRequest("POST", conf.Sc.Submit,
		strings.NewReader(reqBody))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("请求失败!")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "发起请求失败!",
		})
		return
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("appID", conf.Sc.Appid)
	request.Header.Add("appKey", conf.Sc.Appkey)
	request.Header.Add("token", token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("服务器错误!")
		c.JSON(200, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "服务器错误!",
		})
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取响应失败!")

		c.JSON(200, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "获取数据失败!",
		})
		return
	}
	sign := make(map[string]interface{})
	json.Unmarshal(data, &sign)
	log.Info("123456: ", sign)

	for k, v := range sign {
		if k == "Result" {
			serial = v.(string)
		}
	}
	log.Info("流水号: ", serial)

	// 将数据写入数据库
	var deliver, tag string
	var apply a.Apply

	file, _ := json.Marshal(req.File)
	image, _ := json.Marshal(req.Image)
	video, _ := json.Marshal(req.Video)

	for _, v := range req.Deliver {
		deliver += v + ","
	}
	deliver = strings.TrimRight(deliver, ",")
	for _, v := range req.Tag {
		tag += v + ","
	}
	tag = strings.TrimRight(tag, ",")

	apply.Name = req.Name
	apply.Serial = serial
	apply.ServiceType = req.ServiceType
	apply.Deliver = req.Deliver
	apply.Classify = req.Classify
	apply.Status = 3
	apply.ApplyTime = time.Now().Format("2006-01-02 15:04:05")

	info, _ := json.Marshal(apply)
	//display = display + "(" + mobile + ")"

	db := conf.DB.Exec("insert into service_app set create_at=?,update_at=?,"+
		"user=?,mail=?,department=?,dept_leader=?,service_name=?,service_type=?,"+
		"classify=?,deliver=?,tag=?,brief=?,detail=?,instruction=?,demo_url=?,"+
		"document=?,image=?,video=?,serial_number=?,status=?,apply_info=?",
		time.Now(), time.Now(), display, mail, dept, req.Head,
		req.Name, req.ServiceType, req.Classify, deliver, tag,
		req.Brief, req.Detail, req.Intro, req.Url, string(file),
		string(image), string(video),
		serial, 3, string(info))
	err = db.Error
	if err != nil {
		log.Error("新增数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "申请提交失败!",
		})
		return
	}
	db = conf.DB.Raw("select id from service_app "+
		"where serial_number=?", serial)
	db.Row().Scan(&res.Id)

	res.Serial = apply.Serial
	res.Name = req.Name
	res.Deliver = req.Deliver
	res.Mail = mail
	res.Charge = display
	res.Dept = dept

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    res,
		"message": "提交申请成功!",
	})
}

/*
	A2-服务列表
*/
func (g *Gin) ServiceListHeader(c *gin.Context) {
	var h header
	h.Dept = department
	h.Classify = class

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    h,
		"message": "success!",
	})
}
func (g *Gin) ServiceList(c *gin.Context) {
	var req h.ListReq
	var res h.ServiceList
	var display string
	var str1, str2 string
	err := c.ShouldBind(&req)
	if err != nil {
		log.Error("获取请求失败: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    err,
			"message": "解析数据失败",
		})
		return
	}

	// 解析 token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		log.Info("Token is null!")
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				display = v.(map[string]interface{})["displayName"].(string)
				//mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}

	var col, order string
	var lists = []*h.Service{}
	if req.Sort.Column == "applyNumber" {
		col = "count.apply_number"
	} else if req.Sort.Column == "createTime" {
		col = "service.create_at"
	} else {
		col = "count.s"
	}

	if req.Sort.Sort == 0 {
		order = "desc"
	} else {
		order = "asc"
	}

	str1 = fmt.Sprintf("select count(*) from service_app "+
		"where service_name like '%s' "+
		"and classify like '%s' "+
		"and department like '%s' and deliver like '%s' "+
		"and service_type like '%s'", "%"+req.ServiceName+"%",
		"%"+req.Classify+"%", "%"+req.Dept+"%",
		"%"+req.Deliver+"%", "%"+req.ServiceType+"%")
	str2 = fmt.Sprintf("select * from "+
		"(select id,create_at,service_name,brief,deliver,tag,update_at "+
		"from service_app where service_name like '%s' "+
		"and classify like '%s' "+
		"and department like '%s' and deliver like '%s' "+
		"and service_type like '%s' ) as service "+
		"left join "+
		"(select service_name,"+
		"sum(page_view+apply_number*20+collect_number*5) s ,"+
		"apply_number from service_count "+
		"group by service_name) as count "+
		"on service.service_name=count.service_name "+
		"order by %s limit %d,%d", "%"+req.ServiceName+"%",
		"%"+req.Classify+"%", "%"+req.Dept+"%",
		"%"+req.Deliver+"%", "%"+req.ServiceType+"%",
		col+" "+order, (req.Skip-1)*req.Limit, req.Limit)

	db := conf.DB.Raw(str1)
	db.Row().Scan(&res.Total)
	log.Info("总数: ", res.Total)

	db = conf.DB.Raw(str2)
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "解析数据失败",
		})
		return
	}

	for rows.Next() {
		var list h.Service
		var tag, deliver string
		var ct, ut time.Time
		var str sql.NullString
		var count, apply sql.NullInt64

		rows.Scan(&list.Id, &ct, &list.ServiceName, &list.Brief,
			&deliver, &tag, &ut, &str, &count, &apply)

		db := conf.DB.Raw("select apply_status,collection_status "+
			"from service_use where service_id=? and applicant=?",
			list.Id, display)
		db.Row().Scan(&list.ApplyStatus, &list.CollectionStatus)

		tags := strings.Split(tag, ",")
		for _, v := range tags {
			list.Tag = append(list.Tag, v)
		}
		s := strings.Split(deliver, ",")
		if len(s) == 1 {
			list.Deliver = s[0]
		} else {
			list.Deliver = s[0] + "等"
		}

		list.UpdateTime = ut.Format("2006-01-02 15:04:05")

		lists = append(lists, &list)
	}
	rows.Close()

	res.List = lists

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    res,
		"message": "获取数据成功!",
	})
}

/*
	B-服务详情
	1.只包括服务详情和留言
	2.包括详情，使用，留言
	3.未发布的详情页
	4.查看留言列表
	5.留言
	6.答复留言
*/
// 查看详情和使用说明  GET
func (g *Gin) DetailAndIntro(c *gin.Context) {
	var srvDetail = d.Service{}
	var deliver, tag, display string
	var tm time.Time

	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		log.Info("token is null！")
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			//log.Info("返回值: ", v)
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期 或 token值不正确",
				})
				return
			}
			if k == "data" {
				display = v.(map[string]interface{})["displayName"].(string)
				//token=v.(map[string]interface{})["token"].(string)
				//mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
				//token=v.(map[string]interface{})["token"].(string)
			}
		}
	}

	id := c.Query("id")

	srvDetail.Id, _ = strconv.ParseInt(id, 10, 64)

	db := conf.DB.Raw("select status,apply_status,collection_status "+
		"from service_use where service_id=? and applicant=?",
		id, display)
	db.Row().Scan(&srvDetail.UseApplyStatus,
		&srvDetail.ApplyStatus,
		&srvDetail.CollectionStatus)

	if srvDetail.ApplyStatus == 0 || srvDetail.UseApplyStatus != 1 {
		//未申请使用，只能查看服务的详情
		//var detail sql.NullString
		db = conf.DB.Raw("select service_name,brief,service_type,"+
			"deliver,classify,"+
			"department,user,update_at,tag,status,detail from service_app "+
			"where id=?", id)
		err := db.Row().Scan(&srvDetail.Name, &srvDetail.Brief, &srvDetail.Type,
			&deliver, &srvDetail.Classify, &srvDetail.Dept, &srvDetail.Charge,
			&tm, &tag, &srvDetail.Status, &srvDetail.Detail)
		if err != nil {
			log.Error("获取服务详细介绍失败: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "没有数据!",
			})
			return
		}

		srvDetail.UpdateTime = tm.Format("2006-01-02 15:04:05")

		str := strings.Split(deliver, ",")
		for _, v := range str {
			srvDetail.Deliver = append(srvDetail.Deliver, v)
		}

		tags := strings.Split(tag, ",")
		for _, v := range tags {
			srvDetail.Tag = append(srvDetail.Tag, v)
		}

		db = conf.DB.Raw("select sum(page_view+apply_number*20+collect_number*5) "+
			"from service_count where service_name=?", srvDetail.Name)
		db.Row().Scan(&srvDetail.Count)
	} else {
		//已申请，可以看到详情和使用说明
		var deliver, tag, file, image, video string
		var tm time.Time
		db := conf.DB.Raw("select service_name,brief,service_type,"+
			"deliver,classify,"+
			"department,user,update_at,tag,status,detail,instruction,"+
			"demo_url,document,image,video from service_app "+
			"where id=?", id)
		err := db.Row().Scan(&srvDetail.Name, &srvDetail.Brief, &srvDetail.Type,
			&deliver, &srvDetail.Classify, &srvDetail.Dept, &srvDetail.Charge,
			&tm, &tag, &srvDetail.Status, &srvDetail.Detail, &srvDetail.Intro,
			&srvDetail.Url, &file, &image, &video)
		if err != nil {
			log.Error("获取服务使用说明失败: ", err)
			return
		}
		srvDetail.UpdateTime = tm.Format("2006-01-02 15:04:05")

		json.Unmarshal([]byte(file), &srvDetail.File)

		json.Unmarshal([]byte(image), &srvDetail.Image)

		json.Unmarshal([]byte(video), &srvDetail.Video)

		str := strings.Split(deliver, ",")
		for _, v := range str {
			srvDetail.Deliver = append(srvDetail.Deliver, v)
		}

		tags := strings.Split(tag, ",")
		for _, v := range tags {
			srvDetail.Tag = append(srvDetail.Tag, v)
		}

		db = conf.DB.Raw("select sum(page_view+apply_number*20+collect_number*5) "+
			"from service_count where service_name=?", srvDetail.Name)
		db.Row().Scan(&srvDetail.Count)
	}

	//计算页面点击率 page view
	var t time.Time
	var count int

	db = conf.DB.Raw("select create_at from service_count "+
		"where service_name=? order by create_at desc limit 1",
		srvDetail.Name)
	db.Row().Scan(&t)

	timeStr := time.Now().Format("2006-01-02")

	sum := time.Now().Sub(t)

	if (sum.Hours() / 24) > 1 {
		// 新的一天
		count = count + 1
		log.Info("new page view: ", count)
		db := conf.DB.Exec("insert into service_count set create_at=?,update_at=?,"+
			"service_name=?,charge=?,department=?,page_view=?",
			time.Now(), time.Now(), srvDetail.Name, srvDetail.Charge,
			srvDetail.Dept, count)
		err := db.Error
		if err != nil {
			log.Error("新增PV失败: ", err)
			return
		}
	} else {
		var id int
		db = conf.DB.Raw("select id,page_view from service_count "+
			"where service_name=? and create_at like ?",
			srvDetail.Name, "%"+timeStr+"%")
		db.Row().Scan(&id, &count)

		if id == 0 {
			count = count + 1
			log.Info("page view: ", count)
			db := conf.DB.Exec("insert into service_count set create_at=?,update_at=?,"+
				"service_name=?,charge=?,department=?,page_view=?",
				time.Now(), time.Now(), srvDetail.Name, srvDetail.Charge,
				srvDetail.Dept, count)
			err := db.Error
			if err != nil {
				log.Error("新增PV失败: ", err)
				return
			}
		} else {
			count = count + 1
			log.Info("page view: ", count)
			db = conf.DB.Exec("update service_count set update_at=?,page_view=? "+
				"where id=?", time.Now(), count, id)
			err := db.Error
			if err != nil {
				log.Error("更新PV失败: ", err)
				return
			}
		}
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    srvDetail,
		"message": "获取服务详情成功!",
	})
}

// 留言
func (g *Gin) LeaveMessage(c *gin.Context) {
	var data d.Comment
	var service, tel string
	err := c.ShouldBind(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "数据解析失败!",
		})
		return
	}

	// 解析 token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		log.Info("token is null！")
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				//display = v.(map[string]interface{})["displayName"].(string)
				//mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
				tel = v.(map[string]interface{})["mobile"].(string)
			}
		}
	}

	db := conf.DB.Raw("select service_name from service_app "+
		"where id=?", data.ServiceId)
	db.Row().Scan(&service)

	from := data.Name
	db = conf.DB.Exec("insert into service_bbs set create_at=?,update_at=?,"+
		"service_name=?,service_id=?,msg_from=?,tel=?,content=?,write_time=?",
		time.Now(), time.Now(), service, data.ServiceId, from, tel,
		data.Message, time.Now().Format("2006-01-02 15:04:05"))
	err = db.Error
	if err != nil {
		log.Error("插入留言失败: ", err)
		return
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    err,
		"message": "感谢您的留言！",
	})

}

// 获取服务留言信息
func (g *Gin) GetLeaveMessage(c *gin.Context) {
	var res d.CommRes
	var total int64
	var lists = []*d.Msg{}
	id := c.Query("id")
	skip, _ := strconv.Atoi(c.Query("skip"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	db := conf.DB.Raw("select count(*) from service_bbs "+
		"where service_id=? and `show`=0", id)
	db.Row().Scan(&total)

	db = conf.DB.Raw("select msg_from,content,write_time,"+
		"respondent,reply,reply_time from service_bbs "+
		"where service_id=? and `show`=0 "+
		"order by write_time desc limit ?,?",
		id, (skip-1)*limit, limit)
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取留言信息失败: ", err)
		return
	}

	for rows.Next() {
		var list d.Msg
		rows.Scan(&list.Commenter, &list.Content, &list.WriteTime,
			&list.Respondent, &list.Reply, &list.ReplyTime)

		lists = append(lists, &list)
	}
	rows.Close()

	data := d.Data{}
	data.Total = total
	data.List = lists

	res.Code = 200
	res.Data = &data
	res.Message = "查询留言成功!"

	c.JSON(200, res)
}

// 更改收藏状态
func (g *Gin) ChangeStatus(c *gin.Context) {
	var display, dept, mail, service, charge string

	id := c.Query("id")
	status, _ := strconv.Atoi(c.Query("status"))
	tk := c.Request.Header.Get("Authorization")

	mp, err := util.AnalysisToken(tk)
	if err != nil {
		log.Error("解析token出错: ", err)
		return
	}
	for k, v := range mp {
		if v == nil {
			log.Info(v)
			return
		}
		if k == "data" {
			display = v.(map[string]interface{})["displayName"].(string)
			mail = v.(map[string]interface{})["email"].(string)
			dept = v.(map[string]interface{})["department"].(string)
		}
	}
	// 查询这个用户有没有收藏这个服务
	db := conf.DB.Raw("select service_name,user from service_app "+
		"where id=?", id)
	db.Row().Scan(&service, &charge)

	var tmp int
	db = conf.DB.Raw("select id from service_use "+
		"where service_id=? and email=?", id, mail)
	db.Row().Scan(&tmp)
	fmt.Println("查询到的ID: ", tmp)
	if tmp == 0 {
		db := conf.DB.Exec("insert into service_use set create_at=?,update_at=?,"+
			"service_id=?,service_name=?,charge=?,applicant=?,"+
			"email=?,department=?,collection_status=?",
			time.Now(), time.Now(), id, service, charge,
			display, mail, dept, status)
		err := db.Error
		if err != nil {
			log.Info("新增数据失败: ", err)
			return
		}
		log.Info("新增收藏信息: ", status)
	} else {
		db := conf.DB.Exec("update service_use set update_at=?,"+
			"collection_status=? where id=?",
			time.Now(), status, tmp)
		err = db.Error
		if err != nil {
			log.Error("更新数据失败: ", err)
			return
		}
		log.Info("更新收藏信息: ", status)
	}

	// 更新收藏数
	if status == 1 {
		var count int
		var t time.Time
		db = conf.DB.Raw("select create_at from service_count "+
			"where service_name=? order by create_at desc limit 1", service)
		db.Row().Scan(&t)

		tm := time.Now().Format("2006-01-02")

		sub := time.Now().Sub(t)

		if (sub.Hours() / 24) > 1 {
			count = count + 1
			log.Info("new collect number: ", count)
			db := conf.DB.Exec("insert into service_count set create_at=?,update_at=?,"+
				"service_name=?,charge=?,department=?,collect_number=?",
				time.Now(), time.Now(), service, charge,
				dept, count)
			err := db.Error
			if err != nil {
				log.Error("新增收藏数量失败: ", err)
				return
			}
		} else {
			var id int
			db = conf.DB.Raw("select id,collect_number from service_count "+
				"where service_name=? and create_at kile ?",
				service, "%"+tm+"%")
			db.Row().Scan(&id, &count)

			if id == 0 {
				count = count + 1
				log.Info("collect number: ", count)
				db := conf.DB.Exec("insert into service_count set create_at=?,update_at=?,"+
					"service_name=?,charge=?,department=?,collect_number=?",
					time.Now(), time.Now(), service, charge,
					dept, count)
				err := db.Error
				if err != nil {
					log.Error("新增收藏数量失败: ", err)
					return
				}
			} else {
				count = count + 1
				log.Info("collect number: ", count)
				db = conf.DB.Exec("update service_count set update_at=?,collect_number=? "+
					"where id=?", time.Now(), count, id)
				err := db.Error
				if err != nil {
					log.Error("更新收藏数量失败: ", err)
					return
				}
			}
		}

	} else {
		var count int
		var t time.Time
		db = conf.DB.Raw("select create_at from service_count "+
			"where service_name=? order by create_at desc limit 1", service)
		db.Row().Scan(&t)

		tm := time.Now().Format("2006-01-02")

		sub := time.Now().Sub(t)

		if (sub.Hours() / 24) > 1 {
			count = count - 1
			log.Info("new collect number(-1): ", count)
			db := conf.DB.Exec("insert into service_count set create_at=?,update_at=?,"+
				"service_name=?,charge=?,department=?,collect_number=?",
				time.Now(), time.Now(), service, charge,
				dept, count)
			err := db.Error
			if err != nil {
				log.Error("新增收藏数量失败: ", err)
				return
			}
		} else {
			var id int
			db = conf.DB.Raw("select id,collect_number from service_count "+
				"where service_name=? and create_at like ?",
				service, "%"+tm+"%")
			db.Row().Scan(&id, &count)
			log.Info("收藏数量--ID: ", id)

			count = count - 1
			log.Info("collect number: ", count)
			db = conf.DB.Exec("update service_count set update_at=?,collect_number=? "+
				"where id=?", time.Now(), count, id)
			err := db.Error
			if err != nil {
				log.Error("更新收藏数量失败: ", err)
				return
			}
		}

	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    "",
		"message": "更新收藏状态成功!",
	})
}

/*
	C-服务申请
	1.提交申请表格
	2.对接签审平台，提交使用申请
	3.提交成功，等待签审，等待使用
*/
func (g *Gin) GetHeader(c *gin.Context) {
	var name, brief string
	id := c.Query("id")

	db := conf.DB.Raw("select service_name,brief from service_app "+
		"where id =?", id)
	db.Row().Scan(&name, &brief)

	c.JSON(200, gin.H{
		"code": http.StatusOK,
		"data": struct {
			Name  string `json:"name"`
			Brief string `json:"brief"`
		}{
			Name:  name,
			Brief: brief,
		},
		"message": "success",
	})
}

func (g *Gin) ServiceUseApply(c *gin.Context) {
	var useApply a.UseApply
	var display, mail, dept, serial, token string
	var apply a.ApplyDetail
	var success a.Success

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "数据解析失败!",
		})
		return
	}

	json.Unmarshal(body, &useApply)

	// 解析token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				display = v.(map[string]interface{})["displayName"].(string)
				mail = v.(map[string]interface{})["email"].(string)
				dept = v.(map[string]interface{})["department"].(string)
				token = v.(map[string]interface{})["token"].(string)
			}
		}
	}

	// 调用签审平台，提交申请表
	log.Info("token 值: ", token)

	// 获取服务的管理者，部门领导，邮箱号，姓名
	var charge, devMail, devDept, devLeader, service, srvType string

	db := conf.DB.Raw("select user,mail,department,"+
		"dept_leader,service_name,service_type from service_app where id=?",
		useApply.Id)
	db.Row().Scan(&charge, &devMail, &devDept, &devLeader, &service, &srvType)

	use := using{}
	use.UseName = display
	use.UseMail = mail
	use.UseDept = dept
	use.Developer = charge
	use.DevMail = devMail
	use.DevDept = devDept
	use.DevLeader = devLeader
	use.Service = service
	use.SrvType = srvType
	use.Project = useApply.Name
	use.Desc = useApply.Reason

	useBody, _ := json.Marshal(use)

	reqBody := fmt.Sprintf(`{ "WFID":"%s","WFNodeOperationID":"%s","MainData":%s, "SubDatas":%s}`,
		strconv.Itoa(conf.Sc.UseWFID), strconv.Itoa(conf.Sc.UseId),
		string(useBody), "{}")

	request, err := http.NewRequest("POST", conf.Sc.Submit,
		strings.NewReader(reqBody))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("请求失败!")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "发起请求失败!",
		})
		return
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("appID", conf.Sc.Appid)
	request.Header.Add("appKey", conf.Sc.Appkey)
	request.Header.Add("token", token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("服务器错误!")
		c.JSON(200, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "服务器错误!",
		})
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取响应失败!")

		c.JSON(200, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "获取数据失败!",
		})
		return
	}

	sign := make(map[string]interface{})
	json.Unmarshal(data, &sign)
	log.Info("123456: ", sign)

	for k, v := range sign {
		if k == "Result" {
			serial = v.(string)
		}
	}
	log.Info("流水号: ", serial)

	// 录入数据库
	var id int
	db = conf.DB.Raw("select id from service_use where "+
		"service_id=? and email=?", useApply.Id, mail)
	db.Row().Scan(&id)
	fmt.Println("ID: ", id)

	apply.Service = service
	apply.ServiceId = useApply.Id
	apply.Serial = serial
	apply.Project = useApply.Name
	apply.Reason = useApply.Reason
	apply.ApplyTime = time.Now().Format("2006-01-02 15:04:05")
	apply.Status = 3

	info, _ := json.Marshal(apply)

	if id == 0 {
		db := conf.DB.Exec("insert into service_use set create_at=?,update_at=?,"+
			"service_name=?,service_id=?,charge=?,project_name=?,"+
			"serial_number=?,applicant=?,"+
			"email=?,department=?,apply_time=?,"+
			"description=?,status=?,apply_status=?,apply_info=?",
			time.Now(), time.Now(),
			service, useApply.Id, charge, useApply.Name,
			serial, display, mail, dept,
			time.Now().Format("2006-01-02 15:04:05"),
			useApply.Reason, 3, 1, string(info))
		err := db.Error
		if err != nil {
			log.Error("数据插入失败: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "数据插入失败!",
			})
			return
		}
	} else {
		db := conf.DB.Exec("update service_use set update_at=?,"+
			"serial_number=?,project_name=?,apply_time=?,description=?,"+
			"status=?,apply_status=?,apply_info=? where id=?",
			time.Now(), serial, useApply.Name,
			time.Now().Format("2006-01-02 15:04:05"),
			useApply.Reason, 3, 1, string(info), id)
		err := db.Error
		if err != nil {
			log.Error("数据更新失败: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "数据更新失败!",
			})
			return
		}
		log.Info("更新成功: ", db.RowsAffected)
	}

	success.Serial = apply.Serial
	success.Name = useApply.Name
	success.Reason = useApply.Reason
	success.Mail = mail
	success.Applicant = display
	success.Dept = dept

	// 计算申请数量
	var t time.Time
	var count int

	db = conf.DB.Raw("select create_time from service_count "+
		"where service_name=? order by create_at desc limit 1",
		service)
	db.Row().Scan(&t)

	tm := time.Now().Format("2006-01-02")
	sum := time.Now().Sub(t)

	if (sum.Hours() / 24) > 1 {
		count = count + 1
		log.Info("new apply number: ", count)
		db := conf.DB.Exec("insert into service_count set create_at=?,update_at=?,"+
			"service_name=?,charge=?,department=?,apply_number=?",
			time.Now(), time.Now(),
			service, charge, dept, count)
		err := db.Error
		if err != nil {
			log.Error("新增申请数量失败: ", err)
			return
		}
	} else {
		var id int
		db = conf.DB.Raw("select id,apply_number from service_count "+
			"where service_name=? and create_at like ?",
			service, "%"+tm+"%")
		db.Row().Scan(&id, &count)
		log.Info("申请数量--ID: ", id)

		if id == 0 {
			count = count + 1
			log.Info("apply number: ", count)
			db := conf.DB.Exec("insert into service_count set create_at=?,update_at=?,"+
				"service_name=?,charge=?,department=?,apply_number=?",
				time.Now(), time.Now(),
				service, charge, dept, count)
			err := db.Error
			if err != nil {
				log.Error("新增申请数量失败: ", err)
				return
			}
		} else {
			count = count + 1
			log.Info("apply number: ", count)
			db = conf.DB.Exec("update service_count set update_at=?,apply_number=? "+
				"where id=?", time.Now(), count, id)
			err := db.Error
			if err != nil {
				log.Error("更新申请数量失败: ", err)
				return
			}
		}
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    success,
		"message": "提交申请成功!",
	})

}

/*
	D-个人中心
*/
func (g *Gin) GetMessage(c *gin.Context) {
	var mail string
	var upInfo []UpdateInfo
	var sysInfo []SysInfo
	var upTotal, sysTotal int

	// 解析token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				//display = v.(map[string]interface{})["displayName"].(string)
				mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}

	db := conf.DB.Raw("select count(*) from receive "+
		"where recipient=? and msg_type=? and status=1 "+
		"order by receive_time desc", mail, "update")
	db.Row().Scan(&upTotal)
	db = conf.DB.Raw("select id,content,receive_time,status,service_id from receive "+
		"where recipient=? and msg_type=? "+
		"order by receive_time desc limit 5", mail, "update")
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取消息失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取失败!",
		})
		return
	}

	for rows.Next() {
		var info UpdateInfo
		rows.Scan(&info.Id, &info.Content, &info.MsgTime, &info.Status, &info.ServiceId)
		db := conf.DB.Raw("select classify from service_app "+
			"where id=?", info.ServiceId)
		db.Row().Scan(&info.Classify)

		upInfo = append(upInfo, info)
	}

	db = conf.DB.Raw("select count(*) from receive "+
		"where recipient=? and msg_type=? and status=1 "+
		"order by receive_time desc ", mail, "system")
	db.Row().Scan(&sysTotal)
	db = conf.DB.Raw("select id,content,receive_time,status from receive "+
		"where recipient=? and msg_type=? "+
		"order by receive_time desc limit 5", mail, "system")
	rows, err = db.Rows()
	if err != nil {
		log.Error("获取消息失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取失败!",
		})
		return
	}

	for rows.Next() {
		var info SysInfo
		rows.Scan(&info.Id, &info.Content, &info.MsgTime, &info.Status)

		sysInfo = append(sysInfo, info)
	}
	rows.Close()

	mp := map[string]interface{}{
		"update":   upInfo,
		"upTotal":  upTotal,
		"system":   sysInfo,
		"sysTotal": sysTotal,
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    mp,
		"message": "获取数据成功!",
	})
}

func (g *Gin) DisplayMessage(c *gin.Context) {
	var content string
	id := c.Query("id")
	db := conf.DB.Raw("select content from receive where id=?", id)
	db.Row().Scan(&content)

	db = conf.DB.Exec("update receive set update_at=?,"+
		"status=? where id=?", time.Now(), 0, id)
	log.Info("更新成功: ", db.RowsAffected)

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    content,
		"message": "获取成功!",
	})
}

/*
	D1-服务使用申请情况
	1.查询服务使用申请信息(分页+条件)
	2.查看申请信息
	3.查看使用说明
	4.查看驳回原因
	5.取消申请
*/
func (g *Gin) GetUseApply(c *gin.Context) {
	var req a.UseApplication
	var applicant, mail, sql1, sql2 string
	var total int64
	var uses = []*a.UseInfo{}

	err := c.ShouldBind(&req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "数据解析失败!",
		})
		return
	}

	//解析token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				applicant = v.(map[string]interface{})["displayName"].(string)
				mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}

	if req.Sort.Sort == 0 {
		if req.Tag == 0 {

			sql1 = fmt.Sprintf("select count(*) from service_use "+
				"where applicant='%s'", applicant)

			sql2 = fmt.Sprintf("select id,serial_number,service_name,project_name,"+
				"charge,apply_time,status from service_use "+
				"where applicant='%s' and serial_number like '%s' "+
				"and service_name like '%s' and project_name like '%s' "+
				"and charge like '%s' "+
				"order by update_at desc limit %d,%d",
				applicant, "%"+req.Serial+"%", "%"+req.ServiceName+"%",
				"%"+req.ProjectName+"%", "%"+req.Charge+"%",
				(req.Skip-1)*req.Limit, req.Limit)

		} else {

			sql1 = fmt.Sprintf("select count(*) from service_use "+
				"where applicant='%s' and status=%d", applicant, req.Tag)
			sql2 = fmt.Sprintf("select id,serial_number,service_name,project_name,"+
				"charge,apply_time,status from service_use "+
				"where applicant='%s' and status=%d "+
				"and serial_number like '%s' "+
				"and service_name like '%s' and project_name like '%s' "+
				"and charge like '%s' "+
				"order by update_at desc limit %d,%d",
				applicant, req.Tag,
				"%"+req.Serial+"%", "%"+req.ServiceName+"%",
				"%"+req.ProjectName+"%", "%"+req.Charge+"%",
				(req.Skip-1)*req.Limit, req.Limit)

		}
	} else {
		if req.Tag == 0 {

			sql1 = fmt.Sprintf("select count(*) from service_use "+
				"where applicant='%s'", applicant)

			sql2 = fmt.Sprintf("select id,serial_number,service_name,project_name,"+
				"charge,apply_time,status from service_use "+
				"where applicant='%s' and serial_number like '%s' "+
				"and service_name like '%s' and project_name like '%s' "+
				"and charge like '%s' "+
				"order by update_at asc limit %d,%d",
				applicant, "%"+req.Serial+"%", "%"+req.ServiceName+"%",
				"%"+req.ProjectName+"%", "%"+req.Charge+"%",
				(req.Skip-1)*req.Limit, req.Limit)

		} else {

			sql1 = fmt.Sprintf("select count(*) from service_use "+
				"where applicant='%s' and status=%d", applicant, req.Tag)
			sql2 = fmt.Sprintf("select id,serial_number,service_name,project_name,"+
				"charge,apply_time,status from service_use "+
				"where applicant='%s' and status=%d "+
				"and serial_number like '%s' "+
				"and service_name like '%s' and project_name like '%s' "+
				"and charge like '%s' "+
				"order by update_at asc limit %d,%d",
				applicant, req.Tag,
				"%"+req.Serial+"%", "%"+req.ServiceName+"%",
				"%"+req.ProjectName+"%", "%"+req.Charge+"%",
				(req.Skip-1)*req.Limit, req.Limit)

		}
	}

	db := conf.DB.Raw(sql1)
	db.Row().Scan(&total)

	var pass, reject int64
	db = conf.DB.Raw("select count(*) from receive "+
		"where recipient=? and msg_type=? and status=? "+
		"and apply_status=?", mail, "use", 1, 1)
	db.Row().Scan(&pass)
	db = conf.DB.Raw("select count(*) from receive "+
		"where recipient=? and msg_type=? and status=? "+
		"and apply_status=?", mail, "use", 1, 2)
	db.Row().Scan(&reject)

	db = conf.DB.Raw(sql2)
	rows, err := db.Rows()
	if err != nil {
		log.Error("查询数据失败: ", err)
		return
	}

	for rows.Next() {
		var use a.UseInfo
		rows.Scan(&use.Id, &use.Serial, &use.ServiceName,
			&use.ProjectName, &use.Charge, &use.ApplyTime, &use.Status)

		db := conf.DB.Raw("select id from service_app "+
			"where service_name=? and user like ?",
			use.ServiceName, "%"+use.Charge+"%")
		db.Row().Scan(&use.ServiceId)

		var id, s int
		db = conf.DB.Raw("select id,status from receive "+
			"where recipient=? and service_id=? "+
			"and msg_type=?",
			mail, use.ServiceId, "update")
		db.Row().Scan(&id, &s)

		if id != 0 && s == 1 {
			use.Update = true
		} else {
			use.Update = false
		}

		db = conf.DB.Raw("select id,status from receive "+
			"where recipient=? and service_id=? and msg_type=? "+
			"order by receive_time desc limit 1",
			mail, use.ServiceId, "use")
		db.Row().Scan(&use.MsgId, &use.MsgStatus)

		uses = append(uses, &use)
	}
	rows.Close()

	data := a.UseData{}
	data.Total = total
	data.List = uses
	data.Pass = pass
	data.Reject = reject

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    data,
		"message": "获取数据成功!",
	})

}

func (g *Gin) GetApplyDetail(c *gin.Context) {
	var applyDetail a.ApplyDetail
	var info string
	var sId int

	id := c.Query("id")

	db := conf.DB.Raw("select service_id,apply_info from service_use "+
		"where id=?", id)
	db.Row().Scan(&sId, &info)
	db = conf.DB.Exec("update receive set update_at=?,"+
		"status=? where service_id=?", time.Now(), 0, sId)
	log.Info("更新数量: ", db.RowsAffected)

	json.Unmarshal([]byte(info), &applyDetail)

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    applyDetail,
		"message": "获取申请信息成功!",
	})
}

func (g *Gin) GetInstruction(c *gin.Context) {
	var srvDetail d.Service
	var deliver, tag, file, image, video string
	var tm time.Time
	id := c.Query("id")

	db := conf.DB.Raw("select service_name,brief,service_type,"+
		"deliver,classify,"+
		"department,user,update_at,tag,status,detail,instruction,"+
		"demo_url,document,image,video from service_app "+
		"where id=?", id)
	err := db.Row().Scan(&srvDetail.Name, &srvDetail.Brief, &srvDetail.Type,
		&deliver, &srvDetail.Classify, &srvDetail.Dept, &srvDetail.Charge,
		&tm, &tag, &srvDetail.Status, &srvDetail.Detail, &srvDetail.Intro,
		&srvDetail.Url, &file, &image, &video)
	if err != nil {
		log.Error("获取服务使用说明失败: ", err)
		return
	}
	srvDetail.UpdateTime = tm.Format("2001-01-02 15:04:05")

	json.Unmarshal([]byte(file), &srvDetail.File)

	json.Unmarshal([]byte(image), &srvDetail.Image)

	json.Unmarshal([]byte(video), &srvDetail.Video)

	str := strings.Split(deliver, ",")
	for _, v := range str {
		srvDetail.Deliver = append(srvDetail.Deliver, v)
	}

	tags := strings.Split(tag, ",")
	for _, v := range tags {
		srvDetail.Tag = append(srvDetail.Tag, v)
	}
	srvDetail.ApplyStatus = 1

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    srvDetail,
		"message": "获取使用说明成功!",
	})
}

//查看驳回原因
func (g *Gin) GetRejectionReason(c *gin.Context) {
	var applyDetail a.ApplyDetail
	var info string

	id := c.Query("id")

	db := conf.DB.Raw("select apply_info from service_use "+
		"where id=?", id)
	db.Row().Scan(&info)

	json.Unmarshal([]byte(info), &applyDetail)

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    applyDetail.Data,
		"message": "查看驳回原因成功!",
	})
}

//取消使用申请 调用签审平台
func (g *Gin) CancelApply(c *gin.Context) {
	var serial, token, mail, info string
	id := c.Query("id")

	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				//display = v.(map[string]interface{})["displayName"].(string)
				mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
				token = v.(map[string]interface{})["token"].(string)
			}
		}
	}

	// 调用签审平台，提交申请表
	// 使用
	db := conf.DB.Raw("select serial_number,apply_info from service_use where id=?", id)
	db.Row().Scan(&serial, &info)
	log.Info("使用申请流水号: ", serial)
	use := using{}
	useBody, _ := json.Marshal(use)

	reqBody := fmt.Sprintf(`{ "WFID":"%s","BusinessId":"%s","WFNodeOperationID":"%s","MainData":%s, "SubDatas":%s}`,
		strconv.Itoa(conf.Sc.UseWFID), serial, strconv.Itoa(conf.Sc.UseCancel),
		string(useBody), "{}")

	request, err := http.NewRequest("POST", conf.Sc.Submit,
		strings.NewReader(reqBody))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("请求失败!")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "发起请求失败!",
		})
		return
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("appID", conf.Sc.Appid)
	request.Header.Add("appKey", conf.Sc.Appkey)
	request.Header.Add("token", token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("服务器错误!")
		c.JSON(200, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "服务器错误!",
		})
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取响应失败!")

		c.JSON(200, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "获取数据失败!",
		})
		return
	}

	sign := make(map[string]interface{})
	json.Unmarshal(data, &sign)
	log.Info("123456: ", sign)

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    sign,
		"message": "取消成功!",
	})

	// 更新状态为取消状态
	var appInfo a.ApplyDetail
	json.Unmarshal([]byte(info), appInfo)
	appInfo.Status = 4
	appInfo.StateTime = time.Now().Format("2006-01-02 15:04:05")
	result, _ := json.Marshal(appInfo)

	db = conf.DB.Exec("update service_use set update_at=?,status=?,apply_info=? "+
		"where id=?", time.Now(), 4, string(result), id)
	err = db.Error
	if err != nil {
		log.Error("更新数据失败: ", err)
		return
	}

	// 将申请数减少
	var service string
	var count int
	db = conf.DB.Raw("select service_name from service_use where id =?", id)
	db.Row().Scan(&service)

	db = conf.DB.Raw("select apply_number from service_count where "+
		"service_name=? ", service)
	db.Row().Scan(&count)

	count = count - 1
	db = conf.DB.Exec("update service_count set update_at=?,apply_number=? "+
		"where service_name=?", time.Now(), count, service)
	err = db.Error
	if err != nil {
		log.Error("更改使用申请数量失败: ", err)
		return
	}

	// 发送消息
	var sMsg ServiceMsg
	msgType := "use"

	sMsg.StateTime = appInfo.StateTime
	sMsg.Service = service
	sMsg.Status = 4
	sMsg.Data = ""
	sMsg.Content = "已取消服务: " + service + "的使用申请"
	pub := map[string]interface{}{
		"level":  0,
		"type":   msgType,
		"uid":    mail,
		"status": 1,
		"data":   sMsg,
	}
	msg := &broker.Message{}
	msg.Body, err = json.Marshal(pub)
	if err != nil {
		log.Error("数据转换失败: ", err)
		return
	}
	if err := broker.Publish(topic, msg); err != nil {
		log.Error("发布信息失败: ", err)
		return
	}

	log.Info("发送信息成功: ", string(msg.Body))
}

/*
	D2-服务发布申请情况:
	1.查询服务发布申请信息
	2.获取申请信息
	3.修改发布申请的信息(这个涉及到k2-4-服务修改)
	4.上下架操作
	5.取消申请操作
*/
// 分页查询信息 POST
func (g *Gin) GetRelease(c *gin.Context) {
	var deliver, sql1, sql2 string
	var user string
	var total int64

	var lists = []*a.Details{}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "数据解析失败!",
		})
		return
	}
	// 解析token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				//display = v.(map[string]interface{})["displayName"].(string)
				user = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}

	req := a.Request{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Error("11111: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "数据解析错误!",
		})
		return
	}

	if len(req.Deliver) != 0 {
		for _, v := range req.Deliver {
			deliver += v + ","
		}
		deliver = strings.TrimRight(deliver, ",")
	}

	if req.Sort.Sort == 0 {
		// 降序 desc
		if req.Status == 0 {
			if req.Tag == 0 {
				// 三个参数都为空
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s'"+
					"and service_type like '%s'",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' "+
					"order by update_at DESC limit %d,%d",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%",
					(req.Skip-1)*req.Limit, req.Limit)
			} else if req.Tag == 1 {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and status in (1,5,6) and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' ",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and status in (1,5,6) and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' "+
					"order by update_at DESC limit %d,%d",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%",
					(req.Skip-1)*req.Limit, req.Limit)
			} else {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and status=%d and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' ",
					user, req.Tag, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and status=%d and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' "+
					"order by update_at DESC limit %d,%d",
					user, req.Tag, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%",
					(req.Skip-1)*req.Limit, req.Limit)
			}
		} else {
			// status 不为 0
			if req.Tag == 0 {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d and service_type like '%s' ",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					req.Status, "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d "+
					"and service_type like '%s' "+
					"order by update_at DESC limit %d,%d",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", req.Status,
					"%"+req.ServiceType+"%", (req.Skip-1)*req.Limit, req.Limit)
			} else if req.Tag == 1 {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and status in (1,5,6) and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d and service_type like '%s' ",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					req.Status, "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and status in (1,5,6) and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d "+
					"and service_type like '%s' "+
					"order by update_at DESC limit %d,%d",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", req.Status,
					"%"+req.ServiceType+"%", (req.Skip-1)*req.Limit, req.Limit)
			} else {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and status=%d and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d and service_type like '%s' ",
					user, req.Tag, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					req.Status, "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and status=%d and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d "+
					"and service_type like '%s' "+
					"order by update_at DESC limit %d,%d",
					user, req.Tag, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", req.Status,
					"%"+req.ServiceType+"%", (req.Skip-1)*req.Limit, req.Limit)
			}
		}

	} else {
		// 升序 ASC
		if req.Status == 0 {
			if req.Tag == 0 {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s'"+
					"and service_type like '%s' ",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' "+
					"order by update_at ASC limit %d,%d",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%",
					(req.Skip-1)*req.Limit, req.Limit)
			} else if req.Tag == 1 {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and status in (1,5,6) and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' ",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and status in (1,5,6) and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' "+
					"order by update_at ASC limit %d,%d",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%",
					(req.Skip-1)*req.Limit, req.Limit)
			} else {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and status=%d and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' ",
					user, req.Tag, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and status=%d and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and service_type like '%s' "+
					"order by update_at ASC limit %d,%d",
					user, req.Tag, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", "%"+req.ServiceType+"%",
					(req.Skip-1)*req.Limit, req.Limit)
			}
		} else {
			// status 不为 0
			if req.Tag == 0 {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d and service_type like '%s' ",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					req.Status, "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d "+
					"and service_type like '%s' "+
					"order by update_at ASC limit %d,%d",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					req.Status, "%"+req.ServiceType+"%",
					(req.Skip-1)*req.Limit, req.Limit)
			} else if req.Tag == 1 {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and status in (1,5,6) and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d and service_type like '%s' ",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%", req.Status, "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and status in (1,5,6) and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d and service_type like '%s' "+
					"order by update_at ASC limit %d,%d",
					user, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					req.Status, "%"+req.ServiceType+"%",
					(req.Skip-1)*req.Limit, req.Limit)
			} else {
				sql1 = fmt.Sprintf("select count(*) from service_app "+
					"where mail='%s' and status=%d and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d and service_type like '%s'",
					user, req.Tag, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					req.Status, "%"+req.ServiceType+"%")
				sql2 = fmt.Sprintf("select id,serial_number,classify,service_name,"+
					"deliver,service_type,status,update_at from service_app "+
					"where mail='%s' and status=%d and serial_number like '%s' and classify like '%s' "+
					"and service_name like '%s' and deliver like '%s' "+
					"and status like %d and service_type like '%s' "+
					"order by update_at ASC limit %d,%d",
					user, req.Tag, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					req.Status, "%"+req.ServiceType+"%",
					(req.Skip-1)*req.Limit, req.Limit)
			}
		}

	}

	db := conf.DB.Raw(sql1)
	db.Row().Scan(&total)

	var pass, reject int64
	db = conf.DB.Raw("select count(*) from receive "+
		"where recipient=? and msg_type=? and status=? "+
		"and apply_status=?", user, "publish", 1, 1)
	db.Row().Scan(&pass)
	db = conf.DB.Raw("select count(*) from receive "+
		"where recipient=? and msg_type=? and status=? "+
		"and apply_status=?", user, "publish", 1, 2)
	db.Row().Scan(&reject)

	db = conf.DB.Raw(sql2)

	rows, err := db.Rows()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取用户失败!")

		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取服务器信息失败!",
		})

		return
	}

	for rows.Next() {
		var payment []string
		var tmp string
		var tm time.Time
		var list a.Details
		err := rows.Scan(&list.Id, &list.Serial, &list.Classify, &list.Name,
			&tmp, &list.ServiceType, &list.Status, &tm)

		list.UpdateTime = tm.Format("2006-01-02 15:04:15")

		db = conf.DB.Raw("select id,status from receive "+
			"where recipient=? and service_id=? and msg_type=? "+
			"order by receive_time desc limit 1",
			user, list.Id, "publish")
		db.Row().Scan(&list.MsgId, &list.MsgStatus)

		if err != nil {
			log.Error("22222: ", err)
			return
		}
		str := strings.Split(tmp, ",")
		for _, v := range str {

			payment = append(payment, v)
		}

		list.Deliver = payment

		lists = append(lists, &list)
	}
	rows.Close()

	data := a.Data{}
	data.Total = total
	data.List = lists
	data.Pass = pass
	data.Reject = reject

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    data,
		"message": "获取服务发布申请信息成功!",
	})
}

// 查看申请信息 GET
func (g *Gin) GetApplyInfo(c *gin.Context) {
	var info string
	id := c.Query("id")

	db := conf.DB.Raw("select apply_info from service_app "+
		"where id=?", id)
	db.Row().Scan(&info)
	log.Info("发布申请信息: ", info)

	data := a.Apply{}
	json.Unmarshal([]byte(info), &data)

	c.JSON(200, gin.H{
		"code":    200,
		"data":    data,
		"message": "获取申请详情成功!",
	})
}

/*
	3.修改发布申请的信息(这个涉及到k2-4-服务修改)
*/
// （1）获取某个信息 GET
func (g *Gin) GetOneInfo(c *gin.Context) {
	var info = m.Info{}
	var deliver, tag, file, image, video string
	id := c.Query("id")

	db := conf.DB.Raw("select id,service_name,service_type,classify,deliver,"+
		"tag,brief,status,detail,instruction,demo_url,"+
		"document,image,video from service_app "+
		"where id=?", id)
	err := db.Row().Scan(&info.Id, &info.Name, &info.ServiceType, &info.Classify,
		&deliver, &tag, &info.Brief, &info.Status, &info.Detail, &info.Intro,
		&info.Url, &file, &image, &video)
	if err != nil {
		log.Info("获取信息失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取信息失败!",
		})
		return
	}

	// 将字符串类型转换为结构体
	json.Unmarshal([]byte(file), &info.File)
	json.Unmarshal([]byte(image), &info.Image)
	json.Unmarshal([]byte(video), &info.Video)

	// 处理数组元素
	delivers := strings.Split(deliver, ",")
	for _, v := range delivers {
		info.Deliver = append(info.Deliver, v)
	}
	tags := strings.Split(tag, ",")
	for _, v := range tags {
		info.Tag = append(info.Tag, v)
	}

	info.ChooseClassify = class
	info.ChooseDeliver = de

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    info,
		"message": "success!",
	})
}

// （2）更新
func (g *Gin) ModifyInfo(c *gin.Context) {
	var info = m.Info{}
	var mail, dept string
	tk := c.Request.Header.Get("Authorization")
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "数据解析失败!",
		})
		return
	}
	json.Unmarshal(body, &info)
	log.Info("information: ", info)
	// 解析token
	myMap, _ := util.AnalysisToken(tk)
	for k, v := range myMap {
		if k == "data" {
			//display = v.(map[string]interface{})["displayName"].(string)
			mail = v.(map[string]interface{})["email"].(string)
			dept = v.(map[string]interface{})["department"].(string)
		}
	}

	var deliver, tag string
	for _, v := range info.Deliver {
		deliver += v + ","
	}
	deliver = strings.TrimRight(deliver, ",")

	for _, v := range info.Tag {
		tag += v + ","
	}
	tag = strings.TrimRight(tag, ",")

	file, _ := json.Marshal(info.File)
	image, _ := json.Marshal(info.Image)
	video, _ := json.Marshal(info.Video)

	db := conf.DB.Exec("update service_app set update_at=?,"+
		"department=?,classify=?,deliver=?,tag=?,brief=?,detail=?,"+
		"instruction=?,document=?,image=?,video=? where id=?",
		time.Now(), dept, info.Classify, deliver, tag, info.Brief,
		info.Detail, info.Intro, string(file),
		string(image), string(video), info.Id)
	err = db.Error
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("更新数据失败!")

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "更新数据失败!",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    "",
		"message": "更新数据成功!",
	})

	// 推送服务更新消息,发送到 ws 中
	var data UpdateMsg
	data.Content = info.Name + "更新"
	data.ServiceId = int(info.Id)
	data.MsgTime = time.Now().Format("2006-01-02 15:04:05")
	db = conf.DB.Raw("select applicant from service_use "+
		"where service_id=? and apply_status=1", info.Id)
	rows, err := db.Rows()
	if err != nil {
		log.Error("查询失败: ", err)
		return
	}
	for rows.Next() {
		var send string
		rows.Scan(&send)
		data.SendTo = append(data.SendTo, send)
	}
	pub := map[string]interface{}{
		"level":  0,
		"type":   "update",
		"uid":    mail,
		"dept":   "",
		"status": msgState,
		"data":   data,
	}
	msg := &broker.Message{}
	msg.Body, err = json.Marshal(pub)
	if err != nil {
		log.Error("数据转换失败: ", err)
		return
	}
	if err := broker.Publish(topic, msg); err != nil {
		log.Error("发布信息失败: ", err)
		return
	}

	log.Info("发送信息成功: ", string(msg.Body))

}

// 上架 下架 PUT
func (g *Gin) AlterStatus(c *gin.Context) {

	id := c.Query("id")
	status := c.Query("status")

	db := conf.DB.Exec("update service_app set status=?,update_at=? "+
		"where id=?", status, time.Now(), id)
	err := db.Error
	if err != nil {
		log.Error("更新状态失败: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "更新状态失败!",
		})
		return
	}

	log.WithFields(log.Fields{
		"rowsAffected": db.RowsAffected,
	}).Info("更新状态成功!")

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"error":   "",
		"message": "更新状态成功!",
	})

}

// 取消发布申请 调用签审
func (g *Gin) Unpublish(c *gin.Context) {
	var serial, service, token, mail, dept, info string
	id := c.Query("id")
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				//display = v.(map[string]interface{})["displayName"].(string)
				mail = v.(map[string]interface{})["email"].(string)
				dept = v.(map[string]interface{})["department"].(string)
				token = v.(map[string]interface{})["token"].(string)
			}
		}
	}

	// 调用签审平台，提交申请表
	log.Info("token 值: ", token)

	db := conf.DB.Raw("select service_name,serial_number,apply_info from service_app where id=?", id)
	db.Row().Scan(&service, &serial, &info)
	log.Info("发布申请流水号: ", serial)
	pub := publish{}
	pub.Mail = ""
	pub.Name = ""
	pub.JobNum = ""
	pub.Dept = ""
	pub.ServiceName = ""
	pub.ServiceType = ""
	pub.Link = ""
	pub.DeptLeader = ""
	pub.DeptHeads = ""
	pubData, _ := json.Marshal(pub)

	reqBody := fmt.Sprintf(`{ "WFID":"%s","BusinessId":"%s","WFNodeOperationID":"%s","MainData":%s, "SubDatas":%s}`,
		strconv.Itoa(conf.Sc.PubWFID), serial, strconv.Itoa(conf.Sc.PubCancel),
		string(pubData), "{}")

	request, err := http.NewRequest("POST", conf.Sc.Submit,
		strings.NewReader(reqBody))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("请求失败!")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "发起请求失败!",
		})
		return
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("appID", conf.Sc.Appid)
	request.Header.Add("appKey", conf.Sc.Appkey)
	request.Header.Add("token", token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("服务器错误!")
		c.JSON(200, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "服务器错误!",
		})
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取响应失败!")

		c.JSON(200, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "获取数据失败!",
		})
		return
	}

	sign := make(map[string]interface{})
	json.Unmarshal(data, &sign)
	log.Info("123456: ", sign)

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    sign,
		"message": "取消成功!",
	})


	// 更新
	var apply a.Apply
	json.Unmarshal([]byte(info), &apply)
	apply.Status = 4
	apply.StateTime = time.Now().Format("2006-01-02 15:04:05")
	result, _ := json.Marshal(apply)

	db = conf.DB.Exec("update service_app set update_at=?,status=?,apply_info=? "+
		"where id=?", time.Now(), 4, string(result), id)
	err = db.Error
	if err != nil {
		log.Error("跟新数据失败: ", err)
		return
	}

	// 发送消息
	var sMsg ServiceMsg
	msgType := "publish"

	sMsg.StateTime = apply.StateTime
	sMsg.Service = service
	sMsg.Status = int(apply.Status)
	sMsg.Data = ""
	sMsg.Content = "已取消服务: " + service + "的发布申请!"
	pubMsg := map[string]interface{}{
		"level":  0,
		"type":   msgType,
		"uid":    mail,
		"dept":   dept,
		"status": 1,
		"data":   sMsg,
	}
	msg := &broker.Message{}
	msg.Body, err = json.Marshal(pubMsg)
	if err != nil {
		log.Error("数据转换失败: ", err)
		return
	}
	if err := broker.Publish(topic, msg); err != nil {
		log.Error("发布信息失败: ", err)
		return
	}

	log.Info("发送信息成功: ", string(msg.Body))
}

/*
	D3-我的收藏
*/
func (g *Gin) MyCollection(c *gin.Context) {
	var res h.ServiceList
	var lists = []*h.Service{}
	var display string
	skip, _ := strconv.Atoi(c.Query("skip"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	// 解析 token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				display = v.(map[string]interface{})["displayName"].(string)
				//mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}
	db := conf.DB.Raw("select count(*) from service_use "+
		"where applicant=? and collection_status=?",
		display, 1)
	db.Row().Scan(&res.Total)

	db = conf.DB.Raw(fmt.Sprintf("select * from "+
		"(select update_at,service_id,apply_status,collection_status from service_use "+
		"where applicant='%s' and collection_status=%d) as collect "+
		"left join "+
		"(select id,service_name,brief,tag "+
		"from service_app) as service "+
		"on collect.service_id=service.id "+
		"order by collect.update_at desc limit %d,%d",
		display, 1, (skip-1)*limit, limit))
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取数据失败",
		})
		return
	}

	for rows.Next() {
		var id int
		var tag string
		var tm time.Time
		var list h.Service
		rows.Scan(&tm, &id, &list.ApplyStatus, &list.CollectionStatus,
			&list.Id, &list.ServiceName, &list.Brief, &tag)

		str := strings.Split(tag, ",")
		for _, v := range str {
			list.Tag = append(list.Tag, v)
		}

		lists = append(lists, &list)
	}
	rows.Close()

	res.List = lists
	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    res,
		"message": "获取我的收藏信息成功!",
	})
}

/*
	D4-消息中心
*/
func (g *Gin) MsgCenter(c *gin.Context) {
	var mail string
	var total, unread int
	var center []msgCenter
	skip, _ := strconv.Atoi(c.Query("skip"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	// 解析token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				//display = v.(map[string]interface{})["displayName"].(string)
				mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}

	db := conf.DB.Raw("select count(*) from receive "+
		"where recipient=? and msg_type=?", mail, "system")
	db.Row().Scan(&total)

	db = conf.DB.Raw("select count(*) from receive "+
		"where recipient=? and status=? and msg_type=?",
		mail, 1, "system")
	db.Row().Scan(&unread)

	db = conf.DB.Raw("select id,content,receive_time,status from receive "+
		"where recipient=? and msg_type=? "+
		"order by receive_time desc limit ?,?",
		mail, "system", (skip-1)*limit, limit)
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取消息失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取失败!",
		})
		return
	}

	for rows.Next() {
		var ct msgCenter
		rows.Scan(&ct.Id, &ct.Content, &ct.MsgTime, &ct.Status)

		center = append(center, ct)
	}
	rows.Close()

	mp := map[string]interface{}{
		"list":   center,
		"total":  total,
		"unread": unread,
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    mp,
		"message": "获取消息成功!",
	})
}

func (g *Gin) MsgCenterStatus(c *gin.Context) {
	id := c.Query("id")
	count := 0
	str := strings.Split(id, ",")
	for _, v := range str {
		tmp, _ := strconv.Atoi(v)
		if tmp == 0 {
			log.Info("ID为 0！！！！！")
			c.JSON(202, gin.H{
				"code":    http.StatusAccepted,
				"message": "ID为0",
			})
			return
		}
		db := conf.DB.Exec("update receive set update_at=?,"+
			"status=? where id=?", time.Now(), 0, tmp)
		err := db.Error
		if err != nil {
			log.Error("更新失败: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "更新失败!",
			})
			return
		}
		count++
	}

	log.Info("更新成功: ", count)

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    "",
		"message": "更新成功!",
	})
}

/*
	k2-服务发布管理   (发布者，发布者领导，公司领导，系统管理员)
	1.分页，条件查询
	2.申请信息
	3.服务详情
	4.服务统计
	5.服务修改(详见 ModifyInfo)
*/
func (g *Gin) GetServiceInfo(c *gin.Context) {
	// 没有进行角色分类的情况
	var req = m.Request{}
	var mail, dept, position, deliver, sql1, sql2, sql3, sql4, col, order string
	var total, lId, pass, reject int64
	var res m.Response
	var lists = []*m.Service{}

	err := c.ShouldBind(&req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "数据解析失败!",
		})
		return
	}

	// 解析 token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"data":    "",
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				//display = v.(map[string]interface{})["displayName"].(string)
				mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}

	//json.Unmarshal(body,&req)

	for _, v := range req.Deliver {
		deliver += v + ","
	}
	deliver = strings.TrimRight(deliver, ",")

	if req.Sort.Column == "applyNumber" {
		col = "count.s"
	} else {
		col = "app.update_at"
	}

	if req.Sort.Sort == 0 {
		// 降序
		order = "desc"
	} else {
		// 升序
		order = "asc"
	}

	// 服务发布者领导
	db := conf.DB.Raw("select id,department from middle_leader "+
		"where mail=?", mail)
	db.Row().Scan(&lId, &dept)

	// 发布者
	if lId == 0 {
		if req.Type == 0 {
			// 全部
			if req.Status == 0 {
				// 状态没有传值
				sql1 = fmt.Sprintf("select count(id) from service_app "+
					"where mail ='%s' and delete_at is null "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s'",
					mail, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					"%"+req.Dept+"%", "%"+req.ServiceType+"%", )
				sql2 = fmt.Sprintf("select * from ("+
					"select id,serial_number,classify,"+
					"service_name,deliver,department,user,"+
					"service_type,status,update_at "+
					"from service_app "+
					"where mail='%s' and delete_at is null "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s') as app "+
					"left join ("+
					"select service_name,sum(apply_number) s "+
					"from service_count group by service_name desc) "+
					"as count on app.service_name=count.service_name "+
					"order by %s limit %d,%d",
					mail, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					"%"+req.Dept+"%", "%"+req.ServiceType+"%",
					col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
			} else {
				sql1 = fmt.Sprintf("select count(id) from service_app "+
					"where mail ='%s' and delete_at is null "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s'"+
					"and status like %d",
					mail, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					"%"+req.Dept+"%", "%"+req.ServiceType+"%",
					req.Status, )
				sql2 = fmt.Sprintf("select * from ("+
					"select id,serial_number,classify,"+
					"service_name,deliver,department,user,"+
					"service_type,status,update_at "+
					"from service_app "+
					"where mail='%s' and delete_at is null "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s'"+
					"and status like %d) as app "+
					"left join ("+
					"select service_name,sum(apply_number) s "+
					"from service_count group by service_name desc) "+
					"as count on app.service_name=count.service_name "+
					"order by %s limit %d,%d",
					mail, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					"%"+req.Dept+"%", "%"+req.ServiceType+"%",
					req.Status, col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
			}
		} else if req.Type == 1 {
			if req.Status == 0 {
				sql1 = fmt.Sprintf("select count(id) from service_app "+
					"where mail ='%s' and delete_at is null "+
					"and status in (1,5,6) "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and department like '%s' "+
					"and deliver like '%s' "+
					"and service_type like '%s'",
					mail, "%"+req.Serial+"%",
					"%"+req.Classify+"%", "%"+req.Name+"%",
					"%"+deliver+"%", "%"+req.Dept+"%", "%"+req.ServiceType+"%", )
				sql2 = fmt.Sprintf("select * from ("+
					"select id,serial_number,classify,"+
					"service_name,deliver,department,user,"+
					"service_type,status,update_at "+
					"from service_app "+
					"where mail='%s' and delete_at is null "+
					"and status in (1,5,6) "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s') as app "+
					"left join ("+
					"select service_name,sum(apply_number) s "+
					"from service_count group by service_name desc) "+
					"as count on app.service_name=count.service_name "+
					"order by %s limit %d,%d",
					mail, "%"+req.Serial+"%",
					"%"+req.Classify+"%", "%"+req.Name+"%",
					"%"+deliver+"%", "%"+req.Dept+"%",
					"%"+req.ServiceType+"%", col+" "+order,
					(req.Skip-1)*req.Limit, req.Limit)
			} else {
				sql1 = fmt.Sprintf("select count(id) from service_app "+
					"where mail ='%s' and delete_at is null "+
					"and status in (1,5,6) "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s' "+
					"and status like %d",
					mail, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status, )
				sql2 = fmt.Sprintf("select * from ("+
					"select id,serial_number,classify,"+
					"service_name,deliver,department,user,"+
					"service_type,status,update_at "+
					"from service_app "+
					"where mail='%s' and delete_at is null "+
					"and status in (1,5,6) "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s' "+
					"and status like %d) as app "+
					"left join ("+
					"select service_name,sum(apply_number) s "+
					"from service_count group by service_name desc) "+
					"as count on app.service_name=count.service_name "+
					"order by %s limit %d,%d",
					mail, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status,
					col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
			}
		} else {
			if req.Status == 0 {
				sql1 = fmt.Sprintf("select count(id) from service_app "+
					"where mail ='%s' and delete_at is null "+
					"and status =%d "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s'",
					mail, req.Type, "%"+req.Serial+"%",
					"%"+req.Classify+"%", "%"+req.Name+"%",
					"%"+deliver+"%", "%"+req.Dept+"%", "%"+req.ServiceType+"%", )
				sql2 = fmt.Sprintf("select * from ("+
					"select id,serial_number,classify,"+
					"service_name,deliver,department,user,"+
					"service_type,status,update_at "+
					"from service_app "+
					"where mail='%s' and delete_at is null "+
					"and status = %d "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s') as app "+
					"left join ("+
					"select service_name,sum(apply_number) s "+
					"from service_count group by service_name desc) "+
					"as count on app.service_name=count.service_name "+
					"order by %s limit %d,%d",
					mail, req.Type, "%"+req.Serial+"%",
					"%"+req.Classify+"%", "%"+req.Name+"%",
					"%"+deliver+"%", "%"+req.Dept+"%", "%"+req.ServiceType+"%",
					col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
			} else {
				sql1 = fmt.Sprintf("select count(id) from service_app "+
					"where mail ='%s' and delete_at is null "+
					"and status =%d "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s' "+
					"and status like %d",
					mail, req.Type, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status, )
				sql2 = fmt.Sprintf("select * from ("+
					"select id,serial_number,classify,"+
					"service_name,deliver,department,user,"+
					"service_type,status,update_at "+
					"from service_app "+
					"where mail='%s' and delete_at is null "+
					"and status = %d "+
					"and serial_number like '%s' "+
					"and classify like '%s' "+
					"and service_name like '%s' "+
					"and deliver like '%s' "+
					"and department like '%s' "+
					"and service_type like '%s' "+
					"and status like %d) as app "+
					"left join ("+
					"select service_name,sum(apply_number) s "+
					"from service_count group by service_name desc) "+
					"as count on app.service_name=count.service_name "+
					"order by %s limit %d,%d",
					mail, req.Type, "%"+req.Serial+"%", "%"+req.Classify+"%",
					"%"+req.Name+"%", "%"+deliver+"%",
					"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status,
					col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
			}
		}
		sql3 = fmt.Sprintf("select count(*) from receive "+
			"where recipient='%s' and msg_type='%s' and status=%d "+
			"and apply_status=%d", mail, "publish", 1, 1)
		sql4 = fmt.Sprintf("select count(*) from receive "+
			"where recipient='%s' and msg_type='%s' and status=%d "+
			"and apply_status=%d", mail, "publish", 1, 2)
	} else {
		db = conf.DB.Raw("select user_type from service_admin "+
			"where mail=?", mail)
		db.Row().Scan(&position)

		if position == "公司领导" {
			if req.Type == 0 {
				// 全部
				if req.Status == 0 {
					// 状态没有传值
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where delete_at is null "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s'",
						"%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where delete_at is null "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s') as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						"%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%",
						col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				} else {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where delete_at is null "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s'"+
						"and status like %d",
						"%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%",
						req.Status, )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where delete_at is null "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s'"+
						"and status like %d) as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						"%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%",
						req.Status, col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				}
			} else if req.Type == 1 {
				if req.Status == 0 {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where delete_at is null "+
						"and status in (1,5,6) "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and department like '%s' "+
						"and deliver like '%s' "+
						"and service_type like '%s'",
						"%"+req.Serial+"%",
						"%"+req.Classify+"%", "%"+req.Name+"%",
						"%"+deliver+"%", "%"+req.Dept+"%", "%"+req.ServiceType+"%", )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where delete_at is null "+
						"and status in (1,5,6) "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s') as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limir %d,%d",
						"%"+req.Serial+"%",
						"%"+req.Classify+"%", "%"+req.Name+"%",
						"%"+deliver+"%", "%"+req.Dept+"%",
						"%"+req.ServiceType+"%", col+" "+order,
						(req.Skip-1)*req.Limit, req.Limit)
				} else {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where delete_at is null "+
						"and status in (1,5,6) "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s' "+
						"and status like %d",
						"%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status, )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where delete_at is null "+
						"and status in (1,5,6) "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s' "+
						"and status like %d) as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						"%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status,
						col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				}
			} else {
				if req.Status == 0 {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where and delete_at is null "+
						"and status =%d "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s'",
						req.Type, "%"+req.Serial+"%",
						"%"+req.Classify+"%", "%"+req.Name+"%",
						"%"+deliver+"%", "%"+req.Dept+"%", "%"+req.ServiceType+"%", )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where delete_at is null "+
						"and status = %d "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s') as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						req.Type, "%"+req.Serial+"%",
						"%"+req.Classify+"%", "%"+req.Name+"%",
						"%"+deliver+"%", "%"+req.Dept+"%", "%"+req.ServiceType+"%",
						col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				} else {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where delete_at is null "+
						"and status =%d "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s' "+
						"and status like %d",
						req.Type, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status, )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where delete_at is null "+
						"and status = %d "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s' "+
						"and status like %d) as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						req.Type, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status,
						col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				}
			}
			sql3 = fmt.Sprintf("select count(*) from receive "+
				"where msg_type='%s' and status=%d "+
				"and apply_status=%d", "publish", 1, 1)
			sql4 = fmt.Sprintf("select count(*) from receive "+
				"where msg_type='%s' and status=%d "+
				"and apply_status=%d", "publish", 1, 2)
		} else {
			if req.Type == 0 {
				// 全部
				if req.Status == 0 {
					// 状态没有传值
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where department ='%s' and delete_at is null "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s'",
						dept, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where department='%s' and delete_at is null "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s') as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						dept, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%",
						col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				} else {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where department ='%s' and delete_at is null "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s'"+
						"and status like %d",
						dept, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%",
						req.Status, )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where department='%s' and delete_at is null "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s'"+
						"and status like %d) as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						dept, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%",
						req.Status, col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				}
			} else if req.Type == 1 {
				if req.Status == 0 {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where department ='%s' and delete_at is null "+
						"and status in (1,5,6) "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and department like '%s' "+
						"and deliver like '%s' "+
						"and service_type like '%s'",
						dept, "%"+req.Serial+"%",
						"%"+req.Classify+"%", "%"+req.Name+"%",
						"%"+deliver+"%", "%"+req.Dept+"%", "%"+req.ServiceType+"%", )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where department='%s' and delete_at is null "+
						"and status in (1,5,6) "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s') as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limir %d,%d",
						dept, "%"+req.Serial+"%",
						"%"+req.Classify+"%", "%"+req.Name+"%",
						"%"+deliver+"%", "%"+req.Dept+"%",
						"%"+req.ServiceType+"%", col+" "+order,
						(req.Skip-1)*req.Limit, req.Limit)
				} else {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where department ='%s' and delete_at is null "+
						"and status in (1,5,6) "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s' "+
						"and status like %d",
						dept, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status, )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where department='%s' and delete_at is null "+
						"and status in (1,5,6) "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s' "+
						"and status like %d) as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						dept, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status,
						col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				}
			} else {
				if req.Status == 0 {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where department ='%s' and delete_at is null "+
						"and status =%d "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s'",
						dept, req.Type, "%"+req.Serial+"%",
						"%"+req.Classify+"%", "%"+req.Name+"%",
						"%"+deliver+"%", "%"+req.Dept+"%", "%"+req.ServiceType+"%", )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where department='%s' and delete_at is null "+
						"and status = %d "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s') as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						dept, req.Type, "%"+req.Serial+"%",
						"%"+req.Classify+"%", "%"+req.Name+"%",
						"%"+deliver+"%", "%"+req.Dept+"%", "%"+req.ServiceType+"%",
						col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				} else {
					sql1 = fmt.Sprintf("select count(id) from service_app "+
						"where department ='%s' and delete_at is null "+
						"and status =%d "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s' "+
						"and status like %d",
						dept, req.Type, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status, )
					sql2 = fmt.Sprintf("select * from ("+
						"select id,serial_number,classify,"+
						"service_name,deliver,department,user,"+
						"service_type,status,update_at "+
						"from service_app "+
						"where department='%s' and delete_at is null "+
						"and status = %d "+
						"and serial_number like '%s' "+
						"and classify like '%s' "+
						"and service_name like '%s' "+
						"and deliver like '%s' "+
						"and department like '%s' "+
						"and service_type like '%s' "+
						"and status like %d) as app "+
						"left join ("+
						"select service_name,sum(apply_number) s "+
						"from service_count group by service_name desc) "+
						"as count on app.service_name=count.service_name "+
						"order by %s limit %d,%d",
						dept, req.Type, "%"+req.Serial+"%", "%"+req.Classify+"%",
						"%"+req.Name+"%", "%"+deliver+"%",
						"%"+req.Dept+"%", "%"+req.ServiceType+"%", req.Status,
						col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
				}
			}
			sql3 = fmt.Sprintf("select count(*) from receive "+
				"where department='%s' and msg_type='%s' and status=%d "+
				"and apply_status=%d", dept, "publish", 1, 1)
			sql4 = fmt.Sprintf("select count(*) from receive "+
				"where department='%s' and msg_type='%s' and status=%d "+
				"and apply_status=%d", dept, "publish", 1, 2)
		}

	}

	db = conf.DB.Raw(sql1)
	err = db.Row().Scan(&total)
	if err != nil {
		log.Error("查询总数失败: ", err)
		return
	}

	db = conf.DB.Raw(sql3)
	db.Row().Scan(&pass)
	db = conf.DB.Raw(sql4)
	db.Row().Scan(&reject)

	db = conf.DB.Raw(sql2)
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取数据失败: ", err)
		return
	}

	for rows.Next() {
		var list m.Service
		var tm time.Time
		var deliver string
		var service sql.NullString
		var sum sql.NullInt64
		err := rows.Scan(&list.Id, &list.Serial, &list.Classify,
			&list.Name, &deliver, &list.Dept, &list.Charge,
			&list.ServiceType, &list.Status, &tm,
			&service, &sum)
		list.UpdateTime = tm.Format("2006-01-02 15:04:05")
		if err != nil {
			log.Error("获取数据错误", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "查询数据失败!",
			})
			return
		}

		db = conf.DB.Raw("select id,status from receive "+
			"where service_id=? and msg_type=? "+
			"order by receive_time desc limit 1",
			list.Id, "publish")
		db.Row().Scan(&list.MsgId, &list.MsgStatus)

		if sum.Valid {
			list.ApplyNumber = sum.Int64
		} else {
			list.ApplyNumber = 0
		}

		str := strings.Split(deliver, ",")
		for _, v := range str {
			list.Deliver = append(list.Deliver, v)
		}

		lists = append(lists, &list)
	}
	rows.Close()

	data := m.Data{}
	data.Total = total
	data.List = lists
	data.Classify = class
	data.Deliver = de
	data.Type = st
	data.Dept = department
	data.Pass = pass
	data.Reject = reject

	res.Code = 200
	res.Data = &data
	res.Message = "获取服务发布信息成功!"

	c.JSON(200, res)

}

// 获取服务发布申请信息同 GetApplyInfo

// 获取服务详细信息
func (g *Gin) GetServiceDetail(c *gin.Context) {
	var srvDetail m.ServiceDetail
	var deliver, tag, document, image, video string
	var tm time.Time
	id := c.Query("id")

	db := conf.DB.Raw("select id,service_name,deliver,classify,"+
		"service_type,department,user,update_at,tag,brief,"+
		"detail,instruction,demo_url,document,image,video "+
		"from service_app where id=?", id)
	err := db.Row().Scan(&srvDetail.Id, &srvDetail.Name, &deliver, &srvDetail.Classify,
		&srvDetail.ServiceType, &srvDetail.Dept, &srvDetail.Charge,
		&tm, &tag, &srvDetail.Brief,
		&srvDetail.Detail, &srvDetail.Intro, &srvDetail.Url,
		&document, &image, &video)
	if err != nil {
		log.Error("获取数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取服务详情失败!",
		})
		return
	}

	str := strings.Split(deliver, ",")
	for _, v := range str {
		srvDetail.Deliver = append(srvDetail.Deliver, v)
	}

	tags := strings.Split(tag, ",")
	for _, v := range tags {
		srvDetail.Tag = append(srvDetail.Tag, v)
	}
	srvDetail.UpdateTime = tm.Format("2006-01-02 15:04:05")

	json.Unmarshal([]byte(document), &srvDetail.File)
	json.Unmarshal([]byte(image), &srvDetail.Image)
	json.Unmarshal([]byte(video), &srvDetail.Video)

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    srvDetail,
		"message": "获取服务详情成功!",
	})
}

// 在线查看文档或视频
func (g *Gin) OnlinePreview(c *gin.Context) {
	var content string
	file := c.Query("file")

	// 读取文件
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error("读取文件失败: ", err)
		c.JSON(200, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "读取文件失败!",
		})
		return
	}

	switch {
	case strings.HasSuffix(file, ".pdf"):
		content = "application/pdf"
		break
	case strings.HasSuffix(file, ".mp4"):
		content = "video/mp4"
		break
	case strings.HasSuffix(file, ".avi"):
		content = "video/avi"
		break
	case strings.HasSuffix(file, ".jpeg") || strings.HasSuffix(file, ".jpg"):
		content = "image/jpeg"
		break
	case strings.HasSuffix(file, ".png"):
		content = "image/png"
		break
	case strings.HasSuffix(file, ".gif"):
		content = "image/gif"
		break
	default:
		content = "application/octet-stream"
	}
	c.Writer.Header().Add("Content-Type", content)
	c.Writer.Write(data)
}

// 服务统计
func (g *Gin) StatisticsHeader(c *gin.Context) {
	var service string
	var applyNumber, pv int
	mp := make(map[string]interface{})
	id := c.Query("id")

	db := conf.DB.Raw("select service_name from service_app "+
		"where id=?", id)
	db.Row().Scan(&service)

	db = conf.DB.Raw("select sum(apply_number*20+collect_number*5),"+
		"sum(page_view) from service_count "+
		"where service_name = ?", service)
	db.Row().Scan(&applyNumber, &pv)

	mp["applyNumber"] = applyNumber
	mp["pageView"] = pv

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    mp,
		"message": "获取数据成功!",
	})

}

func (g *Gin) Statistics(c *gin.Context) {
	var service, str1, str2, str3 string
	var ranks = []*m.Rank{}
	var month = []*m.ApplyStats{}
	var res m.Statistics
	id, _ := strconv.Atoi(c.Query("id"))
	skip, _ := strconv.Atoi(c.Query("skip"))
	order := c.Query("order")

	db := conf.DB.Raw("select service_name from service_app "+
		"where id=?", id)
	db.Row().Scan(&service)

	if order == "week" {
		str1 = fmt.Sprintf("select department,date_format(`create_at`, '%s'),"+
			"count(department) from service_use "+
			"where yearweek(date_format(`create_at`, '%s')) = yearweek(now()) "+
			"and apply_status=%d "+
			"group by week(`create_at`),department "+
			"order by count(department) desc,"+
			"week(`create_at`) desc limit %d,%d",
			"%Y-%m-%d", "%Y-%m-%d", 1, (skip-1)*10, 10)
		str2 = fmt.Sprintf("select sum(apply_number*20+5*collect_number),"+
			"date_format(`create_at`, '%s') "+
			"from service_count "+
			"where yearweek(date_format(`create_at`, '%s')) = yearweek(now()) "+
			"and service_name='%s' "+
			"group by date_format(`create_at`, '%s') "+
			"order by date_format(`create_at`, '%s') asc",
			"%Y-%m-%d", "%Y-%m-%d", service,
			"%Y-%m-%d", "%Y-%m-%d")
		str3 = fmt.Sprintf("select count(*) from "+
			"(select department,date_format(`create_at`, '%s') "+
			"from service_use "+
			"where yearweek(date_format(`create_at`, '%s')) = yearweek(now()) "+
			"and apply_status=%d "+
			"group by week(`create_at`),department) as count",
			"%Y-%m-%d", "%Y-%m-%d", 1)
	} else if order == "month" {
		str1 = fmt.Sprintf("select department,"+
			"date_format(`create_at`,'%s'),"+
			"count(department) from service_use "+
			"where apply_status=%d "+
			"group by date_format(`create_at`,'%s'),department "+
			"order by count(department) desc,"+
			"date_format(`create_at`,'%s') desc limit %d,%d",
			"%Y-%m", 1, "%Y-%m", "%Y-%m",
			(skip-1)*10, 10)
		str2 = fmt.Sprintf("select sum(apply_number*20+5*collect_number),"+
			"date_format(`create_at`, '%s') "+
			"from service_count "+
			"where date_format(`create_at`, '%s') = date_format(now(), '%s') "+
			"and service_name='%s' "+
			"group by date_format(`create_at`, '%s') "+
			"order by date_format(`create_at`, '%s') asc",
			"%Y-%m-%d", "%Y-%m", "%Y-%m", service,
			"%Y-%m-%d", "%Y-%m-%d")
		str3 = fmt.Sprintf("select count(*) from "+
			"(select department,"+
			"date_format(`create_at`,'%s') "+
			"from service_use "+
			"where apply_status=%d "+
			"group by date_format(`create_at`,'%s'),department) "+
			"as count", "%Y-%m", 1, "%Y-%m")
	} else {
		str1 = fmt.Sprintf("select department,"+
			"date_format(`create_at`,'%s'),"+
			"count(department) from service_use "+
			"where apply_status=%d "+
			"group by date_format(`create_at`,'%s'),department "+
			"order by count(department) desc,"+
			"date_format(`create_at`,'%s') desc limit %d,%d",
			"%Y", 1, "%Y", "%Y",
			(skip-1)*10, 10)
		str2 = fmt.Sprintf("select sum(apply_number*20+5*collect_number),"+
			"date_format(`create_at`, '%s') "+
			"from service_count "+
			"where date_format(`create_at`, '%s') = date_format(now(), '%s') "+
			"and service_name='%s' "+
			"group by date_format(`create_at`, '%s') "+
			"order by date_format(`create_at`, '%s') asc",
			"%Y-%m", "%Y", "%Y", service,
			"%Y-%m", "%Y-%m")
		str3 = fmt.Sprintf("select count(*) from "+
			"(select department,"+
			"date_format(`create_at`,'%s') "+
			"from service_use "+
			"where apply_status=%d "+
			"group by date_format(`create_at`,'%s'),department) "+
			"as count", "%Y", 1, "%Y")
	}
	db = conf.DB.Raw(str3)
	db.Row().Scan(&res.Total)

	db = conf.DB.Raw(str1)
	rows, err := db.Rows()
	if err != nil {
		log.Error("计算部门每年申请数量失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "统计申请数量失败!",
		})
		return
	}

	for rows.Next() {
		var rank m.Rank
		var date string
		rows.Scan(&rank.Dept, &date, &rank.Count)

		ranks = append(ranks, &rank)
	}

	db = conf.DB.Raw(str2)
	rows, err = db.Rows()
	if err != nil {
		log.Error("计算每月申请数失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "统计每月申请数量失败!",
		})
		return
	}

	for rows.Next() {
		var mon m.ApplyStats
		rows.Scan(&mon.Count, &mon.Date)

		month = append(month, &mon)
	}

	res.Apply = month
	res.Rank = ranks

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    res,
		"message": "获取统计结果成功!",
	})

}

// 服务修改 同 GetOneInfo->ModifyInfo

/*
	K3-服务使用申请
	1.查询服务使用申请信息
	2.查看申请详情
	3.查看驳回原因
	4.取消申请
*/
func (g *Gin) GetUseApplication(c *gin.Context) {
	var req a.UseRequest
	var display, sql1, sql2 string
	var total int64
	var uses = []*a.List{}

	err := c.ShouldBind(&req)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "数据解析失败!",
		})
		return
	}

	// 解析 token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				display = v.(map[string]interface{})["displayName"].(string)
				//mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}

	if req.Sort.Sort == 0 {
		if req.Type == 0 {
			sql1 = fmt.Sprintf("select count(*) from service_use "+
				"where applicant = '%s'", display)

			sql2 = fmt.Sprintf("select id,email,applicant,serial_number,"+
				"service_name,service_id,project_name,charge,apply_time,status from service_use "+
				"where applicant='%s' "+
				"and applicant like '%s' and email like '%s' "+
				"and charge like '%s' and serial_number like '%s' "+
				"and service_name like '%s' and project_name like '%s' "+
				"order by apply_time desc limit %d,%d",
				display, "%"+req.Applicant+"%", "%"+req.Mail+"%", "%"+req.Charge+"%",
				"%"+req.Serial+"%", "%"+req.Service+"%", "%"+req.Project+"%",
				(req.Skip-1)*req.Limit, req.Limit)

		} else {

			sql1 = fmt.Sprintf("select count(*) from service_use "+
				"where applicant = '%s' and status=%d",
				display, req.Type)

			sql2 = fmt.Sprintf("select id,email,applicant,serial_number,"+
				"service_name,service_id,project_name,charge,apply_time,status from service_use "+
				"where applicant = '%s' and status=%d "+
				"and email like '%s' "+
				"and applicant like '%s' and serial_number like '%s' "+
				"and service_name like '%s' and project_name like '%s' "+
				"and charge like '%s' "+
				"order by apply_time desc limit %d,%d",
				display, req.Type,
				"%"+req.Mail+"%", "%"+req.Applicant+"%",
				"%"+req.Serial+"%", "%"+req.Service+"%",
				"%"+req.Project+"%", "%"+req.Charge+"%",
				(req.Skip-1)*req.Limit, req.Limit)

		}
	} else {
		if req.Type == 0 {
			sql1 = fmt.Sprintf("select count(*) from service_use "+
				"where charge like '%s'", "%"+display+"%")

			sql2 = fmt.Sprintf("select id,email,applicant,serial_number,"+
				"service_name,service_id,project_name,charge,apply_time,status from service_use "+
				"where applicant = '%s' and email like '%s' "+
				"and applicant like '%s' and serial_number like '%s' "+
				"and service_name like '%s' and project_name like '%s' "+
				"and charge like '%s' "+
				"order by apply_time asc limit %d,%d",
				display, "%"+req.Mail+"%", "%"+req.Applicant+"%",
				"%"+req.Serial+"%", "%"+req.Service+"%", "%"+req.Project+"%",
				"%"+req.Charge+"%", (req.Skip-1)*req.Limit, req.Limit)

		} else {

			sql1 = fmt.Sprintf("select count(*) from service_use "+
				"where charge like '%s' and status=%d",
				"%"+display+"%", req.Type)

			sql2 = fmt.Sprintf("select id,email,applicant,serial_number,"+
				"service_name,project_name,charge,apply_time,status from service_use "+
				"where applicant = '%s' and status=%d "+
				"and email like '%s' "+
				"and applicant like '%s' and serial_number like '%s' "+
				"and service_name like '%s' and project_name like '%s' "+
				"and charge like '%s' "+
				"order by apply_time asc limit %d,%d",
				display, req.Type,
				"%"+req.Mail+"%", "%"+req.Applicant+"%",
				"%"+req.Serial+"%", "%"+req.Service+"%",
				"%"+req.Project+"%", "%"+req.Charge+"%",
				(req.Skip-1)*req.Limit, req.Limit)

		}
	}

	db := conf.DB.Raw(sql1)
	db.Row().Scan(&total)

	var pass, reject int64
	db = conf.DB.Raw("select count(*) from receive "+
		"where msg_type=? and status=? "+
		"and apply_status=?", "use", 1, 1)
	db.Row().Scan(&pass)
	db = conf.DB.Raw("select count(*) from receive "+
		"where msg_type=? and status=? "+
		"and apply_status=?", "use", 1, 2)
	db.Row().Scan(&reject)

	db = conf.DB.Raw(sql2)
	rows, err := db.Rows()
	if err != nil {
		log.Error("查询数据失败: ", err)
		return
	}

	for rows.Next() {
		var use a.List
		rows.Scan(&use.Id, &use.Mail, &use.Applicant, &use.Serial,
			&use.Service, &use.ServiceId, &use.Project, &use.Charge,
			&use.ApplyTime, &use.Status)

		db = conf.DB.Raw("select id,status from receive "+
			"where service_id=? and msg_type=? "+
			"order by receive_time desc limit 1",
			use.ServiceId, "use")
		db.Row().Scan(&use.MsgId, &use.MsgStatus)

		uses = append(uses, &use)
	}
	rows.Close()

	data := a.UseResponse{}
	data.List = uses
	data.Total = total
	data.Pass = pass
	data.Reject = reject

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    data,
		"message": "获取数据成功!",
	})
}

// 查看申请详情 用 D1-GetApplyDetail
// 取消申请    用 D1-CancelApply

/*
	L1-广告管理
*/
func (g *Gin) AdManage(c *gin.Context) {
	var res ad.AdList
	var lists = []*ad.AdInfo{}
	skip, _ := strconv.Atoi(c.Query("skip"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	db := conf.DB.Raw("select count(*) from ad")
	db.Row().Scan(&res.Total)

	db = conf.DB.Raw("select id,position,remark,update_at "+
		"from ad limit ?,?", (skip-1)*limit, limit)
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取广告信息失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取广告信息失败!",
		})
		return
	}

	for rows.Next() {
		var tm time.Time
		var list ad.AdInfo
		rows.Scan(&list.Id, &list.Position, &list.Remark, &tm)

		list.UpdateTime = tm.Format("2006-01-02 15:04:05")

		lists = append(lists, &list)
	}
	rows.Close()

	res.List = lists

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    res,
		"message": "success!",
	})
}

func (g *Gin) GetAd(c *gin.Context) {
	var cont []ad.Content
	var adDetail adDetail
	var str string
	id := c.Query("id")

	db := conf.DB.Raw("select id,position,remark,content from ad "+
		"where id=?", id)
	err := db.Error
	if err != nil {
		log.Error("获取广告详情失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取广告详情失败!",
		})
		return
	}

	db.Row().Scan(&adDetail.Id, &adDetail.Position, &adDetail.Remark, &str)
	json.Unmarshal([]byte(str), &cont)

	adDetail.Content = cont

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    adDetail,
		"message": "获取广告信息成功!",
	})
}

func (g *Gin) AdEdit(c *gin.Context) {
	var adInfo adDetail
	err := c.ShouldBind(&adInfo)
	if err != nil {
		log.Error("请求失败: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    err,
			"message": "数据解析失败!",
		})
		return
	}

	obj, _ := json.Marshal(adInfo.Content)

	db := conf.DB.Exec("update ad set update_at=?,remark=?,content=? "+
		"where id=?",
		time.Now(), adInfo.Remark, string(obj), adInfo.Id)
	err = db.Error
	if err != nil {
		log.Error("更新失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "更新失败!",
		})
		return
	}

	log.Info("更新成功: ", db.RowsAffected)

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    "",
		"message": "更新成功!",
	})
}

/*
	L2-留言管理 (发布者，发布者领导，公司领导，系统管理员)
*/
func (g *Gin) ShowMessage(c *gin.Context) {
	var req lm.MsgReq
	var res lm.MessageList
	var lists = []*lm.Msg{}
	var display, col, order, s, ss string

	err := c.ShouldBind(&req)
	if err != nil {
		log.Error("请求失败: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    err,
			"message": "数据解析失败!",
		})
		return
	}

	// 解析 token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				display = v.(map[string]interface{})["displayName"].(string)
				//mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}

	col = "bbs.write_time"
	//col="write_time"

	if req.Sort.Sort == 0 {
		order = "desc"
	} else {
		order = "asc"
	}

	// show (0:显示，1:屏蔽)
	// read (0:已读，1:未读)

	if req.Show == 2 {
		s = fmt.Sprintf("select count(*) from "+
			"(select service_id from service_bbs "+
			"where service_name like '%s' and msg_from like '%s') as bbs "+
			"left join "+
			"(select id from service_app where user='%s') as app "+
			"on bbs.service_id=app.id",
			"%"+req.ServiceName+"%", "%"+req.Commenter+"%", display)
		ss = fmt.Sprintf("select * from "+
			"(select id,service_id,service_name,"+
			"msg_from,content,write_time,`read`,`show`,"+
			"reply from service_bbs where service_name like '%s' "+
			"and msg_from like '%s') as bbs "+
			"left join "+
			"(select id from service_app where user='%s') as app "+
			"on bbs.service_id=app.id "+
			"order by %s limit %d,%d",
			"%"+req.ServiceName+"%", "%"+req.Commenter+"%", display,
			col+" "+order, (req.Skip-1)*req.Limit, req.Limit)

	} else {
		s = fmt.Sprintf("select count(*) from "+
			"(select service_id from service_bbs "+
			"where service_name like '%s' "+
			"and msg_from like '%s' and `show` like %d) as bbs "+
			"left join "+
			"(select id from service_app where user='%s') as app "+
			"on bbs.service_id=app.id",
			"%"+req.ServiceName+"%", "%"+req.Commenter+"%",
			req.Show, display)
		ss = fmt.Sprintf("select * from "+
			"(select id,service_id,service_name,"+
			"msg_from,content,write_time,`read`,`show`,"+
			"reply from service_bbs where service_name like '%s' "+
			"and msg_from like '%s' and `show` like %d) as bbs "+
			"left join "+
			"(select id from service_app where user='%s') as app "+
			"on bbs.service_id=app.id "+
			"order by %s limit %d,%d",
			"%"+req.ServiceName+"%", "%"+req.Commenter+"%",
			req.Show, display,
			col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
	}

	db := conf.DB.Raw(s)
	err = db.Row().Scan(&res.Total)
	if err != nil {
		log.Error("获取总数失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取总数失败!",
		})
		return
	}
	var count int64
	db = conf.DB.Raw("select count(*) from service_bbs "+
		"where `read`=? and `show`=?", 1, 0)
	db.Row().Scan(&count)

	db = conf.DB.Raw(ss)
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取数据失败!",
		})
		return
	}
	for rows.Next() {
		var list lm.Msg
		var id int
		rows.Scan(&list.Id, &list.ServiceId, &list.ServiceName, &list.Commenter,
			&list.Content, &list.MsgTime, &list.Read, &list.Show,
			&list.Reply, &id)

		lists = append(lists, &list)
	}
	rows.Close()

	res.List = lists
	res.New = count

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    res,
		"message": "获取数据成功!",
	})
}

// 答复留言
func (g *Gin) Reply(c *gin.Context) {
	var data lm.Reply
	err := c.ShouldBind(&data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "数据解析失败!",
		})
		return
	}

	respondent := "系统管理员"
	data.ReplyTime = time.Now().Format("2006-01-02 15:04:05")

	db := conf.DB.Exec("update service_bbs set update_at=?,"+
		"respondent=?,reply=?,reply_time=? where id=?", time.Now(),
		respondent, data.Reply, data.ReplyTime, data.Id)
	err = db.Error
	if err != nil {
		log.Error("答复信息失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "答复信息失败!",
		})

		return
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    "",
		"message": "答复成功!",
	})
}

func (g *Gin) Show(c *gin.Context) {
	var req lm.Show
	err := c.ShouldBind(&req)
	if err != nil {
		log.Error("获取请求失败: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    err,
			"message": "数据解析失败",
		})
		return
	}

	for _, v := range req.Ids {
		db := conf.DB.Exec("update service_bbs set update_at=?,"+
			"`show`=? where id=?", time.Now(), req.Show, v)
		err := db.Error
		if err != nil {
			log.Error("更改状态失败: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "更改状态失败!",
			})
			return
		}

		log.Info("更改成功: ", db.RowsAffected)
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    "",
		"message": "更改状态成功!",
	})
}

func (g *Gin) Read(c *gin.Context) {
	var req lm.Read
	err := c.ShouldBind(&req)
	if err != nil {
		log.Error("获取请求失败: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    err,
			"message": "数据解析失败",
		})
		return
	}

	for _, v := range req.Ids {
		db := conf.DB.Exec("update service_bbs set update_at=?,"+
			"`read`=? where id=?", time.Now(), req.Read, v)
		err := db.Error
		if err != nil {
			log.Error("更改状态失败: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "更改状态失败!",
			})
			return
		}

		log.Info("更改成功: ", db.RowsAffected)
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    "",
		"message": "更改状态成功!",
	})
}

// 查看留言信息
func (g *Gin) MsgContent(c *gin.Context) {
	var msg lm.MsgDetail
	var user string
	var reply sql.NullString
	id := c.Query("id")

	db := conf.DB.Raw("select id,msg_from,tel,service_name,"+
		"service_id,write_time,content,reply,reply_time,`show` "+
		"from service_bbs "+
		"where id=?", id)
	err := db.Error
	if err != nil {
		log.Error("获取数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取数据失败!",
		})
		return
	}

	db.Row().Scan(&msg.Id, &msg.Commenter, &msg.Tel, &msg.ServiceName,
		&msg.ServiceId, &msg.MsgTime, &msg.Content,
		&reply, &msg.ReplyTime, &msg.Show)
	if reply.Valid {
		msg.Reply = reply.String
	} else {
		msg.Reply = ""
	}
	db = conf.DB.Raw("select user from service_app "+
		"where id=?", msg.ServiceId)
	db.Row().Scan(&user)

	msg.Respondent = strings.Split(user, "(")[0]

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    msg,
		"message": "获取数据成功!",
	})
}

/*
	L3-消息管理
*/
func (g *Gin) SysMsg(c *gin.Context) {
	var req m.SysReq
	var res m.MsgList
	var lists = []*m.SysMsg{}
	var col, order, str1, str2 string
	err := c.ShouldBind(&req)
	if err != nil {
		log.Error("请求失败: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    err,
			"message": "数据解析失败!",
		})
		return
	}

	// 解析 token
	//tk := c.Request.Header.Get("Authorization")
	//log.Info("Token: ", tk)
	//if tk == "null" {
	//	c.JSON(401, gin.H{
	//		"code":    http.StatusUnauthorized,
	//		"data":    "",
	//		"message": "token验证失败",
	//	})
	//	return
	//} else {
	//	mp, err := util.AnalysisToken(tk)
	//	if err != nil {
	//		log.Error("解析token出错: ", err)
	//		return
	//	}
	//	for k, v := range mp {
	//		if v == nil{
	//			c.JSON(401,gin.H{
	//				"code":http.StatusUnauthorized,
	//				"message":"token已过期",
	//			})
	//			return
	//		}
	//		if k == "data" {
	//			display = v.(map[string]interface{})["displayName"].(string)
	//			mail = v.(map[string]interface{})["email"].(string)
	//			//dept = v.(map[string]interface{})["department"].(string)
	//		}
	//	}
	//}

	if req.Sort.Column == "sendTime" {
		col = "state_time"
	} else {
		col = "create_time"
	}

	if req.Sort.Sort == 0 {
		order = "desc"
	} else {
		order = "asc"
	}

	if req.Status == 0 {
		str1 = fmt.Sprintf("select count(*) from sys_message "+
			"where creator like '%s' "+
			"and mail like '%s'",
			"%"+req.Creator+"%", "%"+req.Mail+"%")
		str2 = fmt.Sprintf("select id,content,creator,mail,create_time,"+
			"state_time,status from sys_message "+
			"where creator like '%s' "+
			"and mail like '%s' order by %s limit %d,%d",
			"%"+req.Creator+"%", "%"+req.Mail+"%",
			col+" "+order, (req.Skip-1)*req.Limit, req.Limit)
	} else {
		str1 = fmt.Sprintf("select count(*) from sys_message "+
			"where creator like '%s' "+
			"and mail like '%s' and status like %d ",
			"%"+req.Creator+"%", "%"+req.Mail+"%", req.Status)
		str2 = fmt.Sprintf("select id,content,creator,mail,create_time,"+
			"state_time,status from sys_message "+
			"where creator like '%s' "+
			"and mail like '%s' and status like %d "+
			"order by %s limit %d,%d",
			"%"+req.Creator+"%", "%"+req.Mail+"%",
			req.Status, col+" "+order,
			(req.Skip-1)*req.Limit, req.Limit)
	}
	db := conf.DB.Raw(str1)
	db.Row().Scan(&res.Total)

	db = conf.DB.Raw(str2)
	rows, err := db.Rows()
	if err != nil {
		log.Error("获取数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取数据失败!",
		})
		return
	}
	for rows.Next() {
		var list m.SysMsg
		rows.Scan(&list.Id, &list.Content, &list.Creator, &list.Mail,
			&list.CreateTime, &list.SendTime, &list.Status)

		lists = append(lists, &list)
	}
	rows.Close()

	res.List = lists

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    res,
		"message": "获取数据成功!",
	})

}

func (g *Gin) NewMsg(c *gin.Context) {
	var msg m.NewMsg
	var status int
	var display, mail, tm string
	var send []string

	err := c.ShouldBind(&msg)
	if err != nil {
		log.Error("请求失败: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    err,
			"message": "数据解析失败!",
		})
		return
	}

	// 解析 token
	tk := c.Request.Header.Get("Authorization")
	log.Info("Token: ", tk)
	if tk == "null" {
		c.JSON(401, gin.H{
			"code":    http.StatusUnauthorized,
			"data":    "",
			"message": "token验证失败",
		})
		return
	} else {
		mp, err := util.AnalysisToken(tk)
		if err != nil {
			log.Error("解析token出错: ", err)
			return
		}
		for k, v := range mp {
			if v == nil {
				c.JSON(401, gin.H{
					"code":    http.StatusUnauthorized,
					"message": "token已过期",
				})
				return
			}
			if k == "data" {
				display = v.(map[string]interface{})["displayName"].(string)
				mail = v.(map[string]interface{})["email"].(string)
				//dept = v.(map[string]interface{})["department"].(string)
			}
		}
	}

	if msg.Send == true {
		status = 2
		tm = time.Now().Format("2006-01-02 15:04:05")
		db := conf.DB.Exec("insert into sys_message set create_at=?,"+
			"update_at=?,creator=?,content=?,mail=?,create_time=?,"+
			"state_time=?,status=?", time.Now(), time.Now(),
			display, msg.Content, mail,
			tm, tm, status)
		err = db.Error
		if err != nil {
			log.Error("新增消息失败: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "新增消息失败!",
			})
			return
		}

		log.Info("新增消息成功: ", db.RowsAffected)

		c.JSON(200, gin.H{
			"code":    http.StatusOK,
			"data":    "",
			"message": "success!",
		})

		// 发送消息
		var data SystemMsg
		db = conf.DB.Raw("select distinct mail " +
			"from service_app ")
		rows, err := db.Rows()
		if err != nil {
			log.Error("获取邮箱失败: ", err)
			return
		}
		for rows.Next() {
			var mail string
			rows.Scan(&mail)
			send = append(send, mail)
		}

		db = conf.DB.Raw("select distinct email " +
			"from service_use")
		rows, err = db.Rows()
		if err != nil {
			log.Error("获取邮箱失败: ", err)
			return
		}
		for rows.Next() {
			var mail string
			rows.Scan(&mail)
			send = append(send, mail)
		}
		send = util.Distinct(send)
		fmt.Println("send to: ", send)

		data.Theme = "系统消息"
		data.SendTo = send
		data.Content = msg.Content
		data.MsgTime = time.Now().Unix()
		pub := map[string]interface{}{
			"level":  1,
			"type":   "system",
			"uid":    mail,
			"dept":   "",
			"status": msgState,
			"data":   data,
		}
		msg := &broker.Message{}
		msg.Body, err = json.Marshal(pub)
		if err != nil {
			log.Error("数据转换失败: ", err)
			return
		}
		if err := broker.Publish(topic, msg); err != nil {
			log.Error("发布信息失败: ", err)
			return
		}

		log.Info("发送信息成功: ", string(msg.Body))

		return

	} else {
		status = 1
		//if msg.SendTime == 0{
		//	status = 1
		//}else{
		//	status = 2
		//}
		ct := time.Now().Format("2006-01-02 15:04:05")
		tm = time.Unix(msg.SendTime, 0).Format("2006-01-02 15:04:05")
		db := conf.DB.Exec("insert into sys_message set create_at=?,"+
			"update_at=?,creator=?,content=?,mail=?,create_time=?,"+
			"state_time=?,status=?", time.Now(), time.Now(),
			display, msg.Content, mail, ct, tm, status)
		err = db.Error
		if err != nil {
			log.Error("新增消息失败: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "新增消息失败!",
			})
			return
		}

		log.Info("新增消息成功: ", db.RowsAffected)
		var id int
		db = conf.DB.Raw("select id from sys_message "+
			"where create_time=?", ct)
		db.Row().Scan(&id)
		log.Info("ID: ", id)

		c.JSON(200, gin.H{
			"code":    http.StatusOK,
			"data":    "",
			"message": "success!",
		})

		// 定时发送
		go func() {
			st := time.Unix(msg.SendTime, 0)
			sub := st.Sub(time.Now())
			select {
			case <-time.Tick(sub):
				// 发送消息
				var data SystemMsg
				db = conf.DB.Raw("select distinct mail " +
					"from service_app ")
				rows, err := db.Rows()
				if err != nil {
					log.Error("获取邮箱失败: ", err)
					return
				}
				for rows.Next() {
					var mail string
					rows.Scan(&mail)
					send = append(send, mail)
				}

				db = conf.DB.Raw("select distinct email " +
					"from service_use")
				rows, err = db.Rows()
				if err != nil {
					log.Error("获取邮箱失败: ", err)
					return
				}
				for rows.Next() {
					var mail string
					rows.Scan(&mail)
					send = append(send, mail)
				}
				send = util.Distinct(send)
				fmt.Println("send to: ", send)

				data.Theme = "系统消息"
				data.SendTo = send
				data.Content = msg.Content
				data.MsgTime = msg.SendTime
				pub := map[string]interface{}{
					"level":  1,
					"type":   "system",
					"uid":    mail,
					"dept":   "",
					"status": msgState,
					"data":   data,
				}
				msg := &broker.Message{}
				msg.Body, err = json.Marshal(pub)
				if err != nil {
					log.Error("数据转换失败: ", err)
					return
				}
				if err := broker.Publish(topic, msg); err != nil {
					log.Error("发布信息失败: ", err)
					return
				}

				log.Info("发送信息成功: ", string(msg.Body))

				// 将发送状态改为已发送
				db = conf.DB.Exec("update sys_message set update_at=?,"+
					"status=2 where id=?", time.Now(), id)
				log.Info("更新成功: ", db.RowsAffected, id)
			}
		}()

	}

}

func (g *Gin) MsgStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("id"))
	status, _ := strconv.Atoi(c.Query("status"))

	db := conf.DB.Exec("update sys_message set update_at=?,"+
		"state_time=?,status=? where id=?", time.Now(),
		time.Now().Format("2006-01-02 15:04:05"),
		status, id)
	err := db.Error
	if err != nil {
		log.Error("更新状态失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "更新状态失败!",
		})
		return
	}
	log.Info(db.RowsAffected)
	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    db.RowsAffected,
		"message": "更新状态成功!",
	})
}

func (g *Gin) GetOneSysMsg(c *gin.Context) {
	var one m.SysMsgInfo
	var creator, mail string
	id := c.Query("id")
	db := conf.DB.Raw("select id,creator,mail,create_time,"+
		"state_time,status,content from sys_message "+
		"where id=?", id)
	err := db.Error
	if err != nil {
		log.Error("获取数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"data":    err,
			"message": "获取数据失败!",
		})
		return
	}

	db.Row().Scan(&one.Id, &creator, &mail, &one.CreateTime,
		&one.StateTime, &one.Status, &one.Content)
	one.Creator = creator + "(" + mail + ")"

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    one,
		"message": "查询信息成功!",
	})
}

/*
	M-统计看板
*/
func (g *Gin) BoardHeader(c *gin.Context) {
	var header BoardHeader
	// 获取发布占比 申请占比
	{
		var ratio []Ratio
		var count int
		db := conf.DB.Raw("select count(*) from service_app")
		db.Row().Scan(&count)
		log.Info("发布总数: ", count)
		header.PubTotal = count

		db = conf.DB.Raw("select department,count(department) " +
			"from service_app group by department " +
			"order by count(department) desc")
		rows, err := db.Rows()
		if err != nil {
			log.Error("统计发布数据失败: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"error":   err,
				"message": "获取统计数据失败!",
			})
			return
		}

		for rows.Next() {
			var ra Ratio
			rows.Scan(&ra.Dept, &ra.Count)

			//ra.Ratio = float64(ra.Count) / float64(count)
			result := float64(ra.Count) / float64(count)
			ra.Ratio = math.Trunc(result*math.Pow10(4)+0.5) / math.Pow10(4)

			ratio = append(ratio, ra)
		}

		header.PubRatio = ratio
	}

	{
		var ratio []Ratio
		var count int
		db := conf.DB.Raw("select count(*) from service_use")
		db.Row().Scan(&count)
		log.Info("申请总数: ", count)
		header.ApplyTotal = count

		db = conf.DB.Raw("select department,count(department) " +
			"from service_use where apply_status=1 " +
			"group by department " +
			"order by count(department) desc")
		rows, err := db.Rows()
		if err != nil {
			log.Error("统计申请数据失败: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"error":   err,
				"message": "获取统计数据失败!",
			})
			return
		}

		for rows.Next() {
			var ra Ratio
			rows.Scan(&ra.Dept, &ra.Count)

			//ra.Ratio = float64(ra.Count) / float64(count)
			result := float64(ra.Count) / float64(count)
			ra.Ratio = math.Trunc(result*math.Pow10(4)+0.5) / math.Pow10(4)

			ratio = append(ratio, ra)
		}

		header.ApplyRatio = ratio
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    header,
		"message": "获取比率成功",
	})
}

func (g *Gin) StatisticsBoard(c *gin.Context) {
	var board Board
	var str1, str2, str3, str4, str5, str6 string
	appCycle := c.Query("appCycle")
	pubCycle := c.Query("pubCycle")
	pubSkip, _ := strconv.Atoi(c.Query("pubSkip"))
	appSkip, _ := strconv.Atoi(c.Query("appSkip"))

	// 按照 周，月，年 统计，周-7天，月-30/31天，年-12个月
	// 发布 app
	if pubCycle == "week" {
		// %Y-%m-%d
		str1 = fmt.Sprintf("select count(*),"+
			"date_format(`create_at`, '%s') "+
			"from service_app "+
			"where yearweek(date_format(`create_at`, '%s')) = yearweek(now()) "+
			"group by date_format(`create_at`, '%s') "+
			"order by date_format(`create_at`, '%s') asc",
			"%Y-%m-%d", "%Y-%m-%d", "%Y-%m-%d", "%Y-%m-%d")
		// 发布排名
		str5 = fmt.Sprintf("select department,date_format(`create_at`,'%s'),"+
			"count(department) from service_app "+
			"where yearweek(date_format(`create_at`, '%s')) = yearweek(now()) "+
			"and status in (5,6) "+
			"group by week(`create_at`),department "+
			"order by count(department) desc,"+
			"week(`create_at`) desc limit %d,%d",
			"%Y-%m-%d", "%Y-%m-%d", (pubSkip-1)*10, 10)
		str6 = fmt.Sprintf("select count(*) from "+
			"(select department,date_format(`create_at`,'%s') "+
			"from service_app "+
			"where yearweek(date_format(`create_at`, '%s')) = yearweek(now()) "+
			"and status in (5,6) "+
			"group by week(`create_at`),department) as c",
			"%Y-%m-%d", "%Y-%m-%d")
	} else if pubCycle == "month" {
		str1 = fmt.Sprintf("select count(*),"+
			"date_format(`create_at`, '%s') "+
			"from service_app "+
			"where date_format(`create_at`, '%s') = date_format(now(), '%s') "+
			"group by date_format(`create_at`, '%s') "+
			"order by date_format(`create_at`, '%s') asc",
			"%Y-%m-%d", "%Y-%m", "%Y-%m", "%Y-%m-%d", "%Y-%m-%d")
		// publish app
		str5 = fmt.Sprintf("select department,"+
			"date_format(`create_at`,'%s'),"+
			"count(department) from service_app "+
			"where status in(5,6) "+
			"group by date_format(`create_at`,'%s'),department "+
			"order by count(department) desc,"+
			"date_format(`create_at`,'%s') desc limit %d,%d",
			"%Y-%m", "%Y-%m", "%Y-%m", (pubSkip-1)*10, 10)
		str6 = fmt.Sprintf("select count(*) from "+
			"(select department,"+
			"date_format(`create_at`,'%s') "+
			"from service_app "+
			"where status in(5,6) "+
			"group by date_format(`create_at`,'%s'),department) "+
			"as c", "%Y-%m", "%Y-%m")
	} else {
		str1 = fmt.Sprintf("select count(*),"+
			"date_format(`create_at`, '%s') "+
			"from service_app "+
			"where date_format(`create_at`, '%s') = date_format(now(), '%s') "+
			"group by date_format(`create_at`, '%s') "+
			"order by date_format(`create_at`, '%s') asc",
			"%Y-%m", "%Y", "%Y", "%Y-%m", "%Y-%m")
		// publish app
		str5 = fmt.Sprintf("select department,"+
			"date_format(`create_at`,'%s'),"+
			"count(department) from service_app "+
			"where status in(5,6) "+
			"group by date_format(`create_at`,'%s'),department "+
			"order by count(department) desc,"+
			"date_format(`create_at`,'%s') desc limit %d,%d",
			"%Y", "%Y", "%Y", (pubSkip-1)*10, 10)
		str6 = fmt.Sprintf("select count(*) from "+
			"(select department,"+
			"date_format(`create_at`,'%s') "+
			"from service_app "+
			"where status in(5,6) "+
			"group by date_format(`create_at`,'%s'),department) "+
			"as c", "%Y", "%Y")
	}

	if appCycle == "week" {
		// %Y-%m-%d
		str2 = fmt.Sprintf("select count(*),"+
			"date_format(`create_at`, '%s') "+
			"from service_use "+
			"where yearweek(date_format(`create_at`, '%s')) = yearweek(now()) "+
			"and apply_status=1 "+
			"group by date_format(`create_at`, '%s') "+
			"order by date_format(`create_at`, '%s') asc",
			"%Y-%m-%d", "%Y-%m-%d", "%Y-%m-%d", "%Y-%m-%d")
		// 使用申请排名
		str3 = fmt.Sprintf("select department,date_format(`create_at`,'%s'),"+
			"count(department) from service_use "+
			"where yearweek(date_format(`create_at`, '%s')) = yearweek(now()) "+
			"and apply_status=%d "+
			"group by week(`create_at`),department "+
			"order by count(department) desc,"+
			"week(`create_at`) desc limit %d,%d",
			"%Y-%m-%d", "%Y-%m-%d", 1, (appSkip-1)*10, 10)
		str4 = fmt.Sprintf("select count(*) from "+
			"(select department,date_format(`create_at`,'%s'),"+
			"count(department) from service_use "+
			"where yearweek(date_format(`create_at`, '%s')) = yearweek(now())"+
			"and apply_status=%d "+
			"group by week(`create_at`),department) as count",
			"%Y-%m-%d", "%Y-%m-%d", 1)
	} else if appCycle == "month" {
		str2 = fmt.Sprintf("select count(*),"+
			"date_format(`create_at`, '%s') "+
			"from service_use "+
			"where date_format(`create_at`, '%s') = date_format(now(), '%s') "+
			"and apply_status=1 "+
			"group by date_format(`create_at`, '%s') "+
			"order by date_format(`create_at`, '%s') asc",
			"%Y-%m-%d", "%Y-%m", "%Y-%m", "%Y-%m-%d", "%Y-%m-%d")
		// application use
		str3 = fmt.Sprintf("select department,"+
			"date_format(`create_at`,'%s'),"+
			"count(department) from service_use "+
			"where apply_status=%d "+
			"group by date_format(`create_at`,'%s'),department "+
			"order by count(department) desc,"+
			"date_format(`create_at`,'%s') desc limit %d,%d",
			"%Y-%m", 1, "%Y-%m", "%Y-%m", (appSkip-1)*10, 10)
		str4 = fmt.Sprintf("select count(*) from "+
			"(select department,"+
			"date_format(`create_at`,'%s') "+
			"from service_use "+
			"where apply_status=%d "+
			"group by date_format(`create_at`,'%s'),department) "+
			"as c", "%Y-%m", 1, "%Y-%m")
	} else {
		str2 = fmt.Sprintf("select count(*),"+
			"date_format(`create_at`, '%s') "+
			"from service_use "+
			"where date_format(`create_at`, '%s') = date_format(now(), '%s') "+
			"group by date_format(`create_at`, '%s') "+
			"order by date_format(`create_at`, '%s') asc",
			"%Y-%m", "%Y", "%Y", "%Y-%m", "%Y-%m")
		// application use
		str3 = fmt.Sprintf("select department,"+
			"date_format(`create_at`,'%s'),"+
			"count(department) from service_use "+
			"where apply_status=%d "+
			"group by date_format(`create_at`,'%s'),department "+
			"order by count(department) desc,"+
			"date_format(`create_at`,'%s') desc limit %d,%d",
			"%Y", 1, "%Y", "%Y", (appSkip-1)*10, 10)
		str4 = fmt.Sprintf("select count(*) from "+
			"(select department,"+
			"date_format(`create_at`,'%s') "+
			"from service_use "+
			"where apply_status=%d "+
			"group by date_format(`create_at`,'%s'),department) "+
			"as c", "%Y", 1, "%Y")
	}
	// 获取发布的统计数据
	var num1 []Number
	db := conf.DB.Raw(str1)
	rows, err := db.Rows()
	if err != nil {
		log.Error("统计数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "获取统计数据失败!",
		})
		return
	}
	for rows.Next() {
		var n Number
		rows.Scan(&n.Count, &n.Date)

		num1 = append(num1, n)
	}

	board.Publish = num1

	// 获取申请的统计数据
	var num2 []Number
	db = conf.DB.Raw(str2)
	rows, err = db.Rows()
	if err != nil {
		log.Error("统计数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "获取统计数据失败!",
		})
		return
	}
	for rows.Next() {
		var n Number
		rows.Scan(&n.Count, &n.Date)

		num2 = append(num2, n)
	}

	board.Apply = num2

	// 获取部门申请排名
	var rank1 = []*m.Rank{}
	db = conf.DB.Raw(str3)
	rows, err = db.Rows()
	if err != nil {
		log.Error("统计数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "获取统计数据失败!",
		})
		return
	}
	for rows.Next() {
		var r m.Rank
		var date string
		rows.Scan(&r.Dept, &date, &r.Count)

		rank1 = append(rank1, &r)
	}

	var ar deptRank
	ar.Rank = rank1
	db = conf.DB.Raw(str4)
	db.Row().Scan(&ar.Total)
	board.ApplyRank = ar

	// 获取部门发布排名
	var rank2 = []*m.Rank{}
	db = conf.DB.Raw(str5)
	rows, err = db.Rows()
	if err != nil {
		log.Error("统计数据失败: ", err)
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"error":   err,
			"message": "获取统计数据失败!",
		})
		return
	}
	for rows.Next() {
		var r m.Rank
		var date string
		rows.Scan(&r.Dept, &date, &r.Count)

		rank2 = append(rank2, &r)
	}

	var pr deptRank
	pr.Rank = rank2
	db = conf.DB.Raw(str6)
	db.Row().Scan(&pr.Total)
	board.PubRank = pr

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    board,
		"message": "获取统计信息成功!",
	})

}

/*
	公共接口:上传文档、图片、视频
*/
func (g *Gin) UploadImage(c *gin.Context) {
	var mail string
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "获取数据失败!",
		})
		return
	}

	// 获取文件名
	fileName := header.Filename
	str := strings.Split(fileName, ".")
	layout := strings.ToLower(str[len(str)-1])
	if layout != "jpeg" && layout != "png" && layout != "jpg" && layout != "gif" {
		log.Error("文件格式不正确!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   "",
			"message": "文件格式不正确!",
		})
		return
	}

	if header.Size > 10000000 {
		//判断大小是否大于10M
		log.Error("文件过大!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "文件大于10M，请重新上传",
		})
		return
	}

	t := strconv.FormatInt(time.Now().Unix(), 10)
	filePath := "static/image/" + t + "-" + mail + "-" + fileName

	out, err := os.Create(filePath)
	if err != nil {
		log.Error("创建文件失败!")
		c.JSON(500, gin.H{
			"code":    500,
			"data":    err,
			"message": "创建文件失败!",
		})
		return
	}

	_, err = io.Copy(out, file)
	if err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    500,
			"data":    err,
			"message": "保存文件失败!",
		})
		return
	}
	file.Close()
	out.Close()

	c.JSON(200, gin.H{
		"code":    200,
		"data":    filePath,
		"message": "上传图片成功!",
	})

}

func (g *Gin) UploadDoc(c *gin.Context) {
	var mail string
	//layout != "doc" && layout != "docx" && layout != "pdf"
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "获取数据失败!",
		})
		return
	}

	// 获取文件名
	fileName := header.Filename
	str := strings.Split(fileName, ".")
	layout := strings.ToLower(str[len(str)-1])
	if layout != "doc" && layout != "docx" && layout != "pdf" &&
		layout != "ppt" && layout != "pptx" && layout != "zip" &&
		layout != "rar" && layout != "7z" &&
		layout != "xls" && layout != "xlsx" {
		log.Error("文件格式不正确!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   "",
			"message": "文件格式不正确!",
		})
		return
	}

	if header.Size > 10000000 {
		//判断大小是否大于10M
		log.Error("文件过大!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "文件大于10M，请重新上传",
		})
		return
	}

	t := strconv.FormatInt(time.Now().Unix(), 10)
	filePath := "static/file/" + t + "-" + mail + "-" + fileName

	out, err := os.Create(filePath)
	if err != nil {
		log.Error("创建文件失败!")
		c.JSON(500, gin.H{
			"code":    500,
			"data":    err,
			"message": "创建文件失败!",
		})
		return
	}

	_, err = io.Copy(out, file)
	if err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    500,
			"data":    err,
			"message": "保存文件失败!",
		})
		return
	}
	file.Close()
	out.Close()

	c.JSON(200, gin.H{
		"code":    200,
		"data":    filePath,
		"message": "上传文档成功!",
	})
}

func (g *Gin) UploadVideo(c *gin.Context) {
	var mail string
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "获取数据失败!",
		})
		return
	}

	// 获取文件名
	fileName := header.Filename
	str := strings.Split(fileName, ".")
	layout := strings.ToLower(str[len(str)-1])
	if layout != "mp4" && layout != "flv" && layout != "rmb" &&
		layout != "mpg" && layout != "avi" && layout != "mov" {
		log.Error("文件格式不正确!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   "",
			"message": "文件格式不正确!",
		})
		return
	}

	if header.Size > 500000000 {
		//判断大小是否大于500M
		log.Error("文件过大!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"data":    "",
			"message": "文件大于500M，请重新上传",
		})
		return
	}

	t := strconv.FormatInt(time.Now().Unix(), 10)
	filePath := "static/video/" + t + "-" + mail + "-" + fileName

	out, err := os.Create(filePath)
	if err != nil {
		log.Error("创建文件失败!")
		c.JSON(500, gin.H{
			"code":    500,
			"data":    err,
			"message": "创建文件失败!",
		})
		return
	}

	_, err = io.Copy(out, file)
	if err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    500,
			"data":    err,
			"message": "保存文件失败!",
		})
		return
	}
	file.Close()
	out.Close()

	c.JSON(200, gin.H{
		"code":    200,
		"data":    filePath,
		"message": "上传视频成功!",
	})

}

// 下载文档，图片
func (g *Gin) Download(c *gin.Context) {
	file := c.DefaultQuery("file", "")
	if file == "" {
		log.Error("文件路径及名称为空!")
		c.JSON(200, gin.H{
			"code":    400,
			"message": "文件路径及名称为空!",
		})
		return
	}

	c.Writer.Header().Add("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", file))
	c.Writer.Header().Add("Content-Type", "application/octet-stream")
	c.File(file)
}

// 删除文件
func (g *Gin) DeleteDocument(c *gin.Context) {
	var req delReq
	var file, image, video string
	var doc, img, vd []m.Document

	err := c.ShouldBind(&req)
	if err != nil {
		log.Error("获取数据失败: ", err)
		log.WithFields(log.Fields{
			"err": err,
		}).Error("获取请求失败!")

		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"error":   err,
			"message": "数据解析失败!",
		})
		return
	}

	l1 := len(req.File)
	l2 := len(req.Image)
	l3 := len(req.Video)

	if l1 == 0 && l2 == 0 && l3 == 0 {
		log.Info("没有需要删除的文件!")
		c.JSON(200, gin.H{
			"code":    http.StatusOK,
			"data":    "",
			"message": "没有需要删除的文件!",
		})
		return
	}

	if req.Id != 0 {
		// 进行修改,需要修改数据库
		db := conf.DB.Raw("select document,image,video from service_app "+
			"where id=? ", req.Id)
		db.Row().Scan(&file, &image, &video)

		err := json.Unmarshal([]byte(file), &doc)
		if err != nil {
			log.Info("111111: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    err,
				"message": "删除文件失败!",
			})
			return
		}
		err = json.Unmarshal([]byte(image), &img)
		if err != nil {
			if err != nil {
				log.Info("222222: ", err)
				c.JSON(500, gin.H{
					"code":    http.StatusInternalServerError,
					"data":    err,
					"message": "删除文件失败!",
				})
				return
			}
		}

		err = json.Unmarshal([]byte(video), &vd)
		if err != nil {
			if err != nil {
				log.Info("333333: ", err)
				c.JSON(500, gin.H{
					"code":    http.StatusInternalServerError,
					"data":    err,
					"message": "删除文件失败!",
				})
				return
			}
		}

		for _, vv := range req.File {
			for n, v := range doc {
				if vv == v.Url {
					doc = append(doc[:n], doc[n+1:]...)
				}
			}

			log.Info("文件: ", doc)

			exist, err := util.PathExists(vv)
			if err != nil {
				log.Error(err)
				return
			}

			if exist {
				err := os.Remove(vv)
				if err != nil {
					log.Error("删除文件失败: ", err)
					return
				}
			} else {
				log.Info("文件不存在!")
				c.JSON(500, gin.H{
					"code":    http.StatusInternalServerError,
					"data":    "文件不存在!",
					"message": "删除文件失败!",
				})
				return
			}
		}

		for _, vv := range req.Image {
			for n, v := range img {
				if vv == v.Url {
					img = append(img[:n], img[n+1:]...)
				}
			}

			exist, err := util.PathExists(vv)
			if err != nil {
				log.Error(err)
				return
			}

			if exist {
				err := os.Remove(vv)
				if err != nil {
					log.Error("删除文件失败: ", err)
					return
				}
			} else {
				log.Info("文件不存在!")
				return
			}
		}

		for _, vv := range req.Video {
			for n, v := range vd {
				if vv == v.Url {
					vd = append(vd[:n], vd[n+1:]...)
				}
			}

			exist, err := util.PathExists(vv)
			if err != nil {
				log.Error(err)
				return
			}

			if exist {
				err := os.Remove(vv)
				if err != nil {
					log.Error("删除文件失败: ", err)
					return
				}
			} else {
				log.Info("文件不存在!")
				return
			}
		}

		d1, _ := json.Marshal(doc)
		d2, _ := json.Marshal(img)
		d3, _ := json.Marshal(vd)

		db = conf.DB.Exec("update service_app set update_at=?,document=?,"+
			"image=?,video=? where id=?", time.Now(),
			string(d1), string(d2), string(d3), req.Id)
		err = db.Error
		if err != nil {
			log.Error("更新数据失败: ", err)
			c.JSON(500, gin.H{
				"code":    http.StatusInternalServerError,
				"data":    "文件不存在!",
				"message": "删除文件失败!",
			})
			return
		}
		log.Info("更新成功: ", db.RowsAffected)
	} else {
		for _, v := range req.File {
			exist, err := util.PathExists(v)
			if err != nil {
				log.Error(err)
				return
			}

			if exist {
				err := os.Remove(v)
				if err != nil {
					log.Error("删除文件失败: ", err)
					return
				}
			} else {
				log.Info("文件不存在!")
				c.JSON(500, gin.H{
					"code":    http.StatusInternalServerError,
					"data":    "文件不存在!",
					"message": "删除文件失败!",
				})
				return
			}
		}

		for _, v := range req.Image {
			exist, err := util.PathExists(v)
			if err != nil {
				log.Error(err)
				return
			}

			if exist {
				err := os.Remove(v)
				if err != nil {
					log.Error("删除文件失败: ", err)
					return
				}
			} else {
				log.Info("文件不存在!")
				return
			}
		}

		for _, v := range req.Video {
			exist, err := util.PathExists(v)
			if err != nil {
				log.Error(err)
				return
			}

			if exist {
				err := os.Remove(v)
				if err != nil {
					log.Error("删除文件失败: ", err)
					return
				}
			} else {
				log.Info("文件不存在!")
				return
			}
		}
	}

	c.JSON(200, gin.H{
		"code":    http.StatusOK,
		"data":    "",
		"message": "删除文件成功!",
	})
}

func (g *Gin) PubSignResult(c *gin.Context) {
	var body, operation, remark interface{}
	var mail, dept, service, serial, info string
	var id int
	err := c.ShouldBind(&body)
	if err != nil {
		log.Error("请求解析失败: ", err)
		return
	}

	res := body.(map[string]interface{})
	for _, v := range res {
		info := v.([]interface{})
		d := info[0].(map[string]interface{})
		mail = d["邮箱号"].(string)
		service = d["服务名称"].(string)
		//dept=d["部门"].(string)
	}
	fmt.Println("mail: ", mail, " service: ", service)
	db := conf.DB.Raw("select id,department,serial_number,apply_info from service_app "+
		"where mail=? and service_name=?",
		mail, service)
	db.Row().Scan(&id, &dept, &serial, &info)
	fmt.Println(serial)

	wf := workflow.NewWorkflowClient()
	token, err := wf.GetToken("180435", "Wpf!2345")
	if err != nil {
		log.Error("获取 token 失败: ", err)
		return
	}
	log.Info("token: ", token)

	WFID := strconv.Itoa(conf.Sc.PubWFID)
	url := "http://10.2.44.61/wfapi/api/Business?wfID=" + WFID + "&businessID=" + serial
	request, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		log.Error("请求失败!")

		return
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("appID", conf.Sc.Appid)
	request.Header.Add("appKey", conf.Sc.Appkey)
	request.Header.Add("token", token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Error("服务器错误: ", err)
		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("获取响应失败: ", err)

		return
	}
	result := make(map[string]interface{})

	json.Unmarshal(data, &result)
	log.Info("123456: ", result)

	for k, v := range result {
		if k == "Operations" {
			data := v.([]interface{})
			//opt1:=data[0].(map[string]interface{})
			//tag=opt1["WFNodeName"].(string)
			opt := data[len(data)-1].(map[string]interface{})
			operation = opt["Operation"].(string)
			remark = opt["Remark"]
		}
	}

	log.Info("审批结果: ", operation)
	log.Info("审批意见: ", remark)

	// 更新
	var apply a.Apply
	var status int
	json.Unmarshal([]byte(info), &apply)
	if operation == "通过" {
		apply.Status = 1
		status = 5
	} else if operation == "驳回" {
		apply.Status = 2
		status = 2
		if remark != nil {
			apply.Data = remark.(string)
		} else {
			apply.Data = ""
		}

	} else {
		return
	}

	apply.StateTime = time.Now().Format("2006-01-02 15:04:05")

	fmt.Println("apply info: ", apply)
	appInfo, _ := json.Marshal(apply)

	db = conf.DB.Exec("update service_app set update_at=?,"+
		"status=?,apply_info=? where id=?", time.Now(),
		status, string(appInfo), id)
	log.Info("已更新: ", db.RowsAffected, " 条")

	// 发送消息，发送到 ws 中
	var sMsg ServiceMsg
	msgType := "publish"

	sMsg.StateTime = apply.StateTime
	sMsg.Service = service
	sMsg.Status = int(apply.Status)
	sMsg.Data = apply.Data
	sMsg.Content = service + "签审结束，签审结果为:" + operation.(string)
	pub := map[string]interface{}{
		"level":  0,
		"type":   msgType,
		"uid":    mail,
		"dept":   dept,
		"status": 1,
		"data":   sMsg,
	}
	msg := &broker.Message{}
	msg.Body, err = json.Marshal(pub)
	if err != nil {
		log.Error("数据转换失败: ", err)
		return
	}
	if err := broker.Publish(topic, msg); err != nil {
		log.Error("发布信息失败: ", err)
		return
	}

	log.Info("发送信息成功: ", string(msg.Body))
}

func (g *Gin) UseSignResult(c *gin.Context) {
	var body, operation, remark, mail, service interface{}
	var serial, info string
	var id, sId int
	err := c.ShouldBind(&body)
	if err != nil {
		log.Error("请求解析失败: ", err)
		return
	}

	res := body.(map[string]interface{})
	for _, v := range res {
		info := v.([]interface{})
		d := info[0].(map[string]interface{})
		mail = d["申请人邮箱号"].(string)
		service = d["服务名称"].(string)
	}
	fmt.Println("mail: ", mail, " service: ", service)

	db := conf.DB.Raw("select id,service_id,serial_number,apply_info from service_use "+
		"where email=? and service_name=?",
		mail, service)
	db.Row().Scan(&id, &sId, &serial, &info)

	wf := workflow.NewWorkflowClient()
	token, err := wf.GetToken("180435", "Wpf!2345")
	if err != nil {
		log.Error("获取 token 失败: ", err)
		return
	}
	log.Info("token: ", token)

	WFID := strconv.Itoa(conf.Sc.UseWFID)
	url := "http://10.2.44.61/wfapi/api/Business?wfID=" + WFID + "&businessID=" + serial
	request, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		log.Error("请求失败!")

		return
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("appID", conf.Sc.Appid)
	request.Header.Add("appKey", conf.Sc.Appkey)
	request.Header.Add("token", token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Error("服务器错误: ", err)

		return
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("获取响应失败: ", err)

		return
	}
	result := make(map[string]interface{})

	json.Unmarshal(data, &result)
	log.Info("使用申请回调结果: ", result)

	for k, v := range result {
		if k == "Operations" {
			data := v.([]interface{})
			//opt1:=data[0].(map[string]interface{})
			//tag=opt1["WFNodeName"].(string)
			opt := data[len(data)-1].(map[string]interface{})
			operation = opt["Operation"].(string)
			remark = opt["Remark"]
		}
	}

	log.Info("审批结果: ", operation)
	log.Info("审批意见: ", remark)

	// 更新
	var useDetail a.ApplyDetail
	var status, applyStatus int
	json.Unmarshal([]byte(info), &useDetail)

	if operation == "通过" {
		useDetail.Status = 1
		status = 1
		applyStatus = 1
	} else if operation == "驳回" {
		useDetail.Status = 2
		applyStatus = 0
		status = 2
		if remark != nil {
			useDetail.Data = remark.(string)
		} else {
			useDetail.Data = ""
		}
	} else {
		return
	}

	useDetail.StateTime = time.Now().Format("2006-01-02 15:04:05")
	useApply, _ := json.Marshal(useDetail)

	db = conf.DB.Exec("update service_use set update_at=?,"+
		"status=?,apply_status=?,apply_info=? where id=?", time.Now(),
		status, applyStatus, string(useApply), id)
	log.Info("已更新: ", db.RowsAffected, " 条")

	// 发送消息，发送到 ws 中
	var sMsg ServiceMsg
	msgType := "use"

	sMsg.StateTime = useDetail.StateTime
	sMsg.Service = service.(string)
	sMsg.Status = status
	sMsg.Data = useDetail.Data
	sMsg.Content = service.(string) + "签审结束，签审结果为:" + operation.(string)
	pub := map[string]interface{}{
		"level":  0,
		"type":   msgType,
		"uid":    mail,
		"status": 1,
		"data":   sMsg,
	}
	msg := &broker.Message{}
	msg.Body, err = json.Marshal(pub)
	if err != nil {
		log.Error("数据转换失败: ", err)
		return
	}
	if err := broker.Publish(topic, msg); err != nil {
		log.Error("发布信息失败: ", err)
		return
	}

	log.Info("发送信息成功: ", string(msg.Body))
}