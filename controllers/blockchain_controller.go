package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"sea_trace_server_V2.0/services"
	"sea_trace_server_V2.0/utils"
)

// BlockchainController 区块链控制器
type BlockchainController struct {
	web.Controller
}

// CreateBlockchainUser 创建区块链用户
// @router /api/blockchain/user/create [post]
func (c *BlockchainController) CreateBlockchainUser() {
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 获取请求参数
	username := c.GetString("username")
	userType, _ := c.GetInt("type", 2) // 默认使用外部用户类型
	returnPrivateKey, _ := c.GetBool("return_private_key", true)

	if username == "" {
		c.Data["json"] = utils.ErrorResponse("用户名不能为空")
		c.ServeJSON()
		return
	}

	// 创建区块链用户
	webaseService := services.NewWebaseService()
	blockchainUser, err := webaseService.CreateBlockchainUser(username, userType, returnPrivateKey)
	if err != nil {
		logs.Error("创建区块链用户失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("创建区块链用户失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 记录操作日志
	logs.Info("创建区块链用户成功 [username=%s, address=%s, type=%d, 操作者=%s, time=%s]",
		username, blockchainUser.Address, userType, c.Ctx.Input.GetData("username"), "2025-05-14 09:05:03")

	// 返回成功结果
	c.Data["json"] = utils.SuccessResponse(blockchainUser)
	c.ServeJSON()
}

// GetSuperAdminAddress 获取超级管理员区块链地址
// @router /api/blockchain/superadmin/address [get]
func (c *BlockchainController) GetSuperAdminAddress() {
	webaseService := services.NewWebaseService()
	address := webaseService.GetSuperAdminBlockchainAddress()

	c.Data["json"] = utils.SuccessResponse(map[string]string{
		"address": address,
	})
	c.ServeJSON()
}

// UpdateSuperAdminAddress 更新超级管理员区块链地址
// @router /api/blockchain/superadmin/address [put]
func (c *BlockchainController) UpdateSuperAdminAddress() {
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	address := c.GetString("address")
	if address == "" {
		c.Data["json"] = utils.ErrorResponse("区块链地址不能为空")
		c.ServeJSON()
		return
	}

	webaseService := services.NewWebaseService()
	err := webaseService.SetSuperAdminBlockchainAddress(address)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("更新超级管理员区块链地址失败: " + err.Error())
		c.ServeJSON()
		return
	}

	logs.Info("超级管理员区块链地址已更新 [address=%s, 操作者=%s, time=%s]",
		address, c.Ctx.Input.GetData("username"), "2025-05-14 09:05:03")

	c.Data["json"] = utils.SuccessResponse(map[string]string{
		"address": address,
	})
	c.ServeJSON()
}
