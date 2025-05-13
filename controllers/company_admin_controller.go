package controllers

import (
	"encoding/json"
	"strconv"

	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// CompanyAdminController 公司管理员控制器
type CompanyAdminController struct {
	web.Controller
}

// CompanyInfo 获取公司信息
// @router /api/admin/company/info [get]
func (c *CompanyAdminController) CompanyInfo() {
	companyID := c.Ctx.Input.GetData("company_id").(int)

	company, err := models.GetCompanyByID(companyID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取公司信息失败")
		c.ServeJSON()
		return
	}

	operators, _ := models.GetCompanyOperators(companyID)

	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"company":        company,
		"operators":      operators,
		"operator_count": len(operators),
	})
	c.ServeJSON()
}

// UpdateCompanyInfoRequest 更新公司信息请求
type UpdateCompanyInfoRequest struct {
	Address string `json:"address"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
}

// UpdateCompanyInfo 更新公司信息
// @router /api/admin/company/info [put]
func (c *CompanyAdminController) UpdateCompanyInfo() {
	companyID := c.Ctx.Input.GetData("company_id").(int)
	role := c.Ctx.Input.GetData("role").(string)

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

// CreateOperatorRequest 创建操作员请求
type CreateOperatorRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
}

// CreateOperator 创建操作员
// @router /api/admin/company/operator/create [post]
func (c *CompanyAdminController) CreateOperator() {
	companyID := c.Ctx.Input.GetData("company_id").(int)
	role := c.Ctx.Input.GetData("role").(string)

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

	user, err := models.CreateUser(
		req.Username,
		req.Password,
		req.RealName,
		"operator",
		companyID,
	)

	if err != nil {
		logs.Error("创建操作员失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("创建操作员失败: " + err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(user)
	c.ServeJSON()
}

// DeleteOperator 删除操作员
// @router /api/admin/company/operator/delete/:id [delete]
func (c *CompanyAdminController) DeleteOperator() {
	companyID := c.Ctx.Input.GetData("company_id").(int)
	role := c.Ctx.Input.GetData("role").(string)

	// 公司管理员才能删除操作员
	if role != "company_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的操作员ID")
		c.ServeJSON()
		return
	}

	// 确保操作员属于当前公司
	user, err := models.GetUserByID(id)
	if err != nil || user.CompanyID != companyID || user.Role != "operator" {
		c.Data["json"] = utils.ErrorResponse("操作员不存在或不属于当前公司")
		c.ServeJSON()
		return
	}

	if err := models.DeleteUserByID(id); err != nil {
		logs.Error("删除操作员失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("删除操作员失败")
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(nil)
	c.ServeJSON()
}
