package controllers

import (
	"fmt"
	"time"

	"sea_trace_server_V2.0/models"
	"sea_trace_server_V2.0/services"
	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// AdminController 管理员控制器
type AdminController struct {
	web.Controller
}

// CompanyTypeDistribution 公司类型分布
type CompanyTypeDistribution struct {
	Producer  int64 `json:"producer"`
	Shipper   int64 `json:"shipper"`
	Inspector int64 `json:"inspector"`
	Dealer    int64 `json:"dealer"`
}

// GoodsWeeklyData 每周货物数据
type GoodsWeeklyData struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ActivityInfo 活动信息
type ActivityInfo struct {
	Time        string `json:"time"`
	Type        string `json:"type"`
	GoodID      string `json:"good_id"`
	Description string `json:"description"`
	Operator    string `json:"operator"`
}

// Stats 获取管理员仪表盘统计数据
// @router /api/admin/stats [get]
func (c *AdminController) Stats() {
	// 获取统计数据
	companyCount, userCount, goodCount := getBasicStats()

	// 获取公司类型分布
	distribution := getCompanyDistribution()

	// 获取一周内货物数据
	weeklyData := getWeeklyGoodsData()

	// 获取最近活动
	activities := getRecentActivities()

	// 获取区块链高度
	blockNumber := getBlockchainHeight()

	// 组装响应数据
	data := map[string]interface{}{
		"companyCount":            companyCount,
		"userCount":               userCount,
		"goodCount":               goodCount,
		"blockNumber":             blockNumber,
		"companyTypeDistribution": distribution,
		"goodsWeeklyData":         weeklyData,
		"recentActivities":        activities,
	}

	c.Data["json"] = utils.SuccessResponse(data)
	c.ServeJSON()
}

// getBasicStats 获取基本统计数据
func getBasicStats() (int64, int64, int64) {
	o := models.GetOrm()

	// 获取公司数量
	companyCount, err := o.QueryTable(new(models.Company)).Count()
	if err != nil {
		logs.Error("获取公司数量失败: %v", err)
		companyCount = 0
	}

	// 获取用户数量
	userCount, err := o.QueryTable(new(models.User)).Count()
	if err != nil {
		logs.Error("获取用户数量失败: %v", err)
		userCount = 0
	}

	// 获取货物数量
	goodCount, err := o.QueryTable(new(models.Goods)).Count()
	if err != nil {
		logs.Error("获取货物数量失败: %v", err)
		goodCount = 0
	}

	return companyCount, userCount, goodCount
}

// getCompanyDistribution 获取公司类型分布
func getCompanyDistribution() CompanyTypeDistribution {
	o := models.GetOrm()

	// 获取不同类型公司数量
	producerCount, _ := o.QueryTable(new(models.Company)).Filter("company_type", models.Producer).Count()
	shipperCount, _ := o.QueryTable(new(models.Company)).Filter("company_type", models.Shipper).Count()
	portCount, _ := o.QueryTable(new(models.Company)).Filter("company_type", models.Port).Count()
	dealerCount, _ := o.QueryTable(new(models.Company)).Filter("company_type", models.Dealer).Count()

	return CompanyTypeDistribution{
		Producer:  producerCount,
		Shipper:   shipperCount,
		Inspector: portCount,
		Dealer:    dealerCount,
	}
}

// getWeeklyGoodsData 获取一周内每天的货物数据
func getWeeklyGoodsData() []GoodsWeeklyData {
	o := models.GetOrm()
	result := make([]GoodsWeeklyData, 0, 7)

	// 获取最近7天的数据
	now := time.Now()

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		endTime := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.Local)

		// 获取当天创建的货物数量
		count, err := o.QueryTable(new(models.Goods)).
			Filter("created_at__gte", startTime).
			Filter("created_at__lte", endTime).
			Count()

		if err != nil {
			logs.Error("获取日期 %s 的货物数据失败: %v", date.Format("2006-01-02"), err)
			count = 0
		}

		// 格式化日期为 "M-DD" 格式
		dateStr := fmt.Sprintf("%d-%02d", int(date.Month()), date.Day())

		result = append(result, GoodsWeeklyData{
			Date:  dateStr,
			Count: count,
		})
	}

	return result
}

// getRecentActivities 获取最近活动
func getRecentActivities() []ActivityInfo {
	// 这里需要实现获取最近活动的逻辑
	// 由于现有代码中没有直接对应的活动表，我们可以使用模拟数据
	// 实际项目中，你应该从数据库中查询最近的活动

	activities := []ActivityInfo{
		{
			Time:        "2025-05-13 16:30:22",
			Type:        "收货",
			GoodID:      "PROD20250513001",
			Description: "厦门海鲜市场收货",
			Operator:    "张三",
		},
		{
			Time:        "2025-05-13 14:25:17",
			Type:        "验货",
			GoodID:      "PROD20250513001",
			Description: "厦门港口验收通过",
			Operator:    "李四",
		},
		{
			Time:        "2025-05-13 12:15:45",
			Type:        "运输",
			GoodID:      "PROD20250513001",
			Description: "福州到厦门运输启动",
			Operator:    "王五",
		},
		{
			Time:        "2025-05-13 09:45:12",
			Type:        "注册",
			GoodID:      "PROD20250513001",
			Description: "福建带鱼注册",
			Operator:    "赵六",
		},
		{
			Time:        "2025-05-12 15:20:33",
			Type:        "收货",
			GoodID:      "PROD20250512003",
			Description: "厦门海鲜市场收货",
			Operator:    "张三",
		},
	}

	return activities
}

// getBlockchainHeight 获取区块链高度
func getBlockchainHeight() int64 {
	webaseService := services.NewWebaseService()
	chainInfo, err := webaseService.GetChainInfo()
	if err != nil {
		logs.Error("获取区块链高度失败: %v", err)
		return 0
	}

	// 从 chainInfo.Data 中提取区块高度
	// 由于GetChainInfo返回的是interface{}类型，需要转换
	blockNumberStr, ok := chainInfo.Data["blockNumber"].(string)
	if !ok {
		logs.Error("解析区块高度失败")
		return 0
	}

	// 将字符串转换为int64
	blockNumber, err := utils.StringToInt64(blockNumberStr)
	if err != nil {
		logs.Error("区块高度转换失败: %v", err)
		return 0
	}

	return blockNumber
}
