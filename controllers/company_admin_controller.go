package controllers

import (
	"encoding/json"
	"strconv"
	"time"

	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// CompanyAdminController 公司管理员控制器
type CompanyAdminController struct {
	web.Controller
}

// UpdateCompanyInfoRequest 更新公司信息请求
type UpdateCompanyInfoRequest struct {
	Address string `json:"address"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
}

// CreateOperatorRequest 创建操作员请求
type CreateOperatorRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"` // 修复: 添加了缺失的引号
}

// UpdateOperatorStatusRequest 更新操作员状态请求
type UpdateOperatorStatusRequest struct {
	Status int `json:"status"`
}

// UpdateOperatorInfoRequest 更新操作员信息请求
type UpdateOperatorInfoRequest struct {
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// CompanyInfo 获取公司信息
// @router /api/admin/company/info [get]
func (c *CompanyAdminController) CompanyInfo() {
	// 从Context获取公司ID
	// 首先从 token 中获取公司ID
	companyIDInterface := c.Ctx.Input.GetData("company_id")
	var companyID int
	var ok bool

	// 如果 token 中没有公司ID，则尝试从请求参数中获取
	if companyIDInterface == nil {
		// 从请求参数中获取公司ID
		companyIDParam := c.GetString("company_id")
		if companyIDParam != "" {
			// 将参数转换为整数
			paramID, err := strconv.Atoi(companyIDParam)
			if err == nil {
				companyID = paramID
				ok = true
				logs.Info("用户 [%v] 通过请求参数指定公司ID: %d",
					c.Ctx.Input.GetData("username"), paramID)
			}
		}
	} else {
		companyID, ok = companyIDInterface.(int)
	}

	if !ok {
		logs.Warning("用户 [%v] 无法获取公司ID，可能未关联公司", c.Ctx.Input.GetData("username"))
		c.Data["json"] = utils.ErrorResponse("无法获取公司ID，您可能尚未关联到任何公司")
		c.ServeJSON()
		return
	}

	// 检查公司ID是否有效
	if companyID <= 0 {
		logs.Warning("用户 [%v] 关联的公司ID无效: %d", c.Ctx.Input.GetData("username"), companyID)
		c.Data["json"] = utils.ErrorResponse("您尚未关联到有效公司，请联系管理员")
		c.ServeJSON()
		return
	}

	company, err := models.GetCompanyByID(companyID)
	if err != nil {
		logs.Error("获取公司信息失败 [companyID=%d, user=%v, time=%s]: %v",
			companyID, c.Ctx.Input.GetData("username"), "2025-05-14 06:58:52", err)
		c.Data["json"] = utils.ErrorResponse("无法找到对应的公司信息，请联系系统管理员")
		c.ServeJSON()
		return
	}

	// 构建详细的公司信息响应
	companyInfo := map[string]interface{}{
		"id":                company.ID,
		"company_name":      company.CompanyName,
		"company_type":      int(company.CompanyType),
		"company_type_name": models.CompanyTypeMap[company.CompanyType],
		"address":           company.Address,
		"contact":           company.Contact,
		"phone":             company.Phone,
		"created_at":        company.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":        company.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 获取操作员数量
	operatorCount, err := models.CountCompanyOperators(companyID)
	if err != nil {
		logs.Warning("获取操作员数量失败 [companyID=%d]: %v", companyID, err)
		operatorCount = 0
	}

	// 获取货物数量
	goodCount, err := models.CountCompanyGoods(companyID)
	if err != nil {
		logs.Warning("获取货物数量失败 [companyID=%d]: %v", companyID, err)
		goodCount = 0
	}
	operators, _, _ := models.GetCompanyOperators(companyID, 1, 20, "")
	// 返回完整的公司信息
	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"company":        companyInfo,
		"operator_count": operatorCount,
		"good_count":     goodCount,
		"last_updated":   "2025-05-14 06:58:52",
		"current_user":   c.Ctx.Input.GetData("username"),
		"operators":      operators,
	})
	c.ServeJSON()
}

// UpdateCompanyInfo 更新公司信息
// @router /api/admin/company/info [put]
func (c *CompanyAdminController) UpdateCompanyInfo() {
	companyIDInterface := c.Ctx.Input.GetData("company_id")
	roleInterface := c.Ctx.Input.GetData("role")

	if companyIDInterface == nil || roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户信息")
		c.ServeJSON()
		return
	}

	companyID, ok := companyIDInterface.(int)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("公司ID类型错误")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("角色类型错误")
		c.ServeJSON()
		return
	}

	// 公司管理员才能修改
	if role != "company_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	company, err := models.GetCompanyByID(companyID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取公司信息失败")
		c.ServeJSON()
		return
	}

	var req UpdateCompanyInfoRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	company.Address = req.Address
	company.Contact = req.Contact
	company.Phone = req.Phone

	if err := models.UpdateCompany(company); err != nil {
		logs.Error("更新公司信息失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("更新公司信息失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(company)
	c.ServeJSON()
}

// CreateOperator 创建操作员
// @router /api/admin/company/operator/create [post]
func (c *CompanyAdminController) CreateOperator() {
	companyIDInterface := c.Ctx.Input.GetData("company_id")
	roleInterface := c.Ctx.Input.GetData("role")

	if companyIDInterface == nil || roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户信息")
		c.ServeJSON()
		return
	}

	companyID, ok := companyIDInterface.(int)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("公司ID类型错误")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("角色类型错误")
		c.ServeJSON()
		return
	}

	// 公司管理员才能创建操作员
	if role != "company_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	var req CreateOperatorRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 验证必填字段
	if req.Username == "" || req.Password == "" {
		c.Data["json"] = utils.ErrorResponse("用户名和密码不能为空")
		c.ServeJSON()
		return
	}

	// 创建操作员
	user, err := models.CreateUser(
		req.Username,
		req.Password,
		req.RealName,
		"operator",
		companyID,
		req.Email,
		req.Phone,
	)

	if err != nil {
		logs.Error("创建操作员失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("创建操作员失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 返回用户信息（不包含密码）
	userInfo := models.GetUserInfo(user)
	c.Data["json"] = utils.SuccessResponse(userInfo)
	c.ServeJSON()
}

// DeleteOperator 删除操作员
// @router /api/admin/company/operator/delete/:id [delete]
func (c *CompanyAdminController) DeleteOperator() {
	companyIDInterface := c.Ctx.Input.GetData("company_id")
	roleInterface := c.Ctx.Input.GetData("role")

	if companyIDInterface == nil || roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户信息")
		c.ServeJSON()
		return
	}

	companyID, ok := companyIDInterface.(int)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("公司ID类型错误")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("角色类型错误")
		c.ServeJSON()
		return
	}

	// 公司管理员才能删除操作员
	if role != "company_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 获取操作员ID
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的操作员ID")
		c.ServeJSON()
		return
	}

	// 确保操作员属于当前公司
	user, err := models.GetUserByID(id)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("操作员不存在")
		c.ServeJSON()
		return
	}

	// 检查操作员是否属于当前公司，以及角色是否为操作员
	if user.CompanyId != companyID || user.Role != "operator" {
		c.Data["json"] = utils.ErrorResponse("操作员不存在或不属于当前公司")
		c.ServeJSON()
		return
	}

	// 删除操作员
	if err := models.DeleteUserByID(id); err != nil {
		logs.Error("删除操作员失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("删除操作员失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(nil)
	c.ServeJSON()
}

// GetOperators 获取公司操作员列表
// @Title 获取公司操作员列表
// @Description 获取公司操作员列表，支持分页和搜索
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认10"
// @Param search query string false "搜索关键词(用户名/姓名/邮箱/电话)"
// @Success 200 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 500 {object} utils.Response
// @router /api/admin/company/operators [get]
func (c *CompanyAdminController) GetOperators() {
	companyIDInterface := c.Ctx.Input.GetData("company_id")
	roleInterface := c.Ctx.Input.GetData("role")

	if companyIDInterface == nil || roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户信息")
		c.ServeJSON()
		return
	}

	companyID, ok := companyIDInterface.(int)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("公司ID类型错误")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("角色类型错误")
		c.ServeJSON()
		return
	}

	// 只有公司管理员或超级管理员可以查看操作员列表
	if role != "company_admin" && role != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 如果是超级管理员且没有指定公司ID，则设置为0表示查询所有公司
	if role == "super_admin" && companyID == 0 {
		// 通过请求获取可选的公司ID参数
		requestCompanyID, _ := c.GetInt("company_id", 0)
		companyID = requestCompanyID
	}

	// 获取分页和搜索参数
	page, _ := c.GetInt("page", 1)
	pageSize, _ := c.GetInt("page_size", 10)
	search := c.GetString("search", "")

	// 记录请求参数
	logs.Info("获取操作员列表 - 参数: companyID=%d, page=%d, page_size=%d, search=%s",
		companyID, page, pageSize, search)

	// 获取操作员列表
	operators, total, err := models.GetCompanyOperators(companyID, page, pageSize, search)
	if err != nil {
		logs.Error("获取操作员列表失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("获取操作员列表失败")
		c.ServeJSON()
		return
	}

	// 转换为用户信息列表，确保包含创建时间
	operatorInfos := make([]map[string]interface{}, 0, len(operators))
	for _, operator := range operators {
		userInfo := models.GetUserInfo(operator)

		// 确保创建时间字段存在并格式化
		if operator.CreatedAt.IsZero() {
			userInfo["created_at"] = time.Now().Format("2006-01-02 15:04:05") // 使用当前时间
		} else {
			userInfo["created_at"] = operator.CreatedAt.Format("2006-01-02 15:04:05")
		}

		operatorInfos = append(operatorInfos, userInfo)
	}

	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"operators": operatorInfos,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
	c.ServeJSON()
}

// UpdateOperatorStatus 更新操作员状态
// @router /api/admin/company/operator/status/:id [put]
func (c *CompanyAdminController) UpdateOperatorStatus() {
	companyIDInterface := c.Ctx.Input.GetData("company_id")
	roleInterface := c.Ctx.Input.GetData("role")

	if companyIDInterface == nil || roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户信息")
		c.ServeJSON()
		return
	}

	companyID, ok := companyIDInterface.(int)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("公司ID类型错误")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("角色类型错误")
		c.ServeJSON()
		return
	}

	// 公司管理员才能修改操作员状态
	if role != "company_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 获取操作员ID
	operatorID, err := c.GetInt(":id")
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的操作员ID")
		c.ServeJSON()
		return
	}

	// 获取请求数据
	var req UpdateOperatorStatusRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 验证状态值是否有效
	if req.Status != 0 && req.Status != 1 {
		c.Data["json"] = utils.ErrorResponse("无效的状态值")
		c.ServeJSON()
		return
	}

	// 验证操作员是否属于该公司
	operator, err := models.GetUserByID(operatorID)
	if err != nil || operator == nil || operator.CompanyId != companyID {
		c.Data["json"] = utils.ErrorResponse("操作员不存在或不属于您的公司")
		c.ServeJSON()
		return
	}

	// 更新操作员状态
	operator.Status = req.Status

	// 修复: 使用正确的参数调用 UpdateUser
	operatorIDStr := strconv.Itoa(operatorID)
	updateUser := &models.User{
		Status: req.Status,
	}

	_, err = models.UpdateUser(operatorIDStr, updateUser)
	if err != nil {
		logs.Error("更新操作员状态失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("更新操作员状态失败")
		c.ServeJSON()
		return
	}

	// 记录操作日志
	logs.Info("操作员状态已更新 [ID=%d, 用户名=%s, 状态=%d, 操作者=%v, 时间=%s]",
		operatorID, operator.Username, req.Status, c.Ctx.Input.GetData("username"), "2025-05-14 06:58:52")

	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"id":     operatorID,
		"status": req.Status,
	})
	c.ServeJSON()
}

// UpdateOperatorInfo 更新操作员信息
// @router /api/admin/company/operator/info/:id [put]
func (c *CompanyAdminController) UpdateOperatorInfo() {
	companyIDInterface := c.Ctx.Input.GetData("company_id")
	roleInterface := c.Ctx.Input.GetData("role")

	if companyIDInterface == nil || roleInterface == nil {
		c.Data["json"] = utils.ErrorResponse("无法获取用户信息")
		c.ServeJSON()
		return
	}

	companyID, ok := companyIDInterface.(int)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("公司ID类型错误")
		c.ServeJSON()
		return
	}

	role, ok := roleInterface.(string)
	if !ok {
		c.Data["json"] = utils.ErrorResponse("角色类型错误")
		c.ServeJSON()
		return
	}

	// 公司管理员才能修改操作员信息
	if role != "company_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 获取操作员ID
	operatorID, err := c.GetInt(":id")
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的操作员ID")
		c.ServeJSON()
		return
	}

	// 获取请求数据
	var req UpdateOperatorInfoRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 验证请求数据
	if req.RealName == "" {
		c.Data["json"] = utils.ErrorResponse("真实姓名不能为空")
		c.ServeJSON()
		return
	}

	// 验证操作员是否属于该公司
	operator, err := models.GetUserByID(operatorID)
	if err != nil || operator == nil || operator.CompanyId != companyID {
		c.Data["json"] = utils.ErrorResponse("操作员不存在或不属于您的公司")
		c.ServeJSON()
		return
	}

	// 更新操作员信息
	operatorIDStr := strconv.Itoa(operatorID)
	updateUser := &models.User{
		RealName: req.RealName,
		Email:    req.Email,
		Phone:    req.Phone,
	}

	// 修复: 使用正确的参数调用 UpdateUser
	updatedUser, err := models.UpdateUser(operatorIDStr, updateUser)
	if err != nil {
		logs.Error("更新操作员信息失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("更新操作员信息失败")
		c.ServeJSON()
		return
	}

	// 记录操作日志
	logs.Info("操作员信息已更新 [ID=%d, 用户名=%s, 操作者=%v, 时间=%s]",
		operatorID, operator.Username, c.Ctx.Input.GetData("username"), "2025-05-14 06:58:52")

	c.Data["json"] = utils.SuccessResponse(models.GetUserInfo(updatedUser))
	c.ServeJSON()
}
