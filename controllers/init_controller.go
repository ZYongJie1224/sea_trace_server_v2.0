package controllers

import (
	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// InitController 初始化控制器
type InitController struct {
	web.Controller
}

// InitAdmin 初始化超级管理员
// @router /api/init/admin [get]
func (c *InitController) InitAdmin() {
	o := orm.NewOrm()

	// 检查是否已存在超级管理员
	adminCount, err := o.QueryTable(new(models.User)).Filter("role", "super_admin").Count()
	if err != nil {
		logs.Error("查询管理员失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("查询管理员失败")
		c.ServeJSON()
		return
	}

	// 如果已存在则拒绝
	if adminCount > 0 {
		c.Data["json"] = utils.ErrorResponse("已存在超级管理员账户，无法再次初始化")
		c.ServeJSON()
		return
	}

	// 创建超级管理员
	password := "admin123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		logs.Error("生成密码哈希失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("生成密码哈希失败")
		c.ServeJSON()
		return
	}

	admin := &models.User{
		Username: "admin",
		Password: hashedPassword,
		RealName: "超级管理员",
		Role:     "super_admin",
		Status:   1,
	}

	_, err = o.Insert(admin)
	if err != nil {
		logs.Error("创建管理员失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("创建管理员失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(map[string]string{
		"username": "admin",
		"password": password,
		"message":  "超级管理员创建成功",
	})
	c.ServeJSON()
}
