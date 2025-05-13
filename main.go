package main

import (
	"sea_trace_server_V2.0/models"
	_ "sea_trace_server_V2.0/routers"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	// 注册数据库
	orm.RegisterDriver("mysql", orm.DRMySQL)

	// 从配置读取数据库连接
	dataSource, _ := web.AppConfig.String("sqlconn")
	err := orm.RegisterDataBase("default", "mysql", dataSource)
	if err != nil {
		logs.Error("数据库连接失败:", err)
		return
	}

	// 注册模型
	orm.RegisterModel(new(models.User), new(models.Company), new(models.Goods))

	// 自动创建表（开发模式）
	orm.RunSyncdb("default", false, true)

	// 日志配置
	logs.SetLogger(logs.AdapterFile, `{"filename":"logs/app.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10}`)
}

func main() {
	// 开启 session
	web.BConfig.WebConfig.Session.SessionOn = true

	// 设置静态文件
	web.SetStaticPath("/static", "static")

	// 日志级别
	logs.SetLevel(logs.LevelDebug)
	logs.Info("启动应用服务...")

	// 运行应用
	web.Run()
}
