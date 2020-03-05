package main

import (
	"SoftwareCloud/conf"
	"SoftwareCloud/handler"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/web"

	log "github.com/sirupsen/logrus"
	"net/http"
)

// 处理跨域请求,支持options访问
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}

func main() {
	service := web.NewService(
		web.Name("cloud.micro.api.v1.cloud"),
	)
	service.Init()
	conf.Init()
	go handler.GetArray()
	//go handler.GetPubStatus()
	//go handler.GetUseStatus()

	//创建 RestFul handler
	g := new(handler.Gin)
	router := gin.Default()
	//使用中间件解决跨域问题
	router.Use(Cors())

	// 上传路由
	r1 := router.Group("/v1/cloud/upload")
	{
		r1.POST("/doc", g.UploadDoc)
		r1.POST("/image", g.UploadImage)
		r1.POST("/video", g.UploadVideo)
	}

	// 服务发布,使用路由
	r2 := router.Group("/v1/cloud/apply")
	{
		r2.POST("/publish/application", g.ServiceApply)
		// 签审结果回写
		r2.POST("/publish/return",g.PubSignResult)

		// B-服务详情页
		r2.GET("/detail", g.DetailAndIntro)
		// 获取服务的使用说明，可在线预览文档，可下载资源
		r2.GET("/comment", g.GetLeaveMessage)
		r2.POST("/comment", g.LeaveMessage)
		// 收藏
		r2.PUT("/collect", g.ChangeStatus)

		// C-服务使用申请
		r2.GET("/use", g.GetHeader)
		r2.POST("/use", g.ServiceUseApply)
		r2.POST("/use/return",g.UseSignResult)

		// D1-服务使用申请
		r2.POST("/use/application", g.GetUseApply)
		r2.GET("/use/application", g.GetApplyDetail)
		r2.GET("/use/instruction", g.GetInstruction)
		r2.GET("/use/rejection", g.GetRejectionReason)
		// 取消使用申请
		r2.GET("/use/cancel", g.CancelApply)

		// D2-服务发布申请
		r2.POST("/release", g.GetRelease)
		r2.GET("/release", g.GetApplyInfo)
		r2.GET("/info", g.GetOneInfo)
		r2.PUT("/info", g.ModifyInfo)
		// 取消发布申请操作
		r2.GET("/cancel", g.Unpublish)
	}

	// 管理员部分
	r3 := router.Group("/v1/cloud/manage")
	{
		// K2-服务发布管理
		r3.POST("/info", g.GetServiceInfo)
		// 获取某个服务的信息，然后进行服务的修改
		r3.GET("/info", g.GetOneInfo)
		r3.PUT("/info", g.ModifyInfo)
		r3.GET("/applyinfo", g.GetApplyInfo)
		// 获取服务的详细介绍，使用说明，在线预览文档以及下载相关资料
		r3.GET("/detail", g.GetServiceDetail)
		// 服务的统计
		r3.GET("/pv",g.StatisticsHeader)
		r3.GET("/count", g.Statistics)

		// k3-服务使用申请
		r3.POST("/use", g.GetUseApplication)
		r3.GET("/use", g.GetApplyDetail)
		r3.GET("/use/cancel", g.CancelApply)
	}

	r4 := router.Group("/v1/cloud")
	{
		// A-首页
		r4.POST("/home", g.Home)
		// A2-服务列表
		r4.GET("/service/list",g.ServiceListHeader)
		r4.POST("/service/list", g.ServiceList)
		r4.GET("/", g.OnlinePreview)
		r4.GET("/download", g.Download)
		r4.DELETE("/remove", g.DeleteDocument)
		//上下架
		r4.PUT("/status", g.AlterStatus)

		// 广告
		r4.GET("/ad/list", g.AdManage)
		r4.GET("/ad", g.GetAd)
		r4.PUT("/ad", g.AdEdit)

		// 留言
		r4.POST("/leaveMessage", g.ShowMessage)
		r4.GET("/leaveMessage", g.MsgContent)
		r4.POST("/reply", g.Reply)
		r4.PUT("/leaveMessage/show", g.Show)
		r4.PUT("/leaveMessage/read", g.Read)

		// 消息管理
		r4.POST("/message/list", g.SysMsg)
		r4.POST("/message", g.NewMsg)
		r4.PUT("/message", g.MsgStatus)
		r4.GET("/message", g.GetOneSysMsg)

		// 我的收藏
		r4.GET("/myCollection", g.MyCollection)

		// 统计看板
		r4.GET("/board/header",g.BoardHeader)
		r4.GET("/board", g.StatisticsBoard)

		// 个人中心
		r4.GET("/personalCenter",g.GetMessage)
		r4.GET("/personalCenter/content",g.DisplayMessage)
		// 消息中心
		r4.GET("/messageCenter",g.MsgCenter)
		r4.PUT("/messageCenter",g.MsgCenterStatus)
	}

	// 注册 handler
	service.Handle("/", router)

	// 运行 api
	err := service.Run()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("服务启动错误!")
		return
	}
}
