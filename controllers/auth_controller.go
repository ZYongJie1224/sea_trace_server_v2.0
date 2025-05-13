package controllers

import (
	"encoding/json"

	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// AuthController 认证控制器
type AuthController struct {
	web.Controller
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login 用户登录
// @router /api/auth/login [post]
func (c *AuthController) Login() {
	var req LoginRequest

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	user, err := models.CheckLogin(req.Username, req.Password)
	if err != nil {
		logs.Error("登录失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("登录失败: " + err.Error())
		c.ServeJSON()
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, user.CompanyID)
	if err != nil {
		logs.Error("生成token失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("生成token失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(map[string]string{
		"token": token,
	})
	c.ServeJSON()
}

// MyInfo 获取用户信息
// @router /api/auth/myinfo [get]
func (c *AuthController) MyInfo() {
	// 从中间件获取用户ID
	userID := c.Ctx.Input.GetData("user_id").(int)

	user, err := models.GetUserByID(userID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取用户信息失败")
		c.ServeJSON()
		return
	}

	info := models.GetUserInfo(user)
	c.Data["json"] = utils.SuccessResponse(info)
	c.ServeJSON()
}
