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

	// 先检查用户是否存在
	user, err := models.GetUserByUsername(req.Username)
	if err != nil {
		logs.Warn("用户登录失败，用户不存在 [username=%s, time=%s]: %v",
			req.Username, "2025-05-14 07:21:42", err)
		c.Data["json"] = utils.ErrorResponse("用户名或密码错误")
		c.ServeJSON()
		return
	}

	// 检查用户状态是否正常
	if user.Status != 1 {
		logs.Warn("被禁用的账户尝试登录 [username=%s, status=%d, time=%s]",
			req.Username, user.Status, "2025-05-14 07:21:42")
		c.Data["json"] = utils.ErrorResponse("账户已被禁用，请联系管理员")
		c.ServeJSON()
		return
	}

	// 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		logs.Warn("用户登录失败，密码错误 [username=%s, time=%s]",
			req.Username, "2025-05-14 07:21:42")
		c.Data["json"] = utils.ErrorResponse("用户名或密码错误")
		c.ServeJSON()
		return
	}

	// 更新最后登录时间
	models.UpdateLastLogin(user.ID)

	// 记录登录成功
	logs.Info("用户登录成功 [username=%s, role=%s, company_id=%d, time=%s]",
		user.Username, user.Role, user.CompanyId, "2025-05-14 07:21:42")

	// 生成令牌
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, user.CompanyId)
	if err != nil {
		logs.Error("生成token失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("生成token失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"token":     token,
		"user_info": models.GetUserInfo(user),
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

	// 检查用户状态是否正常
	if user.Status != 1 {
		logs.Warn("被禁用账户尝试获取信息 [username=%s, id=%d, status=%d, time=%s]",
			user.Username, userID, user.Status, "2025-05-14 07:21:42")
		c.Data["json"] = utils.ErrorResponse("账户已被禁用，请联系管理员")
		c.ServeJSON()
		return
	}

	info := models.GetUserInfo(user)
	c.Data["json"] = utils.SuccessResponse(info)
	c.ServeJSON()
}
