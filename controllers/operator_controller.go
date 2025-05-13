package controllers

import (
	"encoding/json"

	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/services"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// OperatorController 操作员控制器
type OperatorController struct {
	web.Controller
}

// RegisterGoodRequest 注册货物请求
type RegisterGoodRequest struct {
	GoodID      string `json:"good_id"`
	GoodName    string `json:"good_name"`
	Description string `json:"description"`
}

// RegisterGood 货主注册货物
// @router /api/operator/reggood [post]
func (c *OperatorController) RegisterGood() {
	companyID := c.Ctx.Input.GetData("company_id").(int)
	username := c.Ctx.Input.GetData("username").(string)

	// 验证是否为货主公司
	company, err := models.GetCompanyByID(companyID)
	if err != nil || company.CompanyType != models.Producer {
		c.Data["json"] = utils.ErrorResponse("只有生产商才能注册货物")
		c.ServeJSON()
		return
	}

	var req RegisterGoodRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 调用区块链服务
	webaseService := services.NewWebaseService()
	userAddress := "0x" + username // 简化处理，实际中需要获取正确的用户地址

	txHash, err := webaseService.RegisterGood(req.GoodID, req.GoodName, userAddress)
	if err != nil {
		logs.Error("区块链注册货物失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("区块链注册货物失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 保存货物记录到数据库
	_, err = models.SaveGood(req.GoodID, req.GoodName, companyID, req.Description)
	if err != nil {
		logs.Error("保存货物信息失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("保存货物信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(map[string]string{
		"tx_hash": txHash,
		"good_id": req.GoodID,
	})
	c.ServeJSON()
}

// ShipGoodRequest 运输货物请求
type ShipGoodRequest struct {
	GoodID        string `json:"good_id"`
	TransportInfo string `json:"transport_info"`
}

// ShipGood 船东运输货物
// @router /api/operator/shipgood [post]
func (c *OperatorController) ShipGood() {
	companyID := c.Ctx.Input.GetData("company_id").(int)
	username := c.Ctx.Input.GetData("username").(string)

	// 验证是否为运输公司
	company, err := models.GetCompanyByID(companyID)
	if err != nil || company.CompanyType != models.Shipper {
		c.Data["json"] = utils.ErrorResponse("只有运输商才能登记运输")
		c.ServeJSON()
		return
	}

	var req ShipGoodRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 验证货物是否存在
	_, err = models.GetGoodByID(req.GoodID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("货物不存在")
		c.ServeJSON()
		return
	}

	// 调用区块链服务
	webaseService := services.NewWebaseService()
	userAddress := "0x" + username // 简化处理，实际中需要获取正确的用户地址

	txHash, err := webaseService.ShipGood(req.GoodID, req.TransportInfo, userAddress)
	if err != nil {
		logs.Error("区块链运输登记失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("区块链运输登记失败: " + err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(map[string]string{
		"tx_hash": txHash,
	})
	c.ServeJSON()
}

// InspectGoodRequest 验货请求
type InspectGoodRequest struct {
	GoodID         string `json:"good_id"`
	InspectionInfo string `json:"inspection_info"`
}

// InspectGood 港口验货
// @router /api/operator/inspectgood [post]
func (c *OperatorController) InspectGood() {
	companyID := c.Ctx.Input.GetData("company_id").(int)
	username := c.Ctx.Input.GetData("username").(string)

	// 验证是否为港口
	company, err := models.GetCompanyByID(companyID)
	if err != nil || company.CompanyType != models.Port {
		c.Data["json"] = utils.ErrorResponse("只有验货商才能登记验货")
		c.ServeJSON()
		return
	}

	var req InspectGoodRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 验证货物是否存在
	_, err = models.GetGoodByID(req.GoodID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("货物不存在")
		c.ServeJSON()
		return
	}

	// 调用区块链服务
	webaseService := services.NewWebaseService()
	userAddress := "0x" + username

	txHash, err := webaseService.InspectGood(req.GoodID, req.InspectionInfo, userAddress)
	if err != nil {
		logs.Error("区块链验货登记失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("区块链验货登记失败: " + err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(map[string]string{
		"tx_hash": txHash,
	})
	c.ServeJSON()
}

// DeliverGoodRequest 交付货物请求
type DeliverGoodRequest struct {
	GoodID       string `json:"good_id"`
	DeliveryInfo string `json:"delivery_info"`
}

// DeliverGood 经销商收货
// @router /api/operator/delivergood [post]
func (c *OperatorController) DeliverGood() {
	companyID := c.Ctx.Input.GetData("company_id").(int)
	username := c.Ctx.Input.GetData("username").(string)

	// 验证是否为经销商
	company, err := models.GetCompanyByID(companyID)
	if err != nil || company.CompanyType != models.Dealer {
		c.Data["json"] = utils.ErrorResponse("只有经销商才能登记收货")
		c.ServeJSON()
		return
	}

	var req DeliverGoodRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 验证货物是否存在
	_, err = models.GetGoodByID(req.GoodID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("货物不存在")
		c.ServeJSON()
		return
	}

	// 调用区块链服务
	webaseService := services.NewWebaseService()
	userAddress := "0x" + username

	txHash, err := webaseService.DeliverGood(req.GoodID, req.DeliveryInfo, userAddress)
	if err != nil {
		logs.Error("区块链收货登记失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("区块链收货登记失败: " + err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(map[string]string{
		"tx_hash": txHash,
	})
	c.ServeJSON()
}
