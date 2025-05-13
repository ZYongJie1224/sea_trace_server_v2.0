package controllers

import (
	"encoding/json"
	"strconv"

	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/services"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// SuperAdminController 超级管理员控制器
type SuperAdminController struct {
	web.Controller
}

// CompanyList 获取公司列表
// @router /api/su/company/list [get]
func (c *SuperAdminController) CompanyList() {
	companies, err := models.GetCompanyList()
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取公司列表失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"companies": companies,
		"total":     len(companies),
	})
	c.ServeJSON()
}

// CreateCompanyRequest 创建公司请求
type CreateCompanyRequest struct {
	CompanyName string `json:"company_name"`
	CompanyType int    `json:"company_type"`
	Address     string `json:"address"`
	Contact     string `json:"contact"`
	Phone       string `json:"phone"`
}

// CreateCompany 创建公司
// @router /api/su/company/create [post]
func (c *SuperAdminController) CreateCompany() {
	var req CreateCompanyRequest

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	company, err := models.CreateCompany(
		req.CompanyName,
		models.CompanyType(req.CompanyType),
		req.Address,
		req.Contact,
		req.Phone,
	)

	if err != nil {
		logs.Error("创建公司失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("创建公司失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 同步到区块链
	webaseService := services.NewWebaseService()
	adminAddress := "0x0000000000000000000000000000000000000000" // 默认值，后续需替换为实际地址
	_, err = webaseService.RegisterCompany(company.CompanyName, int(company.CompanyType), adminAddress)
	if err != nil {
		logs.Error("区块链注册公司失败: %v", err)
		// 继续流程，不影响返回结果
	}

	c.Data["json"] = utils.SuccessResponse(company)
	c.ServeJSON()
}

// UpdateCompanyRequest 更新公司请求
type UpdateCompanyRequest struct {
	CompanyName string `json:"company_name"`
	CompanyType int    `json:"company_type"`
	Address     string `json:"address"`
	Contact     string `json:"contact"`
	Phone       string `json:"phone"`
}

// UpdateCompany 更新公司
// @router /api/su/company/update/:id [put]
func (c *SuperAdminController) UpdateCompany() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的公司ID")
		c.ServeJSON()
		return
	}

	company, err := models.GetCompanyByID(id)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("公司不存在")
		c.ServeJSON()
		return
	}

	var req UpdateCompanyRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	company.CompanyName = req.CompanyName
	company.CompanyType = models.CompanyType(req.CompanyType)
	company.Address = req.Address
	company.Contact = req.Contact
	company.Phone = req.Phone

	if err := models.UpdateCompany(company); err != nil {
		logs.Error("更新公司失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("更新公司失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(company)
	c.ServeJSON()
}

// DeleteCompany 删除公司
// @router /api/su/company/delete/:id [delete]
func (c *SuperAdminController) DeleteCompany() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的公司ID")
		c.ServeJSON()
		return
	}

	// 检查是否有关联用户
	admins, _ := models.GetCompanyAdmins(id)
	operators, _ := models.GetCompanyOperators(id)

	if len(admins) > 0 || len(operators) > 0 {
		c.Data["json"] = utils.ErrorResponse("公司有关联用户，不能删除")
		c.ServeJSON()
		return
	}

	if err := models.DeleteCompany(id); err != nil {
		logs.Error("删除公司失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("删除公司失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(nil)
	c.ServeJSON()
}

// CreateCompanyAdminRequest 创建公司管理员请求
type CreateCompanyAdminRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	RealName  string `json:"real_name"`
	CompanyID int    `json:"company_id"`
}

// CreateCompanyAdmin 创建公司管理员
// @router /api/su/company/admin/create [post]
func (c *SuperAdminController) CreateCompanyAdmin() {
	var req CreateCompanyAdminRequest

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 验证公司是否存在
	_, err := models.GetCompanyByID(req.CompanyID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("公司不存在")
		c.ServeJSON()
		return
	}

	user, err := models.CreateUser(
		req.Username,
		req.Password,
		req.RealName,
		"company_admin",
		req.CompanyID,
	)

	if err != nil {
		logs.Error("创建公司管理员失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("创建公司管理员失败: " + err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(user)
	c.ServeJSON()
}
