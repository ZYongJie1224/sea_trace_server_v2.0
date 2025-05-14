package routers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"sea_trace_server_V2.0/controllers"
	"sea_trace_server_V2.0/middleware"
)

func init() {
	// 跨域配置
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// 公司管理员路由
	web.Router("/api/admin/company/info", &controllers.CompanyAdminController{}, "get:CompanyInfo;put:UpdateCompanyInfo")

	web.Router("/api/admin/company/operator/create", &controllers.CompanyAdminController{}, "post:CreateOperator")
	web.Router("/api/admin/company/operator/delete/:id", &controllers.CompanyAdminController{}, "delete:DeleteOperator")
	web.Router("/api/admin/company/operators", &controllers.CompanyAdminController{}, "get:GetOperators")
	web.Router("/api/admin/company/operator/status/:id", &controllers.CompanyAdminController{}, "put:UpdateOperatorStatus")
	web.Router("/api/admin/company/operator/info/:id", &controllers.CompanyAdminController{}, "put:UpdateOperatorInfo")

	// Admin API
	web.Router("/api/admin/stats", &controllers.AdminController{}, "get:Stats")

	// Chain API
	web.Router("/api/chain/sysinfo", &controllers.ChainController{}, "get:GetChainInfo")
	web.Router("/api/chain/trace/:goodId", &controllers.ChainController{}, "get:TraceInfo")
	// 初始化路由 - 仅用于系统初始化
	web.Router("/api/init/admin", &controllers.InitController{}, "get:InitAdmin")

	// 认证接口 - 无需认证
	web.Router("/api/auth/login", &controllers.AuthController{}, "post:Login")

	// 公共溯源查询接口 - 无需认证
	web.Router("/api/chain/trace/:goodId", &controllers.ChainController{}, "get:TraceInfo")

	// 用户控制器实例
	authController := &controllers.AuthController{}
	chainController := &controllers.ChainController{}
	superAdminController := &controllers.SuperAdminController{}
	companyAdminController := &controllers.CompanyAdminController{}
	operatorController := &controllers.OperatorController{}

	// 需要JWT认证的路由
	web.Router("/api/auth/myinfo", authController, "get:MyInfo")
	web.InsertFilter("/api/auth/myinfo", web.BeforeRouter, middleware.JWTAuth)

	// 区块链信息路由 - 任何认证用户可访问
	web.Router("/api/chain/sysinfo", chainController, "get:GetChainInfo")
	web.InsertFilter("/api/chain/sysinfo", web.BeforeRouter, middleware.JWTAuth)
	web.Router("/api/chain/nodes", &controllers.ChainController{}, "get:GetNodeInfo")

	// 超级管理员路由
	web.Router("/api/su/company/list", superAdminController, "get:CompanyList")
	web.Router("/api/su/company/create", superAdminController, "post:CreateCompany")
	web.Router("/api/su/company/update/:id", superAdminController, "put:UpdateCompany")
	web.Router("/api/su/company/delete/:id", superAdminController, "delete:DeleteCompany")
	web.Router("/api/su/company/admin/create", superAdminController, "post:CreateCompanyAdmin")

	// 为所有超级管理员路由添加中间件
	web.InsertFilter("/api/su/*", web.BeforeRouter, middleware.JWTAuth)
	web.InsertFilter("/api/su/*", web.BeforeRouter, middleware.SuperAdminAuth)

	// 公司管理员路由
	web.Router("/api/admin/company/info", companyAdminController, "get:CompanyInfo")
	web.Router("/api/admin/company/info", companyAdminController, "put:UpdateCompanyInfo")
	web.Router("/api/admin/company/operator/create", companyAdminController, "post:CreateOperator")
	web.Router("/api/admin/company/operator/delete/:id", companyAdminController, "delete:DeleteOperator")
	// 为所有公司管理员路由添加中间件
	web.InsertFilter("/api/admin/*", web.BeforeRouter, middleware.JWTAuth)
	web.InsertFilter("/api/admin/*", web.BeforeRouter, middleware.CompanyAdminAuth)
	// 用户管理路由组
	web.Router("/api/admin/user/list", &controllers.UserManagementController{}, "get:ListUsers")
	web.Router("/api/admin/user/create", &controllers.UserManagementController{}, "post:CreateUser")
	web.Router("/api/admin/user/update/:id", &controllers.UserManagementController{}, "put:UpdateUser")
	web.Router("/api/admin/user/delete/:id", &controllers.UserManagementController{}, "delete:DeleteUser")

	// 操作员路由 - 根据公司类型进行操作
	web.Router("/api/operator/reggood", operatorController, "post:RegisterGood")    // 货主注册货物
	web.Router("/api/operator/shipgood", operatorController, "post:ShipGood")       // 船东运输登记
	web.Router("/api/operator/inspectgood", operatorController, "post:InspectGood") // 港口验货登记
	web.Router("/api/operator/delivergood", operatorController, "post:DeliverGood") // 经销商收货登记

	// 为所有操作员路由添加中间件
	web.InsertFilter("/api/operator/*", web.BeforeRouter, middleware.JWTAuth)
	web.InsertFilter("/api/operator/*", web.BeforeRouter, middleware.CompanyOperatorAuth)
}
