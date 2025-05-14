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

// GetChainInfo 获取链信息
// @router /api/chain/sysinfo [get]
func (c *ChainController) GetChainInfo() {
	// 创建WeBase服务实例
	webaseService := services.NewWebaseService()

	// 获取链系统信息
	chainInfo, err := webaseService.GetChainSystemInfo()
	if err != nil {
		logs.Error("获取区块链信息失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("获取区块链信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 增加系统当前时间和用户信息
	response := map[string]interface{}{
		"chain_info":   chainInfo,
		"system_time":  "2025-05-14 02:40:11", // 当前时间
		"current_user": "ZYongJie1224",        // 当前用户
	}

	c.Data["json"] = utils.SuccessResponse(response)
	c.ServeJSON()
}

// TraceInfo 获取溯源信息
// @router /api/chain/trace/:goodId [get]
func (c *ChainController) TraceInfo() {
	goodId := c.Ctx.Input.Param(":goodId")
	if goodId == "" {
		c.Data["json"] = utils.ErrorResponse("无效的货物ID")
		c.ServeJSON()
		return
	}

	// 创建WeBase服务实例
	webaseService := services.NewWebaseService()

	// 获取货物溯源信息
	trace, err := webaseService.GetFullTrace(goodId)
	if err != nil {
		logs.Error("获取溯源信息失败 [goodId=%s]: %v", goodId, err)
		c.Data["json"] = utils.ErrorResponse("获取溯源信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 获取货物状态
	statusCode, err := webaseService.GetGoodStatus(goodId)
	if err != nil {
		logs.Warning("获取货物状态失败 [goodId=%s]: %v", goodId, err)
		statusCode = -1 // 使用默认值
	}

	// 状态码转换为状态描述
	statusMap := map[int]string{
		0:  "已登记",
		1:  "运输中",
		2:  "已验货",
		3:  "已交付",
		-1: "未知状态",
	}

	status := statusMap[-1]
	if desc, ok := statusMap[statusCode]; ok {
		status = desc
	}

	// 构建溯源点
	tracePoints := []map[string]interface{}{}

	// 添加登记点
	tracePoints = append(tracePoints, map[string]interface{}{
		"time":      trace.RegisterTime,
		"location":  "生产工厂",
		"operation": "登记生产",
		"operator":  "生产商 ID:" + trace.OwnerCompanyID,
		"info":      "产品名称: " + trace.GoodName,
	})

	// 添加运输点
	if trace.ShipExists {
		tracePoints = append(tracePoints, map[string]interface{}{
			"time":      trace.ShipTime,
			"location":  "物流中心",
			"operation": "货物运输",
			"operator":  "运输商 ID:" + trace.ShipCompanyID,
			"info":      trace.TransportInfo,
		})
	}

	// 添加验货点
	if trace.InspectExists {
		tracePoints = append(tracePoints, map[string]interface{}{
			"time":      trace.InspectTime,
			"location":  "港口/验货点",
			"operation": "货物验收",
			"operator":  "验货商 ID:" + trace.PortCompanyID,
			"info":      trace.InspectionInfo,
		})
	}

	// 添加交付点
	if trace.DeliveryExists {
		tracePoints = append(tracePoints, map[string]interface{}{
			"time":      trace.DeliveryTime,
			"location":  "销售终端",
			"operation": "货物交付",
			"operator":  "经销商 ID:" + trace.DealerCompanyID,
			"info":      trace.DeliveryInfo,
		})
	}

	// 构建响应
	response := map[string]interface{}{
		"good_id":      trace.GoodID,
		"good_name":    trace.GoodName,
		"status":       status,
		"status_code":  statusCode,
		"trace_points": tracePoints,
		"query_time":   "2025-05-14 02:40:11", // 当前时间
		"current_user": "ZYongJie1224",        // 当前用户
	}

	c.Data["json"] = utils.SuccessResponse(response)
	c.ServeJSON()
}

// GetNodeInfo 获取节点信息
// @router /api/chain/nodes [get]
func (c *ChainController) GetNodeInfo() {
	// 创建WeBase服务实例
	webaseService := services.NewWebaseService()

	// 获取节点信息
	nodeList, err := webaseService.GetNodeInfo()
	if err != nil {
		logs.Error("获取节点信息失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("获取节点信息失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 构建节点信息响应
	nodes := make([]map[string]interface{}, 0, len(nodeList))
	for _, node := range nodeList {
		nodeInfo := map[string]interface{}{
			"node_id":     node.ID,
			"node_name":   node.Name,
			"status":      node.Status,
			"active":      node.Active,
			"p2p_port":    node.P2PPort,
			"address":     node.Address,
			"update_time": "2025-05-14 02:40:11", // 当前时间
		}
		nodes = append(nodes, nodeInfo)
	}

	// 返回节点信息
	response := map[string]interface{}{
		"nodes":        nodes,
		"total":        len(nodes),
		"system_time":  "2025-05-14 02:40:11", // 当前时间
		"current_user": "ZYongJie1224",        // 当前用户
	}

	c.Data["json"] = utils.SuccessResponse(response)
	c.ServeJSON()
}
