package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"sea_trace_server_V2.0/models"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/google/uuid"
)

// WebaseService 提供与WebaseFront交互的服务
type WebaseService struct {
	BaseURL         string // WebaseFront服务地址
	ContractABI     string // 合约ABI (文件路径或JSON字符串)
	ContractAddress string // 合约地址
	GroupID         int    // 区块链群组ID
	AppKey          string // 访问Webase的AppKey
	AppSecret       string // 访问Webase的AppSecret
	AppID           string // 应用ID，用于创建区块链用户
}

// NewWebaseService 创建WebaseService实例
func NewWebaseService() *WebaseService {
	baseURL, _ := web.AppConfig.String("webase_url")
	contractAddress, _ := web.AppConfig.String("contract_address")
	contractABI, _ := web.AppConfig.String("contract_abi")
	appKey, _ := web.AppConfig.String("webase_appkey")
	appSecret, _ := web.AppConfig.String("webase_appsecret")
	appID, _ := web.AppConfig.String("webase_app_id")
	groupID, _ := web.AppConfig.Int("webase_group_id")

	if groupID <= 0 {
		groupID = 1 // 默认使用Group 1
	}

	if appID == "" {
		appID = "sea_trace_app" // 默认应用ID
	}

	logs.Info("初始化WebaseService [url=%s, contractAddress=%s, groupID=%d, user=%s, time=%s]",
		baseURL, contractAddress, groupID, "ZYongJie1224", "2025-05-14 09:05:03")

	return &WebaseService{
		BaseURL:         baseURL,
		ContractABI:     contractABI,
		ContractAddress: contractAddress,
		GroupID:         groupID,
		AppKey:          appKey,
		AppSecret:       appSecret,
		AppID:           appID,
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
			logs.Error("读取合约ABI文件失败 [path=%s, error=%v]", w.ContractABI, err)
			return "", fmt.Errorf("读取合约ABI文件失败: %v", err)
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
	ContractName    string        `json:"contractName"`
	UseCns          bool          `json:"useCns"`
}

// TransactionResponse 交易响应结构
type TransactionResponse struct {
	Code            int                    `json:"code"`
	Message         string                 `json:"message"`
	Data            map[string]interface{} `json:"data"`
	TransactionHash string                 `json:"transactionHash"`
}

// ClientVersionResponse 客户端版本响应
type ClientVersionResponse struct {
	FiscoBcosVersion string `json:"FISCO-BCOS Version"`
	SupportedVersion string `json:"Supported Version"`
	ChainId          string `json:"Chain Id"`
	BuildTime        string `json:"Build Time"`
	BuildType        string `json:"Build Type"`
	GitBranch        string `json:"Git Branch"`
	GitCommitHash    string `json:"Git Commit Hash"`
}

// TransactionTotalResponse 交易总数响应
type TransactionTotalResponse struct {
	TxSum       int `json:"txSum"`
	BlockNumber int `json:"blockNumber"`
	FailedTxSum int `json:"failedTxSum"`
}

// ChainInfoResponse 区块链信息响应 - 旧版本保留结构
type ChainInfoResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// ChainSystemInfo 链系统信息 - 新版本结构
type ChainSystemInfo struct {
	// 客户端版本信息
	Version   string `json:"version"`
	ChainId   string `json:"chainId"`
	BuildTime string `json:"buildTime"`

	// 交易统计信息
	TxSum         int    `json:"txSum"`
	BlockNumber   int    `json:"blockNumber"`
	FailedTxSum   int    `json:"failedTxSum"`
	SuccessTxSum  int    `json:"successTxSum"`
	TxSuccessRate string `json:"txSuccessRate"`

	// 其他系统信息
	QueryTime string `json:"queryTime"`
	NodeCount int    `json:"nodeCount"`
}

// NodeInfo 节点信息结构
type NodeInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Active  bool   `json:"active"`
	P2PPort int    `json:"p2pPort"`
	Address string `json:"address"`
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

// BlockchainUserResponse WeBASE-Front创建用户响应
type BlockchainUserResponse struct {
	Address    string `json:"address"`    // 区块链地址
	PublicKey  string `json:"publicKey"`  // 公钥
	PrivateKey string `json:"privateKey"` // 私钥（根据请求参数可能为空）
	UserName   string `json:"userName"`   // 用户名（本地用户才有）
	Type       int    `json:"type"`       // 用户类型
	SignUserID string `json:"signUserId"` // WeBASE-Sign中的用户编号
	AppID      string `json:"appId"`      // 应用编号
}

// doGetRequest 执行GET请求
func (w *WebaseService) doGetRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error("创建GET请求失败 [url=%s, error=%v]", url, err)
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if w.AppKey != "" && w.AppSecret != "" {
		req.Header.Set("App-Key", w.AppKey)
		req.Header.Set("App-Secret", w.AppSecret)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("执行GET请求失败 [url=%s, error=%v]", url, err)
		return nil, fmt.Errorf("执行请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("读取响应内容失败 [url=%s, error=%v]", url, err)
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	logs.Debug("GET请求成功 [url=%s, responseSize=%d]", url, len(body))
	return body, nil
}

// doPostRequest 执行POST请求
func (w *WebaseService) doPostRequest(url string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logs.Error("序列化POST请求数据失败 [url=%s, error=%v]", url, err)
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.Error("创建POST请求失败 [url=%s, error=%v]", url, err)
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if w.AppKey != "" && w.AppSecret != "" {
		req.Header.Set("App-Key", w.AppKey)
		req.Header.Set("App-Secret", w.AppSecret)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("执行POST请求失败 [url=%s, error=%v]", url, err)
		return nil, fmt.Errorf("执行请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("读取响应内容失败 [url=%s, error=%v]", url, err)
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	logs.Debug("POST请求成功 [url=%s, responseSize=%d]", url, len(body))
	return body, nil
}

// sendTransaction 发送交易调用请求
func (w *WebaseService) sendTransaction(endpoint string, funcName string, funcParam []interface{}, userID string) (*TransactionResponse, error) {
	contractABI, err := w.readContractABI()
	if err != nil {
		return nil, err
	}

	var abiObj []interface{}
	err = json.Unmarshal([]byte(contractABI), &abiObj)
	if err != nil {
		logs.Error("解析合约ABI失败: %v", err)
		return nil, fmt.Errorf("解析合约ABI失败: %v", err)
	}

	requestBody := TransactionCallRequest{
		GroupID:         w.GroupID,
		ContractABI:     abiObj,
		ContractAddress: w.ContractAddress,
		FuncName:        funcName,
		FuncParam:       funcParam,
		User:            userID, // 当前登录用户
		// SignUserID:      userID,
		ContractName: "Traceability",
		UseCns:       false,
	}

	url := fmt.Sprintf("%s%s", w.BaseURL, endpoint)
	respData, err := w.doPostRequest(url, requestBody)
	if err != nil {
		return nil, err
	}

	var result TransactionResponse
	err = json.Unmarshal(respData, &result)
	if err != nil {
		logs.Error("解析交易响应失败: %v", err)
		return nil, fmt.Errorf("解析交易响应失败: %v", err)
	}

	if result.Code != 0 {
		logs.Error("交易调用失败 [function=%s, message=%s, code=%d]",
			funcName, result.Message, result.Code)
		return &result, fmt.Errorf("交易调用失败: %s", result.Message)
	}

	logs.Info("交易调用成功 [function=%s, time=%s, user=%s]",
		funcName, "2025-05-14 09:05:03", "ZYongJie1224")
	return &result, nil
}

// RegisterCompany 注册公司
func (w *WebaseService) RegisterCompany(name string, companyType int, adminAddress string) (string, error) {
	admin, _ := web.AppConfig.String("super_admin_blockchain_address")
	logs.Info("开始注册公司 [name=%s, type=%d, adminAddress=%s, user=%s, time=%s]",
		name, companyType, adminAddress, "ZYongJie1224", "2025-05-14 09:05:03")

	funcParam := []interface{}{name, companyType, adminAddress}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "registerCompany", funcParam, admin)
	if err != nil {
		return "", err
	}

	if result.TransactionHash != "" {
		logs.Info("公司注册成功 [name=%s, txHash=%s]", name, result.TransactionHash)
		return result.TransactionHash, nil
	}
	return "", errors.New("无法获取交易哈希")
}

// RegisterGood 注册货物
func (w *WebaseService) RegisterGood(goodID string, goodName string, userAddress string) (string, string, error) {
	logs.Info("开始注册货物 [goodID=%s, goodName=%s, userAddress=%s, user=%s, time=%s]",
		goodID, goodName, userAddress, "ZYongJie1224", "2025-05-14 09:05:03")

	funcParam := []interface{}{goodID, goodName}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "registerGood", funcParam, userAddress)
	if err != nil {
		return "", "", err
	}

	if result.TransactionHash != "" {
		logs.Info("货物注册成功 [goodID=%s, txHash=%s]", goodID, result.TransactionHash)
		return result.TransactionHash, result.Message, nil
	}
	return "", "", errors.New("无法获取交易哈希")
}

// ShipGood 运输货物
func (w *WebaseService) ShipGood(goodID string, transportInfo string, userAddress string) (string, string, error) {
	logs.Info("开始货物运输 [goodID=%s, transportInfo=%s, userAddress=%s, user=%s, time=%s]",
		goodID, transportInfo, userAddress, "ZYongJie1224", "2025-05-14 09:05:03")

	funcParam := []interface{}{goodID, transportInfo}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "shipGood", funcParam, userAddress)
	if err != nil {
		return "", result.Message, err
	}
	//TODO
	if result.TransactionHash != "" {
		models.UpdateGoodStatus(goodID, models.GoodsStatusShipped, result.TransactionHash)
		logs.Info("货物注册成功 [goodID=%s, txHash=%s]", goodID, result.TransactionHash)
		return result.TransactionHash, result.Message, nil
	}
	return "", "", errors.New("无法获取交易哈希")
}

// InspectGood 验货
func (w *WebaseService) InspectGood(goodID string, inspectionInfo string, userAddress string) (string, string, error) {
	logs.Info("开始货物验证 [goodID=%s, inspectionInfo=%s, userAddress=%s, user=%s, time=%s]",
		goodID, inspectionInfo, userAddress, "ZYongJie1224", "2025-05-14 09:05:03")

	funcParam := []interface{}{goodID, inspectionInfo}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "inspectGood", funcParam, userAddress)
	if err != nil {
		return "", result.Message, err
	}

	if result.TransactionHash != "" {
		models.UpdateGoodStatus(goodID, models.GoodsStatusInspected, result.TransactionHash)

		logs.Info("货物注册成功 [goodID=%s, txHash=%s]", goodID, result.TransactionHash)
		return result.TransactionHash, result.Message, nil
	}
	return "", "", errors.New("无法获取交易哈希")
}

// DeliverGood 经销商收货
func (w *WebaseService) DeliverGood(goodID string, deliveryInfo string, userAddress string) (string, string, error) {
	logs.Info("开始货物交付 [goodID=%s, deliveryInfo=%s, userAddress=%s, user=%s, time=%s]",
		goodID, deliveryInfo, userAddress, "ZYongJie1224", "2025-05-14 09:05:03")

	funcParam := []interface{}{goodID, deliveryInfo}
	result, err := w.sendTransaction("/WeBASE-Front/trans/handle", "deliverGood", funcParam, userAddress)
	if err != nil {
		return "", result.Message, err
	}

	if result.TransactionHash != "" {
		models.UpdateGoodStatus(goodID, models.GoodsStatusDelivered, result.TransactionHash)

		logs.Info("货物注册成功 [goodID=%s, txHash=%s]", goodID, result.TransactionHash)
		return result.TransactionHash, result.Message, nil
	}
	return "", "", errors.New("无法获取交易哈希")
}

// GetNodeList 获取节点列表
func (w *WebaseService) GetNodeList() ([]string, error) {
	url := fmt.Sprintf("%s/WeBASE-Front/%d/web3/groupPeers", w.BaseURL, w.GroupID)

	nodeData, err := w.doGetRequest(url)
	if err != nil {
		logs.Error("获取节点列表失败: %v", err)
		return nil, fmt.Errorf("获取节点列表失败: %v", err)
	}

	var nodes []string
	err = json.Unmarshal(nodeData, &nodes)
	if err != nil {
		logs.Error("解析节点列表信息失败: %v", err)
		return nil, fmt.Errorf("解析节点列表信息失败: %v", err)
	}

	logs.Info("成功获取节点列表 [数量=%d, user=%s, time=%s]",
		len(nodes), "ZYongJie1224", "2025-05-14 09:05:03")
	return nodes, nil
}

// GetNodeInfo 获取节点详细信息
func (w *WebaseService) GetNodeInfo() ([]NodeInfo, error) {
	// 先获取节点列表
	nodeAddresses, err := w.GetNodeList()
	if err != nil {
		return nil, err
	}

	// 创建节点信息数组
	nodes := make([]NodeInfo, 0, len(nodeAddresses))

	// 由于WeBASE可能没有提供获取节点详细信息的接口，我们使用地址构建基本信息
	for i, address := range nodeAddresses {
		shortAddr := address
		if len(shortAddr) > 20 {
			shortAddr = shortAddr[:20] + "..."
		}

		node := NodeInfo{
			ID:      address,
			Name:    fmt.Sprintf("节点%d", i+1),
			Status:  "正常",
			Active:  true,
			P2PPort: 30300 + i, // 假设端口，实际使用中需要从配置获取
			Address: shortAddr,
		}

		nodes = append(nodes, node)
	}

	logs.Info("成功构建节点详细信息 [数量=%d, user=%s, time=%s]",
		len(nodes), "ZYongJie1224", "2025-05-14 09:05:03")
	return nodes, nil
}

// GetChainInfo 获取区块链信息 - 旧接口方法保留，但内部重定向到新方法
func (w *WebaseService) GetChainInfo() (*ChainInfoResponse, error) {
	// 调用新方法获取系统信息
	sysInfo, err := w.GetChainSystemInfo()
	if err != nil {
		return nil, err
	}

	// 将新格式转换为旧格式以兼容现有代码
	result := &ChainInfoResponse{
		Code:    0,
		Message: "success",
		Data:    make(map[string]interface{}),
	}

	// 填充数据
	result.Data["version"] = sysInfo.Version
	result.Data["chainId"] = sysInfo.ChainId
	result.Data["buildTime"] = sysInfo.BuildTime
	result.Data["txSum"] = sysInfo.TxSum
	result.Data["blockNumber"] = sysInfo.BlockNumber
	result.Data["failedTxSum"] = sysInfo.FailedTxSum
	result.Data["successTxSum"] = sysInfo.SuccessTxSum
	result.Data["txSuccessRate"] = sysInfo.TxSuccessRate
	result.Data["queryTime"] = sysInfo.QueryTime
	result.Data["nodeCount"] = sysInfo.NodeCount

	logs.Info("获取区块链信息(兼容模式) [user=%s, time=%s]",
		"ZYongJie1224", "2025-05-14 09:05:03")
	return result, nil
}

// GetChainSystemInfo 获取区块链系统信息 - 新方法使用新接口
func (w *WebaseService) GetChainSystemInfo() (*ChainSystemInfo, error) {
	// 1. 获取客户端版本信息
	versionUrl := fmt.Sprintf("%s/WeBASE-Front/%d/web3/clientVersion", w.BaseURL, w.GroupID)
	clientVersionData, err := w.doGetRequest(versionUrl)
	if err != nil {
		logs.Error("获取客户端版本失败: %v", err)
		return nil, fmt.Errorf("获取客户端版本失败: %v", err)
	}

	var versionInfo ClientVersionResponse
	err = json.Unmarshal(clientVersionData, &versionInfo)
	if err != nil {
		logs.Error("解析客户端版本信息失败: %v", err)
		return nil, fmt.Errorf("解析客户端版本信息失败: %v", err)
	}

	// 2. 获取交易统计信息
	txUrl := fmt.Sprintf("%s/WeBASE-Front/%d/web3/transaction-total", w.BaseURL, w.GroupID)
	txData, err := w.doGetRequest(txUrl)
	if err != nil {
		logs.Error("获取交易统计失败: %v", err)
		return nil, fmt.Errorf("获取交易统计失败: %v", err)
	}

	var txInfo TransactionTotalResponse
	err = json.Unmarshal(txData, &txInfo)
	if err != nil {
		logs.Error("解析交易统计信息失败: %v", err)
		return nil, fmt.Errorf("解析交易统计信息失败: %v", err)
	}

	// 3. 获取节点数量
	nodeCount := 1 // 默认至少有一个节点
	nodeList, err := w.GetNodeList()
	if err != nil {
		logs.Warning("获取节点列表失败，使用默认节点数量: %v", err)
	} else {
		nodeCount = len(nodeList)
	}

	// 4. 计算成功交易数和成功率
	successTxSum := txInfo.TxSum - txInfo.FailedTxSum
	txSuccessRate := "100.00%"
	if txInfo.TxSum > 0 {
		successRate := float64(successTxSum) / float64(txInfo.TxSum) * 100
		txSuccessRate = fmt.Sprintf("%.2f%%", successRate)
	}

	// 5. 构建合并的响应
	result := &ChainSystemInfo{
		Version:       versionInfo.FiscoBcosVersion,
		ChainId:       versionInfo.ChainId,
		BuildTime:     versionInfo.BuildTime,
		TxSum:         txInfo.TxSum,
		BlockNumber:   txInfo.BlockNumber,
		FailedTxSum:   txInfo.FailedTxSum,
		SuccessTxSum:  successTxSum,
		TxSuccessRate: txSuccessRate,
		QueryTime:     time.Now().Format("2006-01-02 15:04:05"), // 当前时间 - 已更新
		NodeCount:     nodeCount,
	}

	logs.Info("成功获取区块链系统信息 [version=%s, blockNumber=%d, txSum=%d, user=%s, time=%s]",
		result.Version, result.BlockNumber, result.TxSum, "ZYongJie1224", "2025-05-14 09:05:03")
	return result, nil
}

// GetFullTrace 获取完整溯源信息
func (w *WebaseService) GetFullTrace(goodID string) (*TraceRecord, error) {
	logs.Info("开始获取货物溯源信息 [goodID=%s, user=%s, time=%s]",
		goodID, "ZYongJie1224", "2025-05-14 09:05:03")

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
		trace = w.enrichTraceRecord(trace)

		logs.Info("成功获取货物溯源信息 [goodID=%s, goodName=%s, stages=%d]",
			trace.GoodID, trace.GoodName, w.countCompletedStages(trace))
		return trace, nil
	}

	logs.Error("无法解析溯源信息 [goodID=%s]", goodID)
	return nil, errors.New("无法解析溯源信息")
}

// GetGoodStatus 获取货物状态
func (w *WebaseService) GetGoodStatus(goodID string) (int, error) {
	logs.Info("开始获取货物状态 [goodID=%s, user=%s, time=%s]",
		goodID, "ZYongJie1224", "2025-05-14 09:05:03")

	funcParam := []interface{}{goodID}
	result, err := w.sendTransaction("/WeBASE-Front/trans/call", "getGoodStatus", funcParam, "public_user")
	if err != nil {
		return -1, err
	}

	if statusStr, ok := result.Data["result"].(string); ok {
		status, err := strconv.Atoi(statusStr)
		if err != nil {
			logs.Error("解析货物状态失败 [goodID=%s, statusStr=%s, error=%v]",
				goodID, statusStr, err)
			return -1, err
		}

		logs.Info("成功获取货物状态 [goodID=%s, status=%d]", goodID, status)
		return status, nil
	}

	logs.Error("无法获取货物状态 [goodID=%s]", goodID)
	return -1, errors.New("无法获取货物状态")
}

// enrichTraceRecord 丰富溯源记录，添加更多信息
func (w *WebaseService) enrichTraceRecord(trace *TraceRecord) *TraceRecord {
	// 转换时间戳为可读时间
	registerTime, _ := strconv.ParseInt(trace.RegisterTime, 10, 64)
	trace.RegisterTime = time.Unix(registerTime, 0).Format("2006-01-02 15:04:05")

	// 查询并添加公司名称
	if ownerCompanyID, err := strconv.Atoi(trace.OwnerCompanyID); err == nil {
		if ownerCompany, err := models.GetCompanyByID(ownerCompanyID); err == nil {
			trace.GoodName = fmt.Sprintf("%s (生产商: %s)", trace.GoodName, ownerCompany.CompanyName)
		}
	}

	// 处理运输信息
	if trace.ShipExists {
		shipTime, _ := strconv.ParseInt(trace.ShipTime, 10, 64)
		trace.ShipTime = time.Unix(shipTime, 0).Format("2006-01-02 15:04:05")

		shipCompanyID, _ := strconv.Atoi(trace.ShipCompanyID)
		shipCompany, err := models.GetCompanyByID(shipCompanyID)
		if err == nil {
			trace.TransportInfo = fmt.Sprintf("运输商: %s, %s", shipCompany.CompanyName, trace.TransportInfo)
		}
	}

	// 处理验货信息
	if trace.InspectExists {
		inspectTime, _ := strconv.ParseInt(trace.InspectTime, 10, 64)
		trace.InspectTime = time.Unix(inspectTime, 0).Format("2006-01-02 15:04:05")

		portCompanyID, _ := strconv.Atoi(trace.PortCompanyID)
		portCompany, err := models.GetCompanyByID(portCompanyID)
		if err == nil {
			trace.InspectionInfo = fmt.Sprintf("验货商: %s, %s", portCompany.CompanyName, trace.InspectionInfo)
		}
	}

	// 处理交付信息
	if trace.DeliveryExists {
		deliveryTime, _ := strconv.ParseInt(trace.DeliveryTime, 10, 64)
		trace.DeliveryTime = time.Unix(deliveryTime, 0).Format("2006-01-02 15:04:05")

		dealerCompanyID, _ := strconv.Atoi(trace.DealerCompanyID)
		dealerCompany, err := models.GetCompanyByID(dealerCompanyID)
		if err == nil {
			trace.DeliveryInfo = fmt.Sprintf("经销商: %s, %s", dealerCompany.CompanyName, trace.DeliveryInfo)
		}
	}

	return trace
}

// countCompletedStages 计算货物已完成的阶段数
func (w *WebaseService) countCompletedStages(trace *TraceRecord) int {
	count := 1 // 注册阶段总是存在的

	if trace.ShipExists {
		count++
	}

	if trace.InspectExists {
		count++
	}

	if trace.DeliveryExists {
		count++
	}

	return count
}

// GetBlockNumber 获取当前区块高度
func (w *WebaseService) GetBlockNumber() (int64, error) {
	url := fmt.Sprintf("%s/WeBASE-Front/%d/web3/blockNumber", w.BaseURL, w.GroupID)

	blockData, err := w.doGetRequest(url)
	if err != nil {
		logs.Error("获取区块高度失败: %v", err)
		return 0, fmt.Errorf("获取区块高度失败: %v", err)
	}

	// 解析响应，可能是十六进制字符串
	blockNumberStr := string(blockData)
	if len(blockNumberStr) > 2 && blockNumberStr[:2] == "0x" {
		// 十六进制转十进制
		blockNumber, err := strconv.ParseInt(blockNumberStr[2:], 16, 64)
		if err != nil {
			logs.Error("解析区块高度失败: %v", err)
			return 0, fmt.Errorf("解析区块高度失败: %v", err)
		}

		logs.Info("成功获取区块高度 [blockNumber=%d, user=%s, time=%s]",
			blockNumber, "ZYongJie1224", "2025-05-14 09:05:03")
		return blockNumber, nil
	}

	// 尝试直接转换为整数
	blockNumber, err := strconv.ParseInt(blockNumberStr, 10, 64)
	if err != nil {
		logs.Error("解析区块高度失败: %v", err)
		return 0, fmt.Errorf("解析区块高度失败: %v", err)
	}

	logs.Info("成功获取区块高度 [blockNumber=%d, user=%s, time=%s]",
		blockNumber, "ZYongJie1224", "2025-05-14 09:05:03")
	return blockNumber, nil
}

// GetTransactionByHash 根据交易哈希获取交易信息
func (w *WebaseService) GetTransactionByHash(txHash string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/WeBASE-Front/%d/web3/transaction/%s", w.BaseURL, w.GroupID, txHash)

	txData, err := w.doGetRequest(url)
	if err != nil {
		logs.Error("获取交易信息失败 [txHash=%s, error=%v]", txHash, err)
		return nil, fmt.Errorf("获取交易信息失败: %v", err)
	}

	var txInfo map[string]interface{}
	err = json.Unmarshal(txData, &txInfo)
	if err != nil {
		logs.Error("解析交易信息失败 [txHash=%s, error=%v]", txHash, err)
		return nil, fmt.Errorf("解析交易信息失败: %v", err)
	}

	logs.Info("成功获取交易信息 [txHash=%s, user=%s, time=%s]",
		txHash, "ZYongJie1224", "2025-05-14 09:05:03")
	return txInfo, nil
}

// CreateBlockchainUser 创建区块链用户
// userType: 0-本地用户；1-本地随机；2-外部用户
// returnPrivateKey: 是否返回私钥，仅对外部用户有效
func (w *WebaseService) CreateBlockchainUser(username string, userType int, returnPrivateKey bool) (*BlockchainUserResponse, error) {
	logs.Info("开始创建区块链用户 [username=%s, type=%d, time=%s]",
		username, userType, "2025-05-14 09:05:03")

	// 构建请求URL
	baseURL := fmt.Sprintf("%s/WeBASE-Front/privateKey", w.BaseURL)
	params := url.Values{}
	params.Add("type", strconv.Itoa(userType))

	if userType == 0 || userType == 1 {
		// 本地用户需要传入用户名
		params.Add("userName", username)
	} else if userType == 2 {
		// 外部用户需要传入signUserId和appId
		signUserID := uuid.New().String() // 生成唯一的用户ID
		params.Add("signUserId", signUserID)
		params.Add("appId", "")

		// 是否返回私钥
		if returnPrivateKey {
			params.Add("returnPrivateKey", "true")
		}
	}

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	logs.Debug("调用WeBASE创建用户API [url=%s]", requestURL)

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(requestURL)
	if err != nil {
		logs.Error("调用WeBASE创建用户API失败: %v", err)
		return nil, fmt.Errorf("调用区块链服务失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result BlockchainUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logs.Error("解析WeBASE响应失败: %v", err)
		return nil, fmt.Errorf("解析区块链服务响应失败: %v", err)
	}

	logs.Info("成功创建区块链用户 [username=%s, address=%s, time=%s]",
		username, result.Address, "2025-05-14 09:05:03")

	return &result, nil
}

// ValidateBlockchainAddress 验证区块链地址格式
func (w *WebaseService) ValidateBlockchainAddress(address string) bool {
	// 简单的以太坊地址格式验证
	if len(address) != 42 || address[:2] != "0x" {
		return false
	}

	// 可以添加更复杂的验证逻辑
	return true
}

// GetSuperAdminBlockchainAddress 获取超级管理员区块链地址
func (w *WebaseService) GetSuperAdminBlockchainAddress() string {
	// 首先尝试从配置文件读取
	address, err := web.AppConfig.String("super_admin_blockchain_address")
	if err == nil && address != "" {
		return address
	}

	// 如果配置文件中没有，尝试从数据库读取
	config, err := models.GetSystemConfig("super_admin_blockchain_address")
	if err == nil && config != nil {
		return config.Value
	}

	// 默认地址
	logs.Warning("无法获取超级管理员区块链地址，使用默认值")
	return "0xd4a7eb6982f32c8dbcd49010c81b9dd947752467"
}

// SetSuperAdminBlockchainAddress 设置超级管理员区块链地址
func (w *WebaseService) SetSuperAdminBlockchainAddress(address string) error {
	// 验证地址格式
	if !w.ValidateBlockchainAddress(address) {
		return fmt.Errorf("无效的区块链地址格式")
	}

	// 保存到数据库
	err := models.SetSystemConfig("super_admin_blockchain_address", address, "超级管理员的区块链地址")
	if err != nil {
		logs.Error("保存超级管理员区块链地址失败: %v", err)
		return err
	}

	logs.Info("成功更新超级管理员区块链地址 [address=%s, time=%s]",
		address, "2025-05-14 09:05:03")
	return nil
}
