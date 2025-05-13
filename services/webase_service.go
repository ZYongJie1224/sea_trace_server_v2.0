package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"sea_trace_server_V2.0/models"

	"github.com/beego/beego/v2/server/web"
)

// WebaseService 提供与WebaseFront交互的服务
type WebaseService struct {
	BaseURL         string
	ContractABI     string
	ContractAddress string
	GroupID         int
	AppKey          string
	AppSecret       string
}

// NewWebaseService 创建WebaseService实例
func NewWebaseService() *WebaseService {
	baseURL, _ := web.AppConfig.String("webase_url")
	contractAddress, _ := web.AppConfig.String("contract_address")
	contractABI, _ := web.AppConfig.String("contract_abi")
	appKey, _ := web.AppConfig.String("webase_appkey")
	appSecret, _ := web.AppConfig.String("webase_appsecret")

	return &WebaseService{
		BaseURL:         baseURL,
		ContractABI:     contractABI,
		ContractAddress: contractAddress,
		GroupID:         1, // 默认使用Group 1
		AppKey:          appKey,
		AppSecret:       appSecret,
	}
}

// 从文件读取合约ABI
func (w *WebaseService) readContractABI() (string, error) {
	if w.ContractABI == "" {
		return "", errors.New("contract ABI not set")
	}

	// 如果是文件路径，则读取文件
	if w.ContractABI[0] == '.' || w.ContractABI[0] == '/' {
		data, err := ioutil.ReadFile(w.ContractABI)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// 否则直接返回字符串
	return w.ContractABI, nil
}

// TransactionCallRequest 交易调用请求结构
type TransactionCallRequest struct {
	GroupID         int           `json:"groupId"`
	ContractABI     []interface{} `json:"contractAbi"`
	ContractAddress string        `json:"contractAddress"`
	FuncName        string        `json:"funcName"`
	FuncParam       []interface{} `json:"funcParam"`
	User            string        `json:"user"`
	SignUserID      string        `json:"signUserId"`
}

// TransactionResponse 交易响应结构
type TransactionResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// ChainInfoResponse 区块链信息响应
type ChainInfoResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// TraceRecord 溯源记录结构
type TraceRecord struct {
	// 货物信息
	GoodID         string `json:"good_id"`
	OwnerCompanyID string `json:"owner_company_id"`
	GoodName       string `json:"good_name"`
	RegisterTime   string `json:"register_time"`

	// 运输信息
	ShipCompanyID    string `json:"ship_company_id"`
	ShipOperatorAddr string `json:"ship_operator_addr"`
	TransportInfo    string `json:"transport_info"`
	ShipTime         string `json:"ship_time"`
	ShipExists       bool   `json:"ship_exists"`

	// 验货信息
	PortCompanyID       string `json:"port_company_id"`
	InspectOperatorAddr string `json:"inspect_operator_addr"`
	InspectionInfo      string `json:"inspection_info"`
	InspectTime         string `json:"inspect_time"`
	InspectExists       bool   `json:"inspect_exists"`

	// 交付信息
	DealerCompanyID      string `json:"dealer_company_id"`
	DeliveryOperatorAddr string `json:"delivery_operator_addr"`
	DeliveryInfo         string `json:"delivery_info"`
	DeliveryTime         string `json:"delivery_time"`
	DeliveryExists       bool   `json:"delivery_exists"`
}

// 发送交易调用请求
func (w *WebaseService) sendTransaction(endpoint string, funcName string, funcParam []interface{}, userID string) (*TransactionResponse, error) {
	contractABI, err := w.readContractABI()
	if err != nil {
		return nil, err
	}

	var abiObj []interface{}
	err = json.Unmarshal([]byte(contractABI), &abiObj)
	if err != nil {
		return nil, fmt.Errorf("解析合约ABI失败: %v", err)
	}

	requestBody := TransactionCallRequest{
		GroupID:         w.GroupID,
		ContractABI:     abiObj,
		ContractAddress: w.ContractAddress,
		FuncName:        funcName,
		FuncParam:       funcParam,
		User:            "ZYongJie1224", // 实际使用时应该是当前用户
		SignUserID:      userID,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", w.BaseURL, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("App-Key", w.AppKey)
	req.Header.Set("App-Secret", w.AppSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TransactionResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return &result, fmt.Errorf("交易调用失败: %s", result.Message)
	}

	return &result, nil
}

// RegisterCompany 注册公司
func (w *WebaseService) RegisterCompany(name string, companyType int, adminAddress string) (string, error) {
	funcParam := []interface{}{name, companyType, adminAddress}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "registerCompany", funcParam, "super_admin_user")
	if err != nil {
		return "", err
	}

	if txHash, ok := result.Data["transactionHash"].(string); ok {
		return txHash, nil
	}
	return "", errors.New("无法获取交易哈希")
}

// RegisterGood 注册货物
func (w *WebaseService) RegisterGood(goodID string, goodName string, userAddress string) (string, error) {
	funcParam := []interface{}{goodID, goodName}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "registerGood", funcParam, userAddress)
	if err != nil {
		return "", err
	}

	if txHash, ok := result.Data["transactionHash"].(string); ok {
		return txHash, nil
	}
	return "", errors.New("无法获取交易哈希")
}

// ShipGood 运输货物
func (w *WebaseService) ShipGood(goodID string, transportInfo string, userAddress string) (string, error) {
	funcParam := []interface{}{goodID, transportInfo}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "shipGood", funcParam, userAddress)
	if err != nil {
		return "", err
	}

	if txHash, ok := result.Data["transactionHash"].(string); ok {
		return txHash, nil
	}
	return "", errors.New("无法获取交易哈希")
}

// InspectGood 验货
func (w *WebaseService) InspectGood(goodID string, inspectionInfo string, userAddress string) (string, error) {
	funcParam := []interface{}{goodID, inspectionInfo}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "inspectGood", funcParam, userAddress)
	if err != nil {
		return "", err
	}

	if txHash, ok := result.Data["transactionHash"].(string); ok {
		return txHash, nil
	}
	return "", errors.New("无法获取交易哈希")
}

// DeliverGood 经销商收货
func (w *WebaseService) DeliverGood(goodID string, deliveryInfo string, userAddress string) (string, error) {
	funcParam := []interface{}{goodID, deliveryInfo}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "deliverGood", funcParam, userAddress)
	if err != nil {
		return "", err
	}

	if txHash, ok := result.Data["transactionHash"].(string); ok {
		return txHash, nil
	}
	return "", errors.New("无法获取交易哈希")
}

// GetChainInfo 获取区块链信息
func (w *WebaseService) GetChainInfo() (*ChainInfoResponse, error) {
	url := fmt.Sprintf("%s/WeBASE-Front/chain/general/%d", w.BaseURL, w.GroupID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("App-Key", w.AppKey)
	req.Header.Set("App-Secret", w.AppSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ChainInfoResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetFullTrace 获取完整溯源信息
func (w *WebaseService) GetFullTrace(goodID string) (*TraceRecord, error) {
	funcParam := []interface{}{goodID}
	result, err := w.sendTransaction("/WeBASE-Front/trans/call", "getFullTrace", funcParam, "public_user")
	if err != nil {
		return nil, err
	}

	// 解析返回的结果
	if traceResult, ok := result.Data["result"].(map[string]interface{}); ok {
		trace := &TraceRecord{
			GoodID:         traceResult["goodId"].(string),
			OwnerCompanyID: traceResult["ownerCompanyId"].(string),
			GoodName:       traceResult["goodName"].(string),
			RegisterTime:   traceResult["registerTime"].(string),

			ShipCompanyID:    traceResult["shipCompanyId"].(string),
			ShipOperatorAddr: traceResult["shipOperatorAddr"].(string),
			TransportInfo:    traceResult["transportInfo"].(string),
			ShipTime:         traceResult["shipTime"].(string),
			ShipExists:       traceResult["shipExists"].(bool),

			PortCompanyID:       traceResult["portCompanyId"].(string),
			InspectOperatorAddr: traceResult["inspectOperatorAddr"].(string),
			InspectionInfo:      traceResult["inspectionInfo"].(string),
			InspectTime:         traceResult["inspectTime"].(string),
			InspectExists:       traceResult["inspectExists"].(bool),

			DealerCompanyID:      traceResult["dealerCompanyId"].(string),
			DeliveryOperatorAddr: traceResult["deliveryOperatorAddr"].(string),
			DeliveryInfo:         traceResult["deliveryInfo"].(string),
			DeliveryTime:         traceResult["deliveryTime"].(string),
			DeliveryExists:       traceResult["deliveryExists"].(bool),
		}

		// 丰富溯源信息，添加公司名称等
		if trace.ShipExists {
			shipCompanyID, _ := strconv.Atoi(trace.ShipCompanyID)
			shipCompany, err := models.GetCompanyByID(shipCompanyID)
			if err == nil {
				trace.TransportInfo = fmt.Sprintf("运输商: %s, %s", shipCompany.CompanyName, trace.TransportInfo)
			}
		}

		if trace.InspectExists {
			portCompanyID, _ := strconv.Atoi(trace.PortCompanyID)
			portCompany, err := models.GetCompanyByID(portCompanyID)
			if err == nil {
				trace.InspectionInfo = fmt.Sprintf("验货商: %s, %s", portCompany.CompanyName, trace.InspectionInfo)
			}
		}

		if trace.DeliveryExists {
			dealerCompanyID, _ := strconv.Atoi(trace.DealerCompanyID)
			dealerCompany, err := models.GetCompanyByID(dealerCompanyID)
			if err == nil {
				trace.DeliveryInfo = fmt.Sprintf("经销商: %s, %s", dealerCompany.CompanyName, trace.DeliveryInfo)
			}
		}

		// 转换时间戳为可读时间
		registerTime, _ := strconv.ParseInt(trace.RegisterTime, 10, 64)
		trace.RegisterTime = time.Unix(registerTime, 0).Format("2006-01-02 15:04:05")

		if trace.ShipExists {
			shipTime, _ := strconv.ParseInt(trace.ShipTime, 10, 64)
			trace.ShipTime = time.Unix(shipTime, 0).Format("2006-01-02 15:04:05")
		}

		if trace.InspectExists {
			inspectTime, _ := strconv.ParseInt(trace.InspectTime, 10, 64)
			trace.InspectTime = time.Unix(inspectTime, 0).Format("2006-01-02 15:04:05")
		}

		if trace.DeliveryExists {
			deliveryTime, _ := strconv.ParseInt(trace.DeliveryTime, 10, 64)
			trace.DeliveryTime = time.Unix(deliveryTime, 0).Format("2006-01-02 15:04:05")
		}

		return trace, nil
	}

	return nil, errors.New("无法解析溯源信息")
}

// GetGoodStatus 获取货物状态
func (w *WebaseService) GetGoodStatus(goodID string) (int, error) {
	funcParam := []interface{}{goodID}
	result, err := w.sendTransaction("/WeBASE-Front/trans/call", "getGoodStatus", funcParam, "public_user")
	if err != nil {
		return -1, err
	}

	if statusStr, ok := result.Data["result"].(string); ok {
		status, err := strconv.Atoi(statusStr)
		if err != nil {
			return -1, err
		}
		return status, nil
	}
	return -1, errors.New("无法获取货物状态")
}
