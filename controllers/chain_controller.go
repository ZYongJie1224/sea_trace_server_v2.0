package controllers

import (
	"sea_trace_server_V2.0/services"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// ChainController 区块链控制器
type ChainController struct {
	web.Controller
}

// GetChainInfo 获取区块链信息
// @router /api/chain/sysinfo [get]
func (c *ChainController) GetChainInfo() {
	webaseService := services.NewWebaseService()

	chainInfo, err := webaseService.GetChainInfo()
	if err != nil {
		logs.Error("获取区块链信息失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("获取区块链信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(chainInfo.Data)
	c.ServeJSON()
}

// TraceInfo 货物溯源信息
// @router /api/chain/trace/:goodId [get]
func (c *ChainController) TraceInfo() {
	goodID := c.Ctx.Input.Param(":goodId")

	webaseService := services.NewWebaseService()
	traceInfo, err := webaseService.GetFullTrace(goodID)
	if err != nil {
		logs.Error("获取溯源信息失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("获取溯源信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(traceInfo)
	c.ServeJSON()
}
