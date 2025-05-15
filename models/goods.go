package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// GoodsStatus 货物状态枚举
type GoodsStatus int

const (
	GoodsStatusProduced  GoodsStatus = iota + 1 // 已生产
	GoodsStatusShipped                          // 已运输
	GoodsStatusInspected                        // 已验货
	GoodsStatusDelivered                        // 已交付
)

// GoodsStatusMap 货物状态映射
var GoodsStatusMap = map[GoodsStatus]string{
	GoodsStatusProduced:  "已生产",
	GoodsStatusShipped:   "已运输",
	GoodsStatusInspected: "已验货",
	GoodsStatusDelivered: "已交付",
}

// Goods 货物模型
type Goods struct {
	Id               int         `orm:"pk;auto" json:"id"`
	GoodId           string      `orm:"size(64);unique" json:"good_id"`
	GoodName         string      `orm:"size(100)" json:"good_name"`
	OwnerCompanyId   int         `orm:"column(owner_company_id)" json:"owner_company_id"`
	Description      string      `orm:"type(text);null" json:"description"`
	BatchNumber      string      `orm:"size(50);null" json:"batch_number"`
	Status           GoodsStatus `orm:"default(1)" json:"status"`
	CreatedAt        time.Time   `orm:"auto_now_add" json:"created_at"`
	UpdatedAt        time.Time   `orm:"auto_now" json:"updated_at"`
	BlockchainTxHash string      `orm:"size(66);null" json:"blockchain_tx_hash"`
}

// TableName 指定表名
func (g *Goods) TableName() string {
	return "goods"
}

// GoodsProduction 货物生产信息
type GoodsProduction struct {
	Id               int       `orm:"pk;auto" json:"id"`
	GoodsId          int       `orm:"index" json:"goods_id"`
	GoodId           string    `orm:"size(64);index" json:"good_id"`
	ProducedAt       time.Time `orm:"auto_now_add" json:"produced_at"`
	Location         string    `orm:"size(255)" json:"location"`
	BatchInfo        string    `orm:"type(text);null" json:"batch_info"`
	QualityLevel     string    `orm:"size(20);null" json:"quality_level"`
	ExpiryDate       time.Time `orm:"null" json:"expiry_date"`
	OperatorId       int       `orm:"default(0)" json:"operator_id"`
	OperatorName     string    `orm:"size(100);null" json:"operator_name"`
	BlockchainTxHash string    `orm:"size(66);null" json:"blockchain_tx_hash"`
	CreatedAt        time.Time `orm:"auto_now_add" json:"created_at"`
}

// TableName 指定表名
func (g *GoodsProduction) TableName() string {
	return "goods_production"
}

// GoodsTransport 货物运输信息
type GoodsTransport struct {
	Id                int       `orm:"pk;auto" json:"id"`
	GoodsId           int       `orm:"index" json:"goods_id"`
	GoodId            string    `orm:"size(64);index" json:"good_id"`
	TransporterId     int       `orm:"default(0)" json:"transporter_id"`
	TransporterName   string    `orm:"size(100);null" json:"transporter_name"`
	OperatorId        int       `orm:"default(0)" json:"operator_id"`
	OperatorName      string    `orm:"size(100);null" json:"operator_name"`
	StartLocation     string    `orm:"size(255)" json:"start_location"`
	EndLocation       string    `orm:"size(255)" json:"end_location"`
	TransportInfo     string    `orm:"type(text)" json:"transport_info"`
	StartTime         time.Time `orm:"auto_now_add" json:"start_time"`
	EndTime           time.Time `orm:"null" json:"end_time"`
	ActualArrivalTime time.Time `orm:"null" json:"actual_arrival_time"`
	TrackingNumber    string    `orm:"size(50);null" json:"tracking_number"`
	BlockchainTxHash  string    `orm:"size(66);null" json:"blockchain_tx_hash"`
	CreatedAt         time.Time `orm:"auto_now_add" json:"created_at"`
	UpdatedAt         time.Time `orm:"auto_now" json:"updated_at"`
}

// TableName 指定表名
func (g *GoodsTransport) TableName() string {
	return "goods_transport"
}

// GoodsInspection 货物验货信息
type GoodsInspection struct {
	Id               int       `orm:"pk;auto" json:"id"`
	GoodsId          int       `orm:"index" json:"goods_id"`
	GoodId           string    `orm:"size(64);index" json:"good_id"`
	InspectorId      int       `orm:"default(0)" json:"inspector_id"`
	InspectorName    string    `orm:"size(100);null" json:"inspector_name"`
	OperatorId       int       `orm:"default(0)" json:"operator_id"`
	OperatorName     string    `orm:"size(100);null" json:"operator_name"`
	InspectionInfo   string    `orm:"type(text)" json:"inspection_info"`
	QualityScore     int       `orm:"default(0)" json:"quality_score"`
	PassStatus       bool      `orm:"default(true)" json:"pass_status"`
	InspectionTime   time.Time `orm:"auto_now_add" json:"inspection_time"`
	Location         string    `orm:"size(255)" json:"location"`
	Notes            string    `orm:"type(text);null" json:"notes"`
	BlockchainTxHash string    `orm:"size(66);null" json:"blockchain_tx_hash"`
	CreatedAt        time.Time `orm:"auto_now_add" json:"created_at"`
}

// TableName 指定表名
func (g *GoodsInspection) TableName() string {
	return "goods_inspection"
}

// GoodsDelivery 货物交付信息
type GoodsDelivery struct {
	Id               int       `orm:"pk;auto" json:"id"`
	GoodsId          int       `orm:"index" json:"goods_id"`
	GoodId           string    `orm:"size(64);index" json:"good_id"`
	DealerId         int       `orm:"default(0)" json:"dealer_id"`
	DealerName       string    `orm:"size(100);null" json:"dealer_name"`
	OperatorId       int       `orm:"default(0)" json:"operator_id"`
	OperatorName     string    `orm:"size(100);null" json:"operator_name"`
	DeliveryInfo     string    `orm:"type(text)" json:"delivery_info"`
	RecipientName    string    `orm:"size(100)" json:"recipient_name"`
	RecipientContact string    `orm:"size(50)" json:"recipient_contact"`
	DeliveryTime     time.Time `orm:"auto_now_add" json:"delivery_time"`
	Location         string    `orm:"size(255)" json:"location"`
	Notes            string    `orm:"type(text);null" json:"notes"`
	BlockchainTxHash string    `orm:"size(66);null" json:"blockchain_tx_hash"`
	CreatedAt        time.Time `orm:"auto_now_add" json:"created_at"`
}

// TableName 指定表名
func (g *GoodsDelivery) TableName() string {
	return "goods_delivery"
}

// GetGoodByID 根据区块链ID获取货物
func GetGoodByID(goodId string) (*Goods, error) {
	o := GetOrm()
	good := &Goods{}
	err := o.QueryTable(good).Filter("good_id", goodId).One(good)
	if err != nil {
		logs.Error("获取货物信息失败 [goodID=%s, error=%v, time=%s]",
			goodId, err, "2025-05-15 02:50:46")
	}
	return good, err
}

// SaveGood 保存货物信息
func SaveGood(goodID, goodName string, ownerCompanyID int, description string, batchNumber string) (*Goods, error) {
	good := &Goods{
		GoodId:         goodID,
		GoodName:       goodName,
		OwnerCompanyId: ownerCompanyID,
		Description:    description,
		BatchNumber:    batchNumber,
		Status:         GoodsStatusProduced,
	}

	o := GetOrm()
	_, err := o.Insert(good)
	if err != nil {
		logs.Error("保存货物信息失败 [goodID=%s, goodName=%s, error=%v, time=%s]",
			goodID, goodName, err, "2025-05-15 02:50:46")
	} else {
		logs.Info("成功保存货物信息 [goodID=%s, goodName=%s, companyID=%d, time=%s]",
			goodID, goodName, ownerCompanyID, "2025-05-15 02:50:46")
	}
	return good, err
}

// UpdateGood 更新货物信息
func UpdateGood(good *Goods) error {
	o := GetOrm()
	good.UpdatedAt = time.Now()
	_, err := o.Update(good)
	if err != nil {
		logs.Error("更新货物信息失败 [goodID=%s, error=%v, time=%s]",
			good.GoodId, err, "2025-05-15 02:50:46")
	}
	return err
}

// UpdateGoodStatus 更新货物状态
func UpdateGoodStatus(goodID string, status GoodsStatus, blockchainTxHash string) error {
	o := GetOrm()
	good, err := GetGoodByID(goodID)
	if err != nil {
		return err
	}

	good.Status = status
	good.BlockchainTxHash = blockchainTxHash
	good.UpdatedAt = time.Now()

	_, err = o.Update(good, "Status", "BlockchainTxHash", "UpdatedAt")
	if err != nil {
		logs.Error("更新货物状态失败 [goodID=%s, status=%d, error=%v, time=%s]",
			goodID, status, err, "2025-05-15 02:50:46")
	} else {
		logs.Info("成功更新货物状态 [goodID=%s, status=%d, time=%s]",
			goodID, status, "2025-05-15 02:50:46")
	}
	return err
}

// SaveGoodsProduction 保存货物生产信息
func SaveGoodsProduction(goodsID int, goodID string, location, batchInfo, qualityLevel string,
	operatorID int, operatorName string, expiryDate time.Time) (*GoodsProduction, error) {

	production := &GoodsProduction{
		GoodsId:      goodsID,
		GoodId:       goodID,
		Location:     location,
		BatchInfo:    batchInfo,
		QualityLevel: qualityLevel,
		ExpiryDate:   expiryDate,
		OperatorId:   operatorID,
		OperatorName: operatorName,
	}

	o := GetOrm()
	_, err := o.Insert(production)
	if err != nil {
		logs.Error("保存货物生产信息失败 [goodID=%s, error=%v, time=%s]",
			goodID, err, "2025-05-15 02:50:46")
	} else {
		logs.Info("成功保存货物生产信息 [goodID=%s, time=%s]",
			goodID, "2025-05-15 02:50:46")
	}
	return production, err
}

// UpdateGoodsProduction 更新货物生产信息
func UpdateGoodsProduction(production *GoodsProduction) error {
	o := GetOrm()
	_, err := o.Update(production)
	if err != nil {
		logs.Error("更新货物生产信息失败 [goodID=%s, error=%v, time=%s]",
			production.GoodId, err, "2025-05-15 02:50:46")
	}
	return err
}

// SaveGoodsTransport 保存货物运输信息
func SaveGoodsTransport(goodsID int, goodID string, transporterID int, transporterName string,
	operatorID int, operatorName string, startLocation, endLocation, transportInfo string,
	endTime time.Time, trackingNumber string) (*GoodsTransport, error) {

	transport := &GoodsTransport{
		GoodsId:         goodsID,
		GoodId:          goodID,
		TransporterId:   transporterID,
		TransporterName: transporterName,
		OperatorId:      operatorID,
		OperatorName:    operatorName,
		StartLocation:   startLocation,
		EndLocation:     endLocation,
		TransportInfo:   transportInfo,
		EndTime:         endTime,
		TrackingNumber:  trackingNumber,
	}

	o := GetOrm()
	_, err := o.Insert(transport)
	if err != nil {
		logs.Error("保存货物运输信息失败 [goodID=%s, error=%v, time=%s]",
			goodID, err, "2025-05-15 02:50:46")
	} else {
		logs.Info("成功保存货物运输信息 [goodID=%s, time=%s]",
			goodID, "2025-05-15 02:50:46")
	}
	return transport, err
}

// UpdateGoodsTransport 更新货物运输信息
func UpdateGoodsTransport(transport *GoodsTransport) error {
	o := GetOrm()
	transport.UpdatedAt = time.Now()
	_, err := o.Update(transport)
	if err != nil {
		logs.Error("更新货物运输信息失败 [goodID=%s, error=%v, time=%s]",
			transport.GoodId, err, "2025-05-15 02:50:46")
	}
	return err
}

// SaveGoodsInspection 保存货物验货信息
func SaveGoodsInspection(goodsID int, goodID string, inspectorID int, inspectorName string,
	operatorID int, operatorName string, inspectionInfo string, qualityScore int,
	passStatus bool, location, notes string) (*GoodsInspection, error) {

	inspection := &GoodsInspection{
		GoodsId:        goodsID,
		GoodId:         goodID,
		InspectorId:    inspectorID,
		InspectorName:  inspectorName,
		OperatorId:     operatorID,
		OperatorName:   operatorName,
		InspectionInfo: inspectionInfo,
		QualityScore:   qualityScore,
		PassStatus:     passStatus,
		Location:       location,
		Notes:          notes,
	}

	o := GetOrm()
	_, err := o.Insert(inspection)
	if err != nil {
		logs.Error("保存货物验货信息失败 [goodID=%s, error=%v, time=%s]",
			goodID, err, "2025-05-15 02:50:46")
	} else {
		logs.Info("成功保存货物验货信息 [goodID=%s, time=%s]",
			goodID, "2025-05-15 02:50:46")
	}
	return inspection, err
}

// UpdateGoodsInspection 更新货物验货信息
func UpdateGoodsInspection(inspection *GoodsInspection) error {
	o := GetOrm()
	_, err := o.Update(inspection)
	if err != nil {
		logs.Error("更新货物验货信息失败 [goodID=%s, error=%v, time=%s]",
			inspection.GoodId, err, "2025-05-15 02:50:46")
	}
	return err
}

// SaveGoodsDelivery 保存货物交付信息
func SaveGoodsDelivery(goodsID int, goodID string, dealerID int, dealerName string,
	operatorID int, operatorName string, deliveryInfo, recipientName,
	recipientContact, location, notes string) (*GoodsDelivery, error) {

	delivery := &GoodsDelivery{
		GoodsId:          goodsID,
		GoodId:           goodID,
		DealerId:         dealerID,
		DealerName:       dealerName,
		OperatorId:       operatorID,
		OperatorName:     operatorName,
		DeliveryInfo:     deliveryInfo,
		RecipientName:    recipientName,
		RecipientContact: recipientContact,
		Location:         location,
		Notes:            notes,
	}

	o := GetOrm()
	_, err := o.Insert(delivery)
	if err != nil {
		logs.Error("保存货物交付信息失败 [goodID=%s, error=%v, time=%s]",
			goodID, err, "2025-05-15 02:50:46")
	} else {
		logs.Info("成功保存货物交付信息 [goodID=%s, time=%s]",
			goodID, "2025-05-15 02:50:46")
	}
	return delivery, err
}

// UpdateGoodsDelivery 更新货物交付信息
func UpdateGoodsDelivery(delivery *GoodsDelivery) error {
	o := GetOrm()
	_, err := o.Update(delivery)
	if err != nil {
		logs.Error("更新货物交付信息失败 [goodID=%s, error=%v, time=%s]",
			delivery.GoodId, err, "2025-05-15 02:50:46")
	}
	return err
}

// GetGoodsList 获取货物列表
func GetGoodsList(page, pageSize int, companyID int, search string, status int) ([]*Goods, int64, error) {
	o := GetOrm()
	query := o.QueryTable(new(Goods))

	// 按公司ID筛选
	if companyID > 0 {
		query = query.Filter("owner_company_id", companyID)
	}

	// 按状态筛选
	if status > 0 {
		query = query.Filter("status", status)
	}

	// 按关键词搜索
	if search != "" {
		cond := orm.NewCondition()
		searchCond := cond.Or("good_id__icontains", search).
			Or("good_name__icontains", search).
			Or("batch_number__icontains", search).
			Or("description__icontains", search)
		query = query.SetCond(searchCond)
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		logs.Error("统计货物数量失败: %v [time=%s]", err, "2025-05-15 02:50:46")
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	var goods []*Goods
	_, err = query.OrderBy("-id").Limit(pageSize, offset).All(&goods)
	if err != nil {
		logs.Error("获取货物列表失败: %v [time=%s]", err, "2025-05-15 02:50:46")
		return nil, 0, err
	}

	return goods, total, nil
}

// GetTraceInfo 获取完整溯源信息
func GetTraceInfo(goodID string) (map[string]interface{}, error) {
	o := GetOrm()

	// 获取基本信息
	good, err := GetGoodByID(goodID)
	if err != nil {
		logs.Error("获取货物基本信息失败 [goodID=%s, error=%v, time=%s]",
			goodID, err, "2025-05-15 02:50:46")
		return nil, err
	}

	// 获取公司信息
	company, err := GetCompanyByID(good.OwnerCompanyId)
	companyName := ""
	if err == nil {
		companyName = company.CompanyName
	}

	// 获取生产信息
	var production GoodsProduction
	err = o.QueryTable(new(GoodsProduction)).Filter("good_id", goodID).One(&production)
	hasProduction := err == nil

	// 获取运输信息
	var transport GoodsTransport
	err = o.QueryTable(new(GoodsTransport)).Filter("good_id", goodID).One(&transport)
	hasTransport := err == nil

	// 获取验货信息
	var inspection GoodsInspection
	err = o.QueryTable(new(GoodsInspection)).Filter("good_id", goodID).One(&inspection)
	hasInspection := err == nil

	// 获取交付信息
	var delivery GoodsDelivery
	err = o.QueryTable(new(GoodsDelivery)).Filter("good_id", goodID).One(&delivery)
	hasDelivery := err == nil

	// 构建完整溯源信息
	trace := map[string]interface{}{
		"basic": map[string]interface{}{
			"id":               good.Id,
			"good_id":          good.GoodId,
			"good_name":        good.GoodName,
			"batch_number":     good.BatchNumber,
			"description":      good.Description,
			"owner_company_id": good.OwnerCompanyId,
			"owner_company":    companyName,
			"created_at":       good.CreatedAt.Format("2006-01-02 15:04:05"),
			"status":           good.Status,
			"status_text":      GoodsStatusMap[good.Status],
			"blockchain_hash":  good.BlockchainTxHash,
		},
	}

	// 添加生产信息
	if hasProduction {
		trace["production"] = map[string]interface{}{
			"id":              production.Id,
			"location":        production.Location,
			"produced_at":     production.ProducedAt.Format("2006-01-02 15:04:05"),
			"batch_info":      production.BatchInfo,
			"quality_level":   production.QualityLevel,
			"expiry_date":     production.ExpiryDate.Format("2006-01-02"),
			"operator_name":   production.OperatorName,
			"blockchain_hash": production.BlockchainTxHash,
		}
	}

	// 添加运输信息
	if hasTransport {
		trace["transport"] = map[string]interface{}{
			"id":                  transport.Id,
			"transporter_id":      transport.TransporterId,
			"transporter_name":    transport.TransporterName,
			"start_location":      transport.StartLocation,
			"end_location":        transport.EndLocation,
			"transport_info":      transport.TransportInfo,
			"start_time":          transport.StartTime.Format("2006-01-02 15:04:05"),
			"end_time":            transport.EndTime.Format("2006-01-02 15:04:05"),
			"actual_arrival_time": transport.ActualArrivalTime.Format("2006-01-02 15:04:05"),
			"tracking_number":     transport.TrackingNumber,
			"operator_name":       transport.OperatorName,
			"blockchain_hash":     transport.BlockchainTxHash,
		}
	}

	// 添加验货信息
	if hasInspection {
		trace["inspection"] = map[string]interface{}{
			"id":              inspection.Id,
			"inspector_id":    inspection.InspectorId,
			"inspector_name":  inspection.InspectorName,
			"inspection_info": inspection.InspectionInfo,
			"quality_score":   inspection.QualityScore,
			"pass_status":     inspection.PassStatus,
			"inspection_time": inspection.InspectionTime.Format("2006-01-02 15:04:05"),
			"location":        inspection.Location,
			"notes":           inspection.Notes,
			"operator_name":   inspection.OperatorName,
			"blockchain_hash": inspection.BlockchainTxHash,
		}
	}

	// 添加交付信息
	if hasDelivery {
		trace["delivery"] = map[string]interface{}{
			"id":                delivery.Id,
			"dealer_id":         delivery.DealerId,
			"dealer_name":       delivery.DealerName,
			"delivery_info":     delivery.DeliveryInfo,
			"recipient_name":    delivery.RecipientName,
			"recipient_contact": delivery.RecipientContact,
			"delivery_time":     delivery.DeliveryTime.Format("2006-01-02 15:04:05"),
			"location":          delivery.Location,
			"notes":             delivery.Notes,
			"operator_name":     delivery.OperatorName,
			"blockchain_hash":   delivery.BlockchainTxHash,
		}
	}

	return trace, nil
}
