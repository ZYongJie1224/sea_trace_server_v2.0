package controllers

import (
	"encoding/json"
	"strconv"

	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// UserManagementController 用户管理控制器
type UserManagementController struct {
	web.Controller
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Keyword  string `json:"keyword"`
	Role     string `json:"role"`
}

// ListUsers 获取用户列表
// @Title 获取用户列表
// @Description 获取用户列表，支持分页、搜索和角色筛选
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认10"
// @Param search query string false "搜索关键词"
// @Param role query string false "角色筛选"
// @Success 200 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/admin/users [get]
func (c *UserManagementController) ListUsers() {
	// 检查权限
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户角色")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok || (role != "super_admin" && role != "company_admin") {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 获取查询参数
	page, _ := c.GetInt("page", 1)
	pageSize, _ := c.GetInt("page_size", 10)
	keyword := c.GetString("search", "")
	roleFilter := c.GetString("role", "")

	// 公司管理员只能查看自己公司的用户
	companyID := 0
	if role == "company_admin" {
		companyIDInterface := c.Ctx.Input.GetData("company_id")
		if companyIDInterface == nil {
			c.Data["json"] = utils.ErrorResponse("无法获取公司ID")
			c.ServeJSON()
			return
		}

		var ok bool
		companyID, ok = companyIDInterface.(int)
		if !ok {
			c.Data["json"] = utils.ErrorResponse("公司ID类型错误")
			c.ServeJSON()
			return
		}
	}

	// 调用模型层获取数据
	var users []*models.User
	var total int64
	var err error

	if companyID > 0 {
		users, total, err = models.GetUserList(page, pageSize, keyword, roleFilter, companyID)
	} else {
		users, total, err = models.GetUserList(page, pageSize, keyword, roleFilter)
	}

	if err != nil {
		logs.Error("获取用户列表失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("获取用户列表失败")
		c.ServeJSON()
		return
	}

	// 返回结果
	response := map[string]interface{}{
		"users":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}

	c.Data["json"] = utils.SuccessResponse(response)
	c.ServeJSON()
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	RealName  string `json:"real_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	CompanyID int    `json:"company_id"`
}

// CreateUser 创建用户
// @router /api/admin/user/create [post]
func (c *UserManagementController) CreateUser() {
	// 检查权限
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户角色")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok || role != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	var req CreateUserRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 验证必填字段
	if req.Username == "" || req.Password == "" || req.Role == "" {
		c.Data["json"] = utils.ErrorResponse("用户名、密码和角色不能为空")
		c.ServeJSON()
		return
	}

	// 验证角色
	if req.Role != "super_admin" && req.Role != "company_admin" && req.Role != "operator" {
		c.Data["json"] = utils.ErrorResponse("无效的角色")
		c.ServeJSON()
		return
	}

	// 如果不是超级管理员，需要验证公司ID
	if req.Role != "super_admin" && req.CompanyID <= 0 {
		c.Data["json"] = utils.ErrorResponse("公司管理员和操作员需要指定公司ID")
		c.ServeJSON()
		return
	}

	// 创建用户
	user, err := models.CreateUser(
		req.Username,
		req.Password,
		req.RealName,
		req.Role,
		req.CompanyID,
		req.Email,
		req.Phone,
	)

	if err != nil {
		logs.Error("创建用户失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("创建用户失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 更新额外信息
	if req.Email != "" || req.Phone != "" {
		user.Email = req.Email
		user.Phone = req.Phone

		o := models.GetOrm()
		_, err = o.Update(user, "Email", "Phone")
		if err != nil {
			logs.Error("更新用户额外信息失败: %v", err)
		}
	}

	c.Data["json"] = utils.SuccessResponse(user)
	c.ServeJSON()
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	RealName  string `json:"real_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    int    `json:"status"`
	Password  string `json:"password"`
	CompanyID int    `json:"company_id"`
}

// UpdateUser 更新用户信息
// @router /api/admin/user/update/:id [put]
func (c *UserManagementController) UpdateUser() {
	// 检查权限
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户角色")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok || role != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 获取用户ID
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的用户ID")
		c.ServeJSON()
		return
	}

	// 获取用户信息
	user, err := models.GetUserByID(id)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("用户不存在")
		c.ServeJSON()
		return
	}

	var req UpdateUserRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 更新字段
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Status != 0 {
		user.Status = req.Status
	}
	if req.CompanyID > 0 && user.Role != "super_admin" {
		user.CompanyId = req.CompanyID
	}

	// 更新密码
	if req.Password != "" {
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			c.Data["json"] = utils.ErrorResponse("密码加密失败")
			c.ServeJSON()
			return
		}
		user.Password = hashedPassword
	}

	// 保存更新
	o := models.GetOrm()
	_, err = o.Update(user)
	if err != nil {
		logs.Error("更新用户失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("更新用户失败")
		c.ServeJSON()
		return
	}

	// 返回更新后的用户信息
	userInfo := models.GetUserInfo(user)
	c.Data["json"] = utils.SuccessResponse(userInfo)
	c.ServeJSON()
}

// DeleteUser 删除用户
// @router /api/admin/user/delete/:id [delete]
func (c *UserManagementController) DeleteUser() {
	// 检查权限
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户角色")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok || role != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 获取用户ID
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的用户ID")
		c.ServeJSON()
		return
	}

	// 不能删除自己
	currentUserID := c.Ctx.Input.GetData("user_id").(int)
	if id == currentUserID {
		c.Data["json"] = utils.ErrorResponse("不能删除当前登录用户")
		c.ServeJSON()
		return
	}

	// 检查是否为超级管理员
	user, err := models.GetUserByID(id)
	if err == nil && user.Role == "super_admin" {
		// 检查是否为最后一个超级管理员
		adminCount, _ := models.CountAdmins()
		if adminCount <= 1 {
			c.Data["json"] = utils.ErrorResponse("系统必须保留至少一个超级管理员")
			c.ServeJSON()
			return
		}
	}

	// 删除用户
	err = models.DeleteUserByID(id)
	if err != nil {
		logs.Error("删除用户失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("删除用户失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse("删除成功")
	c.ServeJSON()
}
