package controllers

import (
	"encoding/json"
	"strconv"

	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// UserController 用户管理控制器
type UserController struct {
	web.Controller
}

// Post 创建用户
// @Title CreateUser
// @Description 创建新用户
// @Param	body	body 	models.User	true	"用户信息"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @router / [post]
func (u *UserController) Post() {
	var user models.User
	err := json.Unmarshal(u.Ctx.Input.RequestBody, &user)
	if err != nil {
		u.Data["json"] = utils.ErrorResponse("无效的请求数据")
		u.ServeJSON()
		return
	}

	// 验证必填字段
	if user.Username == "" || user.Password == "" {
		u.Data["json"] = utils.ErrorResponse("用户名和密码不能为空")
		u.ServeJSON()
		return
	}

	// 添加用户
	uid := models.AddUser(user)
	if uid == "" {
		u.Data["json"] = utils.ErrorResponse("创建用户失败")
		u.ServeJSON()
		return
	}

	u.Data["json"] = utils.SuccessResponse(map[string]string{"uid": uid})
	u.ServeJSON()
}

// GetAll 获取所有用户
// @Title GetAllUsers
// @Description 获取所有用户列表
// @Success 200 {object} utils.Response
// @router / [get]
func (u *UserController) GetAll() {
	// 从身份验证中间件获取角色信息
	role := u.Ctx.Input.GetData("role")

	// 只有超级管理员可以查看所有用户
	if role != nil && role != "super_admin" {
		u.Data["json"] = utils.ForbiddenResponse()
		u.ServeJSON()
		return
	}

	users := models.GetAllUsers()
	count, _ := models.CountUsers()

	u.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"users": users,
		"total": count,
	})
	u.ServeJSON()
}

// Get 根据ID获取用户
// @Title GetUser
// @Description 通过用户ID获取用户信息
// @Param	uid	path 	string	true	"用户ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @router /:uid [get]
func (u *UserController) Get() {
	uid := u.Ctx.Input.Param(":uid")
	if uid == "" {
		u.Data["json"] = utils.ErrorResponse("用户ID不能为空")
		u.ServeJSON()
		return
	}

	user, err := models.GetUser(uid)
	if err != nil {
		u.Data["json"] = utils.ErrorResponse("获取用户失败: " + err.Error())
		u.ServeJSON()
		return
	}

	// 获取完整用户信息，包括公司信息
	userInfo := models.GetUserInfo(user)
	u.Data["json"] = utils.SuccessResponse(userInfo)
	u.ServeJSON()
}

// Put 更新用户
// @Title UpdateUser
// @Description 更新用户信息
// @Param	uid	path 	string	true	"用户ID"
// @Param	body	body 	models.User	true	"用户信息"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @router /:uid [put]
func (u *UserController) Put() {
	uid := u.Ctx.Input.Param(":uid")
	if uid == "" {
		u.Data["json"] = utils.ErrorResponse("用户ID不能为空")
		u.ServeJSON()
		return
	}

	// 从身份验证中间件获取当前用户信息
	currentUserID := u.Ctx.Input.GetData("user_id")
	currentUserRole := u.Ctx.Input.GetData("role")

	// 权限检查：只有超级管理员或用户本人可以修改用户信息
	if currentUserRole != "super_admin" {
		currentIDStr := strconv.Itoa(currentUserID.(int))
		if currentIDStr != uid {
			u.Data["json"] = utils.ForbiddenResponse()
			u.ServeJSON()
			return
		}
	}

	var userUpdate models.User
	err := json.Unmarshal(u.Ctx.Input.RequestBody, &userUpdate)
	if err != nil {
		u.Data["json"] = utils.ErrorResponse("无效的请求数据")
		u.ServeJSON()
		return
	}

	updatedUser, err := models.UpdateUser(uid, &userUpdate)
	if err != nil {
		u.Data["json"] = utils.ErrorResponse("更新用户失败: " + err.Error())
		u.ServeJSON()
		return
	}

	u.Data["json"] = utils.SuccessResponse(updatedUser)
	u.ServeJSON()
}

// Delete 删除用户
// @Title DeleteUser
// @Description 删除用户
// @Param	uid	path 	string	true	"用户ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @router /:uid [delete]
func (u *UserController) Delete() {
	uid := u.Ctx.Input.Param(":uid")
	if uid == "" {
		u.Data["json"] = utils.ErrorResponse("用户ID不能为空")
		u.ServeJSON()
		return
	}

	// 从身份验证中间件获取角色信息
	role := u.Ctx.Input.GetData("role")

	// 只有超级管理员可以删除用户
	if role != "super_admin" {
		u.Data["json"] = utils.ForbiddenResponse()
		u.ServeJSON()
		return
	}

	err := models.DeleteUser(uid)
	if err != nil {
		u.Data["json"] = utils.ErrorResponse("删除用户失败: " + err.Error())
		u.ServeJSON()
		return
	}

	u.Data["json"] = utils.SuccessResponse("删除成功")
	u.ServeJSON()
}

// Login 用户登录
// @Title Login
// @Description 用户登录接口
// @Param	username	query 	string	true	"用户名"
// @Param	password	query 	string	true	"密码"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @router /login [post]
func (u *UserController) Login() {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.Unmarshal(u.Ctx.Input.RequestBody, &loginReq)
	if err != nil {
		u.Data["json"] = utils.ErrorResponse("无效的请求数据")
		u.ServeJSON()
		return
	}

	user, err := models.CheckLogin(loginReq.Username, loginReq.Password)
	if err != nil {
		logs.Error("登录失败: %v", err)
		u.Data["json"] = utils.ErrorResponse("登录失败: " + err.Error())
		u.ServeJSON()
		return
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, user.CompanyID)
	if err != nil {
		logs.Error("生成token失败: %v", err)
		u.Data["json"] = utils.ErrorResponse("生成token失败")
		u.ServeJSON()
		return
	}

	// 返回用户信息和令牌
	userInfo := models.GetUserInfo(user)
	userInfo["token"] = token

	u.Data["json"] = utils.SuccessResponse(userInfo)
	u.ServeJSON()
}

// Logout 用户登出
// @Title Logout
// @Description 用户登出接口
// @Success 200 {object} utils.Response
// @router /logout [post]
func (u *UserController) Logout() {
	// 客户端处理登出逻辑，后端仅返回成功响应
	u.Data["json"] = utils.SuccessResponse("登出成功")
	u.ServeJSON()
}

// MyInfo 获取当前登录用户信息
// @Title GetMyInfo
// @Description 获取当前登录用户的信息
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @router /myinfo [get]
func (u *UserController) MyInfo() {
	// 从中间件获取用户ID
	userID := u.Ctx.Input.GetData("user_id").(int)

	user, err := models.GetUserByID(userID)
	if err != nil {
		u.Data["json"] = utils.ErrorResponse("获取用户信息失败")
		u.ServeJSON()
		return
	}

	info := models.GetUserInfo(user)
	u.Data["json"] = utils.SuccessResponse(info)
	u.ServeJSON()
}
