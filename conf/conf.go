package conf

import (
	"github.com/jinzhu/gorm"
	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source/file"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlConf struct {
	Dsn     string `json:"dsn"`
	MaxIdle int    `json:"maxIdle"`
	MaxOpen int    `json:"maxOpen"`
}

type Sign struct {
	Appid     string `json:"appid"`
	Appkey    string `json:"appkey"`
	GetToken  string `json:"getToken"`
	Submit    string `json:"submit"`
	UseWFID   int    `json:"useWFID"`
	UseId     int    `json:"useId"`
	UseCancel int    `json:"useCancel"`
	PubWFID   int    `json:"pubWFID"`
	PubId     int    `json:"pubId"`
	PubCancel int    `json:"pubCancel"`
}

//type Redis struct {
//	Addr     string `json:"addr"`
//	Password string `json:"password"`
//	DB       int    `json:"db"`
//	PoolSize int    `json:"poolSize"`
//}

var (
	defaultPath = "api"
	m           sync.RWMutex
	mc          MysqlConf
	DB          *gorm.DB
	Sc          Sign
	//RD          Redis
	//RC          *redis.Client
)

func Init() {
	m.Lock()
	defer m.Unlock()

	err := config.Load(file.NewSource(
		file.WithPath("conf/application.yml"),
	))

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("加载配置文件失败!")

		return
	}

	if err := config.Get(defaultPath, "mysql").Scan(&mc); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("读取mysql配置信息失败!")
		return
	}

	if err := config.Get(defaultPath, "sign").Scan(&Sc); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("读取签审平台配置信息失败!")
		return
	}

	log.Info("获取签审平台配置成功！")

	//初始化
	DB, err = gorm.Open("mysql", mc.Dsn)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("打开数据库失败!")
		return
	}

	DB.DB().SetMaxIdleConns(mc.MaxIdle)
	DB.DB().SetMaxOpenConns(mc.MaxOpen)
	DB.DB().SetConnMaxLifetime(30 * time.Second)

	if err := DB.DB().Ping(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("数据库连接失败!")
		return
	}

	log.Info("连接数据库成功!")

	//初始化 redis
	//if err := config.Get(defaultPath, "redis").Scan(&RD); err != nil {
	//	log.WithFields(log.Fields{
	//		"err": err,
	//	}).Error("读取redis配置信息失败!")
	//	return
	//}
	//
	//RC = redis.NewClient(&redis.Options{
	//	Addr:     RD.Addr,
	//	Password: RD.Password,
	//	DB:       RD.DB,
	//	PoolSize: RD.PoolSize,
	//})
	//
	////激活连接
	//if _, err := RC.Ping().Result(); err != nil {
	//	log.Error("Redis 连接失败: ", err)
	//	return
	//}
	//
	//log.Info("Redis 连接成功!")

}
