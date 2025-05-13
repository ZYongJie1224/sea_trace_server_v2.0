package middleware

import (
	"strings"

	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/server/web/context"
)

// JWTAuth JWT鉴权中间件
func JWTAuth(ctx *context.Context) {
	authHeader := ctx.Input.Header("Authorization")

	if authHeader == "" {
		ctx.Output.JSON(utils.UnauthorizedResponse(), false, false)
		ctx.ResponseWriter.WriteHeader(401)
		ctx.Abort(401, "未授权")
		return
	}

	// 支持 Bearer Token
	tokenString := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = authHeader[7:]
	}

	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		ctx.Output.JSON(utils.UnauthorizedResponse(), false, false)
		ctx.ResponseWriter.WriteHeader(401)
		ctx.Abort(401, "无效token")
		return
	}

	// 将token信息存储到上下文
	ctx.Input.SetData("user_id", claims.UserID)
	ctx.Input.SetData("username", claims.Username)
	ctx.Input.SetData("role", claims.Role)
	ctx.Input.SetData("company_id", claims.CompanyID)
}

// SuperAdminAuth 超级管理员权限中间件
func SuperAdminAuth(ctx *context.Context) {
	role := ctx.Input.GetData("role")

	if role != "super_admin" {
		ctx.Output.JSON(utils.ForbiddenResponse(), false, false)
		ctx.ResponseWriter.WriteHeader(403)
		ctx.Abort(403, "权限不足")
		return
	}
}

// CompanyAdminAuth 公司管理员权限中间件
func CompanyAdminAuth(ctx *context.Context) {
	role := ctx.Input.GetData("role")

	if role != "company_admin" && role != "super_admin" {
		ctx.Output.JSON(utils.ForbiddenResponse(), false, false)
		ctx.ResponseWriter.WriteHeader(403)
		ctx.Abort(403, "权限不足")
		return
	}
}

// CompanyOperatorAuth 公司操作员权限中间件
func CompanyOperatorAuth(ctx *context.Context) {
	role := ctx.Input.GetData("role")

	if role != "operator" && role != "company_admin" && role != "super_admin" {
		ctx.Output.JSON(utils.ForbiddenResponse(), false, false)
		ctx.ResponseWriter.WriteHeader(403)
		ctx.Abort(403, "权限不足")
		return
	}
}
