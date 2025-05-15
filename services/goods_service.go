package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/google/uuid"
	"sea_trace_server_V2.0/models"
)

// GoodsService 货物业务服务
type GoodsService struct {
	WebaseService *WebaseService
}

// NewGoodsService 创建货物服务实例
func NewGoodsService() *GoodsService {
	return &GoodsService{
		WebaseService: NewWebaseService(),
	}
}

// RegisterGood 注册货物
func (s *GoodsService) RegisterGood(req *models.GoodsRegisterRequest, companyID int, operatorID int, operatorName string, blockchainAddress string) (*models.GoodsBasicResponse, error) {
	// 1. 生成唯一货物ID
	goodID := s.generateGoodID(companyID)

	// 2. 获取公司信息
	company, err := models.GetCompanyByID(companyID)
	if err != nil {
		return nil, fmt.Errorf("获取公司信息失败: %v", err)
	}

	// 3. 保存货物基本信息
	good, err := models.SaveGood(goodID, req.GoodName, companyID, req.Description, req.BatchNumber)
	if err != nil {
		return nil, fmt.Errorf("保存货物基本信息失败: %v", err)
	}

	// 4. 保存货物生产信息
	production, err := models.SaveGoodsProduction(
		good.Id,
		goodID,
		req.Location,
		req.BatchInfo,
		req.QualityLevel,
		operatorID,
		operatorName,
		req.ExpiryDate,
	)
	if err != nil {
		return nil, fmt.Errorf("保存货物生产信息失败: %v", err)
	}

	// 5. 将货物信息上链
	txHash, message, err := s.WebaseService.RegisterGood(goodID, req.GoodName, blockchainAddress)
	if err != nil {
		return nil, fmt.Errorf("货物信息上链失败: %v", err)
	}
	if message != "Success" {
		return nil, fmt.Errorf("货物信息上链失败: %v", message)
	}

	// 6. 更新区块链交易哈希
	good.BlockchainTxHash = txHash
	err = models.UpdateGood(good)
	if err != nil {
		logs.Warning("更新货物区块链交易哈希失败: %v", err)
	}

	production.BlockchainTxHash = txHash
	err = models.UpdateGoodsProduction(production)
	if err != nil {
		logs.Warning("更新生产信息区块链交易哈希失败: %v", err)
	}

	// 7. 构建响应
	response := &models.GoodsBasicResponse{
		ID:               good.Id,
		GoodID:           good.GoodId,
		GoodName:         good.GoodName,
		BatchNumber:      good.BatchNumber,
		OwnerCompanyId:   good.OwnerCompanyId,
		OwnerCompany:     company.CompanyName,
		Description:      good.Description,
		Status:           good.Status,
		StatusText:       models.GoodsStatusMap[good.Status],
		CreatedAt:        good.CreatedAt,
		UpdatedAt:        good.UpdatedAt,
		BlockchainTxHash: txHash,
	}

	logs.Info("货物注册成功 [goodID=%s, goodName=%s, companyID=%d, txHash=%s, time=%s]",
		goodID, req.GoodName, companyID, txHash, "2025-05-15 02:50:46")

	return response, nil
}

// ShipGood 运输货物
func (s *GoodsService) ShipGood(req *models.GoodsShipRequest, transporterID int, operatorID int, operatorName string, blockchainAddress string) (*models.GoodsBasicResponse, error) {
	// 1. 获取货物信息
	good, err := models.GetGoodByID(req.GoodID)
	if err != nil {
		return nil, fmt.Errorf("获取货物信息失败: %v", err)
	}

	// 2. 检查货物状态
	if good.Status != models.GoodsStatusProduced {
		return nil, errors.New("当前货物状态不允许运输，只有已生产的货物可以运输")
	}

	// 3. 获取公司信息
	company, err := models.GetCompanyByID(transporterID)
	if err != nil {
		return nil, fmt.Errorf("获取运输商信息失败: %v", err)
	}

	if company.CompanyType != models.Shipper {
		return nil, errors.New("只有运输商公司才能执行运输操作")
	}

	// 4. 保存货物运输信息
	transport, err := models.SaveGoodsTransport(
		good.Id,
		req.GoodID,
		transporterID,
		company.CompanyName,
		operatorID,
		operatorName,
		req.StartLocation,
		req.EndLocation,
		req.TransportInfo,
		req.EndTime,
		req.TrackingNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("保存货物运输信息失败: %v", err)
	}

	// 5. 将货物运输信息上链
	txHash, message, err := s.WebaseService.ShipGood(req.GoodID, req.TransportInfo, blockchainAddress)
	if err != nil {
		return nil, fmt.Errorf("货物信息上链失败: %v", err)
	}
	if message != "Success" {
		return nil, fmt.Errorf("货物信息上链失败: %v", message)
	}

	// 6. 更新区块链交易哈希和货物状态
	transport.BlockchainTxHash = txHash
	err = models.UpdateGoodsTransport(transport)
	if err != nil {
		logs.Warning("更新运输信息区块链交易哈希失败: %v", err)
	}

	err = models.UpdateGoodStatus(req.GoodID, models.GoodsStatusShipped, txHash)
	if err != nil {
		return nil, fmt.Errorf("更新货物状态失败: %v", err)
	}

	// 获取最新货物状态
	good, _ = models.GetGoodByID(req.GoodID)

	// 7. 构建响应
	ownerCompany, _ := models.GetCompanyByID(good.OwnerCompanyId)
	ownerCompanyName := ""
	if ownerCompany != nil {
		ownerCompanyName = ownerCompany.CompanyName
	}

	response := &models.GoodsBasicResponse{
		ID:               good.Id,
		GoodID:           good.GoodId,
		GoodName:         good.GoodName,
		BatchNumber:      good.BatchNumber,
		OwnerCompanyId:   good.OwnerCompanyId,
		OwnerCompany:     ownerCompanyName,
		Description:      good.Description,
		Status:           good.Status,
		StatusText:       models.GoodsStatusMap[good.Status],
		CreatedAt:        good.CreatedAt,
		UpdatedAt:        good.UpdatedAt,
		BlockchainTxHash: txHash,
	}

	logs.Info("货物运输信息记录成功 [goodID=%s, transporter=%s, txHash=%s, time=%s]",
		req.GoodID, company.CompanyName, txHash, "2025-05-15 02:50:46")

	return response, nil
}

// InspectGood 验货
func (s *GoodsService) InspectGood(req *models.GoodsInspectRequest, inspectorID int, operatorID int, operatorName string, blockchainAddress string) (*models.GoodsBasicResponse, error) {
	// 1. 获取货物信息
	good, err := models.GetGoodByID(req.GoodID)
	if err != nil {
		return nil, fmt.Errorf("获取货物信息失败: %v", err)
	}

	// 2. 检查货物状态
	if good.Status != models.GoodsStatusShipped {
		return nil, errors.New("当前货物状态不允许验货，只有已运输的货物可以验货")
	}

	// 3. 获取公司信息
	company, err := models.GetCompanyByID(inspectorID)
	if err != nil {
		return nil, fmt.Errorf("获取验货商信息失败: %v", err)
	}

	if company.CompanyType != models.Port {
		return nil, errors.New("只有验货商公司才能执行验货操作")
	}

	// 4. 保存货物验货信息
	inspection, err := models.SaveGoodsInspection(
		good.Id,
		req.GoodID,
		inspectorID,
		company.CompanyName,
		operatorID,
		operatorName,
		req.InspectionInfo,
		req.QualityScore,
		req.PassStatus,
		req.Location,
		req.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("保存货物验货信息失败: %v", err)
	}

	// 5. 将货物验货信息上链
	txHash, message, err := s.WebaseService.InspectGood(req.GoodID, req.InspectionInfo, blockchainAddress)
	if err != nil {
		return nil, fmt.Errorf("货物验货信息上链失败: %v", err)
	}
	if message != "Success" {
		return nil, fmt.Errorf("货物信息上链失败: %v", message)
	}

	// 6. 更新区块链交易哈希和货物状态
	inspection.BlockchainTxHash = txHash
	err = models.UpdateGoodsInspection(inspection)
	if err != nil {
		logs.Warning("更新验货信息区块链交易哈希失败: %v", err)
	}

	err = models.UpdateGoodStatus(req.GoodID, models.GoodsStatusInspected, txHash)
	if err != nil {
		return nil, fmt.Errorf("更新货物状态失败: %v", err)
	}

	// 获取最新货物状态
	good, _ = models.GetGoodByID(req.GoodID)

	// 7. 构建响应
	ownerCompany, _ := models.GetCompanyByID(good.OwnerCompanyId)
	ownerCompanyName := ""
	if ownerCompany != nil {
		ownerCompanyName = ownerCompany.CompanyName
	}

	response := &models.GoodsBasicResponse{
		ID:               good.Id,
		GoodID:           good.GoodId,
		GoodName:         good.GoodName,
		BatchNumber:      good.BatchNumber,
		OwnerCompanyId:   good.OwnerCompanyId,
		OwnerCompany:     ownerCompanyName,
		Description:      good.Description,
		Status:           good.Status,
		StatusText:       models.GoodsStatusMap[good.Status],
		CreatedAt:        good.CreatedAt,
		UpdatedAt:        good.UpdatedAt,
		BlockchainTxHash: txHash,
	}

	logs.Info("货物验货信息记录成功 [goodID=%s, inspector=%s, passed=%v, txHash=%s, time=%s]",
		req.GoodID, company.CompanyName, req.PassStatus, txHash, "2025-05-15 02:50:46")

	return response, nil
}

// DeliverGood 交付货物
func (s *GoodsService) DeliverGood(req *models.GoodsDeliverRequest, dealerID int, operatorID int, operatorName string, blockchainAddress string) (*models.GoodsBasicResponse, error) {
	// 1. 获取货物信息
	good, err := models.GetGoodByID(req.GoodID)
	if err != nil {
		return nil, fmt.Errorf("获取货物信息失败: %v", err)
	}

	// 2. 检查货物状态
	if good.Status != models.GoodsStatusInspected {
		return nil, errors.New("当前货物状态不允许交付，只有已验货的货物可以交付")
	}

	// 3. 获取公司信息
	company, err := models.GetCompanyByID(dealerID)
	if err != nil {
		return nil, fmt.Errorf("获取经销商信息失败: %v", err)
	}

	if company.CompanyType != models.Dealer {
		return nil, errors.New("只有经销商公司才能执行交付操作")
	}

	// 4. 保存货物交付信息
	delivery, err := models.SaveGoodsDelivery(
		good.Id,
		req.GoodID,
		dealerID,
		company.CompanyName,
		operatorID,
		operatorName,
		req.DeliveryInfo,
		req.RecipientName,
		req.RecipientContact,
		req.Location,
		req.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("保存货物交付信息失败: %v", err)
	}

	// 5. 将货物交付信息上链
	txHash, message, err := s.WebaseService.DeliverGood(req.GoodID, req.DeliveryInfo, blockchainAddress)
	if err != nil {
		return nil, fmt.Errorf("货物交付信息上链失败: %v", err)
	}
	if message != "Success" {
		return nil, fmt.Errorf("货物信息上链失败: %v", message)
	}

	// 6. 更新区块链交易哈希和货物状态
	delivery.BlockchainTxHash = txHash
	err = models.UpdateGoodsDelivery(delivery)
	if err != nil {
		logs.Warning("更新交付信息区块链交易哈希失败: %v", err)
	}

	err = models.UpdateGoodStatus(req.GoodID, models.GoodsStatusDelivered, txHash)
	if err != nil {
		return nil, fmt.Errorf("更新货物状态失败: %v", err)
	}

	// 获取最新货物状态
	good, _ = models.GetGoodByID(req.GoodID)

	// 7. 构建响应
	ownerCompany, _ := models.GetCompanyByID(good.OwnerCompanyId)
	ownerCompanyName := ""
	if ownerCompany != nil {
		ownerCompanyName = ownerCompany.CompanyName
	}

	response := &models.GoodsBasicResponse{
		ID:               good.Id,
		GoodID:           good.GoodId,
		GoodName:         good.GoodName,
		BatchNumber:      good.BatchNumber,
		OwnerCompanyId:   good.OwnerCompanyId,
		OwnerCompany:     ownerCompanyName,
		Description:      good.Description,
		Status:           good.Status,
		StatusText:       models.GoodsStatusMap[good.Status],
		CreatedAt:        good.CreatedAt,
		UpdatedAt:        good.UpdatedAt,
		BlockchainTxHash: txHash,
	}

	logs.Info("货物交付信息记录成功 [goodID=%s, dealer=%s, recipient=%s, txHash=%s, time=%s]",
		req.GoodID, company.CompanyName, req.RecipientName, txHash, "2025-05-15 02:50:46")

	return response, nil
}

// GetGoodsTrace 获取货物溯源信息
func (s *GoodsService) GetGoodsTrace(goodID string) (map[string]interface{}, error) {
	// 1. 获取数据库溯源信息
	trace, err := models.GetTraceInfo(goodID)
	if err != nil {
		return nil, fmt.Errorf("获取货物溯源信息失败: %v", err)
	}

	// 2. 获取区块链溯源记录
	blockchainTrace, err := s.WebaseService.GetFullTrace(goodID)
	if err != nil {
		logs.Warning("获取区块链溯源记录失败: %v", err)
	} else {
		trace["blockchain"] = blockchainTrace
	}

	logs.Info("获取货物溯源信息成功 [goodID=%s, time=%s]", goodID, "2025-05-15 02:50:46")

	return trace, nil
}

// GetGoodsList 获取货物列表
func (s *GoodsService) GetGoodsList(page, pageSize, companyID int, search string, status int) (*models.GoodsListResponse, error) {
	// 1. 获取货物列表
	goods, total, err := models.GetGoodsList(page, pageSize, companyID, search, status)
	if err != nil {
		return nil, fmt.Errorf("获取货物列表失败: %v", err)
	}

	// 2. 转换为响应格式
	list := make([]models.GoodsBasicResponse, 0, len(goods))
	for _, good := range goods {
		// 获取公司信息
		company, err := models.GetCompanyByID(good.OwnerCompanyId)
		companyName := ""
		if err == nil && company != nil {
			companyName = company.CompanyName
		}

		list = append(list, models.GoodsBasicResponse{
			ID:               good.Id,
			GoodID:           good.GoodId,
			GoodName:         good.GoodName,
			BatchNumber:      good.BatchNumber,
			OwnerCompanyId:   good.OwnerCompanyId,
			OwnerCompany:     companyName,
			Description:      good.Description,
			Status:           good.Status,
			StatusText:       models.GoodsStatusMap[good.Status],
			CreatedAt:        good.CreatedAt,
			UpdatedAt:        good.UpdatedAt,
			BlockchainTxHash: good.BlockchainTxHash,
		})
	}

	// 3. 构建响应
	response := &models.GoodsListResponse{
		Total: int(total),
		List:  list,
	}

	logs.Info("获取货物列表成功 [page=%d, pageSize=%d, companyID=%d, total=%d, time=%s]",
		page, pageSize, companyID, total, "2025-05-15 02:50:46")

	return response, nil
}

// generateGoodID 生成唯一货物ID
func (s *GoodsService) generateGoodID(companyID int) string {
	// 格式: G-公司ID-日期-随机字符串
	date := time.Now().Format("20060102")
	uuidStr := uuid.New().String()[:8]
	return fmt.Sprintf("G%d%s%s", companyID, date, uuidStr)
}
