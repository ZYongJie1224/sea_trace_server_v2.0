package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// models/goods_request.go

// GoodsRegisterRequest 货物注册请求
type GoodsRegisterRequest struct {
	GoodName     string    `json:"good_name" binding:"required"`
	BatchNumber  string    `json:"batch_number"`
	Description  string    `json:"description"`
	Location     string    `json:"location" binding:"required"`
	BatchInfo    string    `json:"batch_info"`
	QualityLevel string    `json:"quality_level"`
	ExpiryDate   time.Time `json:"expiry_date" binding:"required"`
}

// UnmarshalJSON 自定义反序列化方法，用于处理多种日期格式
func (r *GoodsRegisterRequest) UnmarshalJSON(data []byte) error {
	// 创建一个匿名结构体，与 GoodsRegisterRequest 具有相同的字段，但 ExpiryDate 是字符串
	type Alias struct {
		GoodName     string `json:"good_name"`
		BatchNumber  string `json:"batch_number"`
		Description  string `json:"description"`
		Location     string `json:"location"`
		BatchInfo    string `json:"batch_info"`
		QualityLevel string `json:"quality_level"`
		ExpiryDate   string `json:"expiry_date"`
	}

	// 使用临时结构进行初始解析
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	// 将普通字段复制到目标结构体
	r.GoodName = alias.GoodName
	r.BatchNumber = alias.BatchNumber
	r.Description = alias.Description
	r.Location = alias.Location
	r.BatchInfo = alias.BatchInfo
	r.QualityLevel = alias.QualityLevel

	// 解析日期字段，支持多种格式
	if alias.ExpiryDate != "" {
		// 首先尝试 RFC3339 格式
		t, err := time.Parse(time.RFC3339, alias.ExpiryDate)
		if err != nil {
			// 再尝试简单日期格式
			t, err = time.Parse("2006-01-02", alias.ExpiryDate)
			if err != nil {
				return fmt.Errorf("无法解析日期 '%s': %v", alias.ExpiryDate, err)
			}
		}
		r.ExpiryDate = t
	}

	return nil
}

// GoodsShipRequest 货物运输请求
type GoodsShipRequest struct {
	GoodID         string    `json:"good_id" binding:"required"`
	StartLocation  string    `json:"start_location" binding:"required"`
	EndLocation    string    `json:"end_location" binding:"required"`
	TransportInfo  string    `json:"transport_info" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required"`
	TrackingNumber string    `json:"tracking_number"`
}

// GoodsInspectRequest 货物验货请求
type GoodsInspectRequest struct {
	GoodID         string `json:"good_id" binding:"required"`
	InspectionInfo string `json:"inspection_info" binding:"required"`
	QualityScore   int    `json:"quality_score" binding:"required,min=0,max=100"`
	PassStatus     bool   `json:"pass_status"`
	Location       string `json:"location" binding:"required"`
	Notes          string `json:"notes"`
}

// GoodsDeliverRequest 货物交付请求
type GoodsDeliverRequest struct {
	GoodID           string `json:"good_id" binding:"required"`
	DeliveryInfo     string `json:"delivery_info" binding:"required"`
	RecipientName    string `json:"recipient_name" binding:"required"`
	RecipientContact string `json:"recipient_contact" binding:"required"`
	Location         string `json:"location" binding:"required"`
	Notes            string `json:"notes"`
}

// GoodsTraceRequest 溯源查询请求
type GoodsTraceRequest struct {
	GoodID string `form:"good_id" binding:"required"`
}

// GoodsListRequest 货物列表请求
type GoodsListRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Status   int    `form:"status"`
	Search   string `form:"search"`
}

// GoodsBasicResponse 货物基本响应
type GoodsBasicResponse struct {
	ID               int         `json:"id"`
	GoodID           string      `json:"good_id"`
	GoodName         string      `json:"good_name"`
	BatchNumber      string      `json:"batch_number"`
	OwnerCompanyId   int         `json:"owner_company_id"`
	OwnerCompany     string      `json:"owner_company"`
	Description      string      `json:"description"`
	Status           GoodsStatus `json:"status"`
	StatusText       string      `json:"status_text"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	BlockchainTxHash string      `json:"blockchain_tx_hash"`
}

// GoodsListResponse 货物列表响应
type GoodsListResponse struct {
	Total int                  `json:"total"`
	List  []GoodsBasicResponse `json:"list"`
}
