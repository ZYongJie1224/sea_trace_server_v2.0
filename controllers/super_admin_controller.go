package controllers

import (
	"encoding/json"
	"strconv"
	"time"

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
// @Title 获取公司列表
// @Description 获取公司列表，支持分页、搜索和类型筛选
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认10"
// @Param search query string false "搜索关键词"
// @Param type query int false "公司类型：-1(全部)，0(生产商)，1(运输商)，2(验货商)，3(经销商)"
// @Success 200 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 500 {object} utils.Response
// @router /api/su/company/list [get]
func (c *SuperAdminController) CompanyList() {
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 获取查询参数
	page, _ := c.GetInt("page", 1)
	pageSize, _ := c.GetInt("page_size", 10)
	keyword := c.GetString("search", "")
	companyType, _ := c.GetInt("type", -1) // -1表示全部类型

	// 调用模型层获取数据
	companies, total, err := models.GetCompanyList(page, pageSize, keyword, companyType)
	if err != nil {
		logs.Error("获取公司列表失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("获取公司列表失败")
		c.ServeJSON()
		return
	}

	// 返回结果
	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"companies": companies,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
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
	AdminName   string `json:"admin_name"`
	Password    string `json:"password"`
	RealName    string `json:"real_name"`
	Email       string `json:"email"`
	Phone2      string `json:"phone2"`
}

// CreateCompany 创建公司并在区块链注册
// @router /api/su/company/create [post]
func (c *SuperAdminController) CreateCompany() {
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	var req CreateCompanyRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 检查公司名称是否已存在
	exists, err := models.CheckCompanyNameExists(req.CompanyName)
	if err != nil {
		logs.Error("检查公司名称失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("检查公司名称失败")
		c.ServeJSON()
		return
	}

	if exists {
		c.Data["json"] = utils.ErrorResponse("公司名称已存在")
		c.ServeJSON()
		return
	}

	// 1. 创建区块链用户 - 用公司名称作为区块链用户名
	webaseService := services.NewWebaseService()
	blockchainUser, err := webaseService.CreateBlockchainUser(
		req.CompanyName,
		0,    // 使用本地用户类型
		true, // 获取私钥
	)

	if err != nil {
		logs.Error("为公司创建区块链用户失败 [company=%s, error=%v, time=%s]",
			req.CompanyName, err, "2025-05-14 12:44:16")
		c.Data["json"] = utils.ErrorResponse("创建区块链用户失败: " + err.Error())
		c.ServeJSON()
		return
	}

	logs.Info("为公司创建区块链用户成功 [company=%s, address=%s, time=%s]",
		req.CompanyName, blockchainUser.Address, "2025-05-14 12:44:16")

	// 2. 创建公司记录 - 包含区块链地址信息
	company := &models.Company{
		CompanyName: req.CompanyName,
		CompanyType: models.CompanyType(req.CompanyType),
		Address:     blockchainUser.Address,
		Contact:     req.Contact,
		Phone:       req.Phone,
		// BlockchainAddress: blockchainUser.Address, // 存储区块链地址到公司记录
	}

	o := models.GetOrm()
	id, err := o.Insert(company)
	if err != nil {
		logs.Error("创建公司失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("创建公司失败: " + err.Error())
		c.ServeJSON()
		return
	}
	company.ID = int(id)

	// 3. 在区块链上注册公司
	txHash, err := webaseService.RegisterCompany(
		company.CompanyName,
		int(company.CompanyType),
		blockchainUser.Address, // 使用新创建的区块链地址
	)

	if err != nil {
		logs.Error("区块链注册公司失败 [company=%s, id=%d, address=%s, error=%v, time=%s]",
			company.CompanyName, company.ID, blockchainUser.Address, err, "2025-05-14 12:44:16")
	} else {
		logs.Info("公司已在区块链成功注册 [company=%s, id=%d, address=%s, txHash=%s, time=%s]",
			company.CompanyName, company.ID, blockchainUser.Address, txHash, "2025-05-14 12:44:16")

		// 将交易哈希保存到公司记录中
		company.BlockchainTxHash = txHash
		err = models.UpdateCompany(company)
		if err != nil {
			logs.Warning("更新公司区块链交易信息失败 [company=%s, id=%d, error=%v, time=%s]",
				company.CompanyName, company.ID, err, "2025-05-14 12:44:16")
		}
	}

	// 记录操作日志
	logs.Info("超级管理员创建公司成功 [公司名=%s, 公司ID=%d, 区块链地址=%s, 操作者=%s, 时间=%s]",
		req.CompanyName, company.ID, blockchainUser.Address,
		c.Ctx.Input.GetData("username"), "2025-05-14 12:44:16")

	// 返回成功信息
	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"company": company,
		"blockchain_info": map[string]interface{}{
			"address":     blockchainUser.Address,
			"public_key":  blockchainUser.PublicKey,
			"private_key": blockchainUser.PrivateKey, // 注意：生产环境中不应返回私钥
			"tx_hash":     txHash,
			"registered":  txHash != "",
		},
	})
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
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

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

	// 检查是否更改了公司名称，如果是，检查新名称是否已存在
	if req.CompanyName != company.CompanyName {
		exists, err := models.CheckCompanyNameExists(req.CompanyName)
		if err != nil {
			logs.Error("检查公司名称失败: %v", err)
			c.Data["json"] = utils.ErrorResponse("检查公司名称失败")
			c.ServeJSON()
			return
		}

		if exists {
			c.Data["json"] = utils.ErrorResponse("公司名称已存在")
			c.ServeJSON()
			return
		}
	}

	oldCompanyType := company.CompanyType
	oldCompanyName := company.CompanyName

	company.CompanyName = req.CompanyName
	company.CompanyType = models.CompanyType(req.CompanyType)
	company.Address = req.Address
	company.Contact = req.Contact
	company.Phone = req.Phone

	if err := models.UpdateCompany(company); err != nil {
		logs.Error("更新公司失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("更新公司失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// TODO: 如果公司类型或名称有更改，可能需要更新区块链信息
	if oldCompanyType != company.CompanyType || oldCompanyName != company.CompanyName {
		logs.Info("公司基本信息已更改，但区块链合约可能不支持更新 [company=%s, id=%d, time=%s]",
			company.CompanyName, company.ID, "2025-05-14 09:59:00")
	}

	logs.Info("超级管理员更新公司成功 [公司名=%s, 公司ID=%d, 操作者=%s, 时间=%s]",
		company.CompanyName, company.ID, c.Ctx.Input.GetData("username"), "2025-05-14 09:59:00")

	c.Data["json"] = utils.SuccessResponse(company)
	c.ServeJSON()
}

// DeleteCompany 删除公司
// @router /api/su/company/delete/:id [delete]
func (c *SuperAdminController) DeleteCompany() {
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的公司ID")
		c.ServeJSON()
		return
	}

	// 获取公司信息，用于日志记录
	company, err := models.GetCompanyByID(id)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("公司不存在")
		c.ServeJSON()
		return
	}

	// 检查是否有关联用户
	admins, _ := models.GetCompanyAdmins(id)
	operators, _, _ := models.GetCompanyOperators(id, 1, 10, "")

	if len(admins) > 0 || len(operators) > 0 {
		c.Data["json"] = utils.ErrorResponse("公司有关联用户，不能删除")
		c.ServeJSON()
		return
	}

	// 检查是否有关联货物
	goodsCount, _ := models.CountCompanyGoods(id)
	if goodsCount > 0 {
		c.Data["json"] = utils.ErrorResponse("公司有关联货物，不能删除")
		c.ServeJSON()
		return
	}

	if err := models.DeleteCompany(id); err != nil {
		logs.Error("删除公司失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("删除公司失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 删除成功后记录日志
	logs.Info("超级管理员删除公司成功 [公司名=%s, 公司ID=%d, 操作者=%s, 时间=%s]",
		company.CompanyName, company.ID, c.Ctx.Input.GetData("username"), "2025-05-14 09:59:00")

	c.Data["json"] = utils.SuccessResponse(nil)
	c.ServeJSON()
}

// CreateCompanyAdminRequest 创建公司管理员请求
type CreateCompanyAdminRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	RealName  string `json:"real_name"`
	CompanyID int    `json:"company_id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// CreateCompanyAdmin 创建公司管理员
// @router /api/su/company/admin/create [post]
func (c *SuperAdminController) CreateCompanyAdmin() {
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	var req CreateCompanyAdminRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的请求数据")
		c.ServeJSON()
		return
	}

	// 验证公司是否存在
	company, err := models.GetCompanyByID(req.CompanyID)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("公司不存在")
		c.ServeJSON()
		return
	}

	// // 1. 首先创建区块链用户
	// webaseService := services.NewWebaseService()
	// blockchainUser, err := webaseService.CreateBlockchainUser(
	// 	req.Username,
	// 	2,    // 使用外部用户类型
	// 	true, // 获取私钥
	// )

	// if err != nil {
	// 	logs.Error("为公司管理员创建区块链用户失败 [username=%s, company=%s, error=%v, time=%s]",
	// 		req.Username, company.CompanyName, err, "2025-05-14 09:59:00")
	// 	c.Data["json"] = utils.ErrorResponse("创建区块链用户失败: " + err.Error())
	// 	c.ServeJSON()
	// 	return
	// }

	// logs.Info("为公司管理员创建区块链用户成功 [username=%s, address=%s, company=%s, time=%s]",
	// 	req.Username, blockchainUser.Address, company.CompanyName, "2025-05-14 09:59:00")
	// logs.Info(req.Password)
	// 2. 创建数据库用户
	user, err := models.CreateUser(
		req.Username,
		req.Password,
		req.RealName,
		"company_admin",
		req.CompanyID,
		req.Email,
		req.Phone,
	)

	if err != nil {
		logs.Error("创建公司管理员失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("创建公司管理员失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// // 3. 更新用户的区块链地址信息
	// user.BlockchainAddr = blockchainUser.Address
	// user.SignUserID = blockchainUser.SignUserID
	// user.BlockchainType = blockchainUser.Type

	// _, err = models.UpdateUser(strconv.Itoa(user.ID), user)
	// if err != nil {
	// 	logs.Warning("更新管理员区块链信息失败 [username=%s, address=%s, error=%v, time=%s]",
	// 		user.Username, blockchainUser.Address, err, "2025-05-14 09:59:00")
	// 	// 不阻止创建流程，仅记录警告
	// }

	// // 4. 如果公司还没有注册到区块链，使用创建的区块链地址注册公司
	// if company.BlockchainTxHash == "" {
	// 	txHash, err := webaseService.RegisterCompany(
	// 		company.CompanyName,
	// 		int(company.CompanyType),
	// 		blockchainUser.Address, // 使用新创建的区块链地址
	// 	)

	// 	if err != nil {
	// 		logs.Error("区块链注册公司失败 [company=%s, id=%d, address=%s, error=%v, time=%s]",
	// 			company.CompanyName, company.ID, blockchainUser.Address, err, "2025-05-14 09:59:00")
	// 	} else {
	// 		logs.Info("公司已在区块链成功注册 [company=%s, id=%d, address=%s, txHash=%s, time=%s]",
	// 			company.CompanyName, company.ID, blockchainUser.Address, txHash, "2025-05-14 09:59:00")

	// 		// 将交易哈希保存到公司记录中
	// 		company.BlockchainTxHash = txHash
	// 		models.UpdateCompany(company)
	// 	}
	// }

	// 记录操作日志
	logs.Info("超级管理员创建公司管理员成功 [username=%s, company=%s, companyID=%d, 操作者=%s, 时间=%s]",
		user.Username, company.CompanyName, company.ID, c.Ctx.Input.GetData("username"), "2025-05-14 09:59:00")

	// 5. 构建响应，包含区块链信息
	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"user":    user,
		"company": company,
		// "blockchain_info": map[string]interface{}{
		// 	"address":     blockchainUser.Address,
		// 	"public_key":  blockchainUser.PublicKey,
		// 	"private_key": blockchainUser.PrivateKey, // 注意：实际环境不应直接返回私钥
		// 	"tx_hash":     company.BlockchainTxHash,
		// 	"registered":  company.BlockchainTxHash != "",
		// },
	})
	c.ServeJSON()
}

// GetCompanyAdmins 获取公司管理员列表
// @router /api/su/company/:id/admins [get]
func (c *SuperAdminController) GetCompanyAdmins() {
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的公司ID")
		c.ServeJSON()
		return
	}

	// 验证公司是否存在
	company, err := models.GetCompanyByID(id)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("公司不存在")
		c.ServeJSON()
		return
	}

	admins, err := models.GetCompanyAdmins(id)
	if err != nil {
		logs.Error("获取公司管理员列表失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("获取公司管理员列表失败: " + err.Error())
		c.ServeJSON()
		return
	}

	c.Data["json"] = utils.SuccessResponse(map[string]interface{}{
		"company": company,
		"admins":  admins,
		"total":   len(admins),
	})
	c.ServeJSON()
}

// DeleteCompanyAdmin 删除公司管理员
// @router /api/su/company/admin/delete/:id [delete]
func (c *SuperAdminController) DeleteCompanyAdmin() {
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("无效的用户ID")
		c.ServeJSON()
		return
	}

	// 获取用户信息，用于日志记录和权限检查
	user, err := models.GetUserByID(id)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("用户不存在")
		c.ServeJSON()
		return
	}

	// 检查是否为公司管理员
	if user.Role != "company_admin" {
		c.Data["json"] = utils.ErrorResponse("该用户不是公司管理员")
		c.ServeJSON()
		return
	}

	// 获取公司信息
	company, err := models.GetCompanyByID(user.CompanyId)
	if err != nil {
		c.Data["json"] = utils.ErrorResponse("公司不存在")
		c.ServeJSON()
		return
	}

	// 检查该公司是否还有其他管理员
	admins, _ := models.GetCompanyAdmins(user.CompanyId)
	if len(admins) <= 1 {
		c.Data["json"] = utils.ErrorResponse("公司至少需要一名管理员，无法删除唯一管理员")
		c.ServeJSON()
		return
	}

	if err := models.DeleteUser(strconv.Itoa(id)); err != nil {
		logs.Error("删除公司管理员失败: %v", err)
		c.Data["json"] = utils.ErrorResponse("删除公司管理员失败: " + err.Error())
		c.ServeJSON()
		return
	}

	// 记录操作日志
	logs.Info("超级管理员删除公司管理员成功 [userID=%d, username=%s, company=%s, companyID=%d, 操作者=%s, 时间=%s]",
		user.Id, user.Username, company.CompanyName, company.ID, c.Ctx.Input.GetData("username"), "2025-05-14 09:59:00")

	c.Data["json"] = utils.SuccessResponse(nil)
	c.ServeJSON()
}

// GetSystemStats 获取系统统计信息
// @router /api/su/stats [get]
func (c *SuperAdminController) GetSystemStats() {
	// 检查权限 - 仅超级管理员可操作
	roleInterface := c.Ctx.Input.GetData("role")
	if roleInterface == nil || roleInterface.(string) != "super_admin" {
		c.Data["json"] = utils.ForbiddenResponse()
		c.ServeJSON()
		return
	}

	// 获取各种统计数据
	companiesCount, _ := models.CountCompanies()
	usersCount, _ := models.CountUsers()
	goodsCount, _ := models.CountCompanyGoods(-1) // -1 表示统计所有公司的货物数量
	transactionsCount, _ := models.CountTransactions()

	// 获取区块链信息
	webaseService := services.NewWebaseService()
	chainInfo, err := webaseService.GetChainSystemInfo()
	if err != nil {
		logs.Error("获取区块链信息失败: %v", err)
		// 继续流程，区块链信息为空
	}

	stats := map[string]interface{}{
		"companies_count":    companiesCount,
		"users_count":        usersCount,
		"goods_count":        goodsCount,
		"transactions_count": transactionsCount,
		"blockchain_info":    chainInfo,
		"query_time":         time.Now().Format("2006-01-02 15:04:05"),
	}

	c.Data["json"] = utils.SuccessResponse(stats)
	c.ServeJSON()
}
