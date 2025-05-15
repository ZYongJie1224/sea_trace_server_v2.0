package controllers

import (
	"encoding/json"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/services"
	"sea_trace_server_V2.0/utils"
)

// GoodsController 货物控制器
type GoodsController struct {
	web.Controller
	GoodsService *services.GoodsService
}

// NewGoodsController 创建货物控制器
func NewGoodsController() *GoodsController {
	return &GoodsController{
		GoodsService: services.NewGoodsService(),
	}
}

// RegisterGood 生产商注册货物
// @router /api/operator/goods/register [post]
func (c *GoodsController) RegisterGood() {
	// 1. 获取当前用户信息
	companyID := c.Ctx.Input.GetData("company_id").(int)
	username := c.Ctx.Input.GetData("username").(string)
	userID := c.Ctx.Input.GetData("user_id").(int)

	// 2. 验证是否为生产商
	company, err := models.GetCompanyByID(companyID)
	if err != nil || company.CompanyType != models.Producer {
		c.Data["json"] = utils.ErrorResponse("只有生产商才能注册货物")
		c.ServeJSON()
		return
	}

	// 3. 解析请求数据
	var req models.GoodsRegisterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logs.Error("解析请求数据失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 4. 验证请求数据
	if req.GoodName == "" {
		logs.Error("货物名称不能为空")
		c.Data["json"] = utils.ErrorResponse("货物名称不能为空")
		c.ServeJSON()
		return
	}

	if req.Location == "" {
		c.Data["json"] = utils.ErrorResponse("生产地点不能为空")
		c.ServeJSON()
		return
	}

	// 5. 获取用户详细信息
	user, err := models.GetUserByID(userID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取用户信息失败")
		c.ServeJSON()
		return
	}

	// 使用公司区块链地址
	blockchainAddress := company.Address

	// 确保有可用的区块链地址
	if blockchainAddress == "" {
		c.Data["json"] = utils.ErrorResponse("公司区块链地址未配置，请联系管理员")
		c.ServeJSON()
		return
	}

	logs.Info("注册货物使用公司区块链地址 [company=%s, address=%s, time=%s]",
		company.CompanyName, blockchainAddress, "2025-05-15 03:06:28")

	// 6. 调用服务层注册货物
	response, err := c.GoodsService.RegisterGood(&req, companyID, userID, user.RealName, blockchainAddress)
	if err != nil {
		logs.Error("注册货物失败: %v [user=%s, company=%s, time=%s]",
			err, username, company.CompanyName, "2025-05-15 03:06:28")
		c.Data["json"] = utils.ErrorResponse("注册货物失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 7. 返回成功响应
	c.Data["json"] = utils.SuccessResponse(response)
	c.ServeJSON()
}

// ShipGood 运输商运输货物
// @router /api/operator/goods/ship [post]
func (c *GoodsController) ShipGood() {
	// 1. 获取当前用户信息
	companyID := c.Ctx.Input.GetData("company_id").(int)
	username := c.Ctx.Input.GetData("username").(string)
	userID := c.Ctx.Input.GetData("user_id").(int)

	// 2. 验证是否为运输商
	company, err := models.GetCompanyByID(companyID)
	if err != nil || company.CompanyType != models.Shipper {
		c.Data["json"] = utils.ErrorResponse("只有运输商才能执行运输操作")
		c.ServeJSON()
		return
	}

	// 3. 解析请求数据
	var req models.GoodsShipRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 4. 验证请求数据
	if req.GoodID == "" {
		c.Data["json"] = utils.ErrorResponse("货物ID不能为空")
		c.ServeJSON()
		return
	}

	// 5. 获取用户详细信息
	user, err := models.GetUserByID(userID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取用户信息失败")
		c.ServeJSON()
		return
	}

	// 使用公司区块链地址
	blockchainAddress := company.Address

	// 确保有可用的区块链地址
	if blockchainAddress == "" {
		c.Data["json"] = utils.ErrorResponse("公司区块链地址未配置，请联系管理员")
		c.ServeJSON()
		return
	}

	logs.Info("运输货物使用公司区块链地址 [company=%s, address=%s, time=%s]",
		company.CompanyName, blockchainAddress, "2025-05-15 03:06:28")

	// 6. 调用服务层记录运输信息
	response, err := c.GoodsService.ShipGood(&req, companyID, userID, user.RealName, blockchainAddress)
	if err != nil {
		logs.Error("记录运输信息失败: %v [user=%s, company=%s, goodID=%s, time=%s]",
			err, username, company.CompanyName, req.GoodID, "2025-05-15 03:06:28")
		c.Data["json"] = utils.ErrorResponse("记录运输信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 7. 返回成功响应
	c.Data["json"] = utils.SuccessResponse(response)
	c.ServeJSON()
}

// InspectGood 验货商验货
// @router /api/operator/goods/inspect [post]
func (c *GoodsController) InspectGood() {
	// 1. 获取当前用户信息
	companyID := c.Ctx.Input.GetData("company_id").(int)
	username := c.Ctx.Input.GetData("username").(string)
	userID := c.Ctx.Input.GetData("user_id").(int)

	// 2. 验证是否为验货商
	company, err := models.GetCompanyByID(companyID)
	if err != nil || company.CompanyType != models.Port {
		c.Data["json"] = utils.ErrorResponse("只有验货商才能执行验货操作")
		c.ServeJSON()
		return
	}

	// 3. 解析请求数据
	var req models.GoodsInspectRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 4. 验证请求数据
	if req.GoodID == "" {
		c.Data["json"] = utils.ErrorResponse("货物ID不能为空")
		c.ServeJSON()
		return
	}

	// 5. 获取用户详细信息
	user, err := models.GetUserByID(userID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取用户信息失败")
		c.ServeJSON()
		return
	}

	// 使用公司区块链地址
	blockchainAddress := company.Address

	// 确保有可用的区块链地址
	if blockchainAddress == "" {
		c.Data["json"] = utils.ErrorResponse("公司区块链地址未配置，请联系管理员")
		c.ServeJSON()
		return
	}

	logs.Info("验货使用公司区块链地址 [company=%s, address=%s, time=%s]",
		company.CompanyName, blockchainAddress, "2025-05-15 03:06:28")

	// 6. 调用服务层记录验货信息
	response, err := c.GoodsService.InspectGood(&req, companyID, userID, user.RealName, blockchainAddress)
	if err != nil {
		logs.Error("记录验货信息失败: %v [user=%s, company=%s, goodID=%s, time=%s]",
			err, username, company.CompanyName, req.GoodID, "2025-05-15 03:06:28")
		c.Data["json"] = utils.ErrorResponse("记录验货信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 7. 返回成功响应
	c.Data["json"] = utils.SuccessResponse(response)
	c.ServeJSON()
}

// DeliverGood 经销商交付货物
// @router /api/operator/goods/deliver [post]
func (c *GoodsController) DeliverGood() {
	// 1. 获取当前用户信息
	companyID := c.Ctx.Input.GetData("company_id").(int)
	username := c.Ctx.Input.GetData("username").(string)
	userID := c.Ctx.Input.GetData("user_id").(int)

	// 2. 验证是否为经销商
	company, err := models.GetCompanyByID(companyID)
	if err != nil || company.CompanyType != models.Dealer {
		c.Data["json"] = utils.ErrorResponse("只有经销商才能执行交付操作")
		c.ServeJSON()
		return
	}

	// 3. 解析请求数据
	var req models.GoodsDeliverRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 4. 验证请求数据
	if req.GoodID == "" {
		c.Data["json"] = utils.ErrorResponse("货物ID不能为空")
		c.ServeJSON()
		return
	}

	// 5. 获取用户详细信息
	user, err := models.GetUserByID(userID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取用户信息失败")
		c.ServeJSON()
		return
	}

	// 使用公司区块链地址
	blockchainAddress := company.Address

	// 确保有可用的区块链地址
	if blockchainAddress == "" {
		c.Data["json"] = utils.ErrorResponse("公司区块链地址未配置，请联系管理员")
		c.ServeJSON()
		return
	}

	logs.Info("交付货物使用公司区块链地址 [company=%s, address=%s, time=%s]",
		company.CompanyName, blockchainAddress, "2025-05-15 03:06:28")

	// 6. 调用服务层记录交付信息
	response, err := c.GoodsService.DeliverGood(&req, companyID, userID, user.RealName, blockchainAddress)
	if err != nil {
		logs.Error("记录交付信息失败: %v [user=%s, company=%s, goodID=%s, time=%s]",
			err, username, company.CompanyName, req.GoodID, "2025-05-15 03:06:28")
		c.Data["json"] = utils.ErrorResponse("记录交付信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 7. 返回成功响应
	c.Data["json"] = utils.SuccessResponse(response)
	c.ServeJSON()
}

// GetGoodsList 获取货物列表
// @router /api/operator/goods/list [get]
func (c *GoodsController) GetGoodsList() {
	// 1. 获取当前用户信息
	companyID := c.Ctx.Input.GetData("company_id").(int)

	// 2. 获取查询参数
	page, _ := c.GetInt("page", 1)
	pageSize, _ := c.GetInt("page_size", 10)
	search := c.GetString("search", "")
	status, _ := c.GetInt("status", 0)

	// 3. 调用服务层获取货物列表
	response, err := c.GoodsService.GetGoodsList(page, pageSize, companyID, search, status)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取货物列表失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 4. 返回成功响应
	c.Data["json"] = utils.SuccessResponse(response)
	c.ServeJSON()
}

// GetGoodsTrace 获取货物溯源信息
// @router /api/operator/goods/trace [get]
func (c *GoodsController) GetGoodsTrace() {
	// 1. 获取货物ID
	goodID := c.GetString("good_id")
	if goodID == "" {
		c.Data["json"] = utils.ErrorResponse("货物ID不能为空")
		c.ServeJSON()
		return
	}

	// 2. 调用服务层获取溯源信息
	trace, err := c.GoodsService.GetGoodsTrace(goodID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取溯源信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 3. 返回成功响应
	c.Data["json"] = utils.SuccessResponse(trace)
	c.ServeJSON()
}

// PublicTrace 公开溯源查询接口
// @router /api/public/trace [get]
func (c *GoodsController) PublicTrace() {
	// 1. 获取货物ID
	goodID := c.GetString("good_id")
	if goodID == "" {
		c.Data["json"] = utils.ErrorResponse("货物ID不能为空")
		c.ServeJSON()
		return
	}

	// 2. 调用服务层获取溯源信息
	trace, err := c.GoodsService.GetGoodsTrace(goodID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("获取溯源信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 3. 记录公开溯源查询
	clientIP := c.Ctx.Input.IP()
	logs.Info("公开溯源查询请求 [goodID=%s, IP=%s, time=%s]",
		goodID, clientIP, "2025-05-15 03:06:28")

	// 4. 返回成功响应
	c.Data["json"] = utils.SuccessResponse(trace)
	c.ServeJSON()
}
