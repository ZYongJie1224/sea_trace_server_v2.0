pragma solidity ^0.6.10;
pragma experimental ABIEncoderV2;

/**
 * 溯源流程合约 v1.0
 * 开发日期: 2025-05-13
 * 开发者: ZYongJie1224
 * 
 * 流程说明：货物由生产商创建，运输商运输，港口验货，最终到达经销商
 */
contract Traceability {
    // 公司类型枚举
    enum CompanyType { Producer, Shipper, Port, Dealer }

    // 公司结构
    struct Company {
        uint256 id;
        string name;
        CompanyType companyType;
        address admin;
        bool exists;
    }

    // 货物结构
    struct Good {
        string goodId;
        uint256 ownerCompanyId;
        string goodName;
        uint256 registerTime;
        bool exists;
    }

    // 运输记录
    struct ShippingRecord {
        uint256 shipCompanyId;
        address operatorAddr;
        string transportInfo;
        uint256 time;
        bool exists;
    }

    // 验货记录
    struct InspectionRecord {
        uint256 portCompanyId;
        address operatorAddr;
        string inspectionInfo;
        uint256 time;
        bool exists;
    }

    // 交付记录
    struct DeliveryRecord {
        uint256 dealerCompanyId;
        address operatorAddr;
        string deliveryInfo;
        uint256 time;
        bool exists;
    }

    // 完整溯源记录结构体
    struct TraceRecord {
        // 货物信息
        string goodId;
        uint256 ownerCompanyId;
        string goodName;
        uint256 registerTime;
        
        // 运输信息
        uint256 shipCompanyId;
        address shipOperatorAddr;
        string transportInfo;
        uint256 shipTime;
        bool shipExists;
        
        // 验货信息
        uint256 portCompanyId;
        address inspectOperatorAddr;
        string inspectionInfo;
        uint256 inspectTime;
        bool inspectExists;
        
        // 交付信息
        uint256 dealerCompanyId;
        address deliveryOperatorAddr;
        string deliveryInfo;
        uint256 deliveryTime;
        bool deliveryExists;
    }

    // 变量声明
    address public superAdmin;
    uint256 public companyCount = 0;

    mapping(uint256 => Company) public companies;
    mapping(address => uint256) public companyOfAdmin;
    
    mapping(string => Good) private goods;
    mapping(string => ShippingRecord) private shippingRecords;
    mapping(string => InspectionRecord) private inspectionRecords;
    mapping(string => DeliveryRecord) private deliveryRecords;

    // 事件声明
    event CompanyRegistered(uint256 indexed id, string name, CompanyType companyType, address admin);
    event GoodRegistered(string indexed goodId, uint256 ownerCompanyId, string goodName, uint256 registerTime);
    event Shipped(string indexed goodId, uint256 shipCompanyId, address operatorAddr, string info, uint256 time);
    event Inspected(string indexed goodId, uint256 portCompanyId, address operatorAddr, string info, uint256 time);
    event Delivered(string indexed goodId, uint256 dealerCompanyId, address operatorAddr, string info, uint256 time);

    // 修饰符
    modifier onlySuperAdmin() {
        require(msg.sender == superAdmin, "只有超级管理员可执行此操作");
        _;
    }

    modifier onlyCompany(CompanyType companyType) {
        uint256 companyId = companyOfAdmin[msg.sender];
        require(companies[companyId].exists, "公司不存在");
        require(companies[companyId].companyType == companyType, "公司类型不匹配");
        _;
    }

    // 构造函数
    constructor() public {
        superAdmin = msg.sender;
    }

    // 注册公司 (仅超级管理员)
    function registerCompany(
        string memory name,
        CompanyType companyType,
        address admin
    ) public onlySuperAdmin returns (uint256) {
        companyCount++;
        companies[companyCount] = Company(companyCount, name, companyType, admin, true);
        companyOfAdmin[admin] = companyCount;
        emit CompanyRegistered(companyCount, name, companyType, admin);
        return companyCount;
    }

    // 注册货物 (仅生产商)
    function registerGood(
        string memory goodId,
        string memory goodName
    ) public onlyCompany(CompanyType.Producer) returns (bool) {
        uint256 companyId = companyOfAdmin[msg.sender];
        require(!goods[goodId].exists, "货物ID已存在");
        
        goods[goodId] = Good(goodId, companyId, goodName, block.timestamp, true);
        emit GoodRegistered(goodId, companyId, goodName, block.timestamp);
        return true;
    }

    // 船东登记运输
    function shipGood(
        string memory goodId,
        string memory transportInfo
    ) public onlyCompany(CompanyType.Shipper) returns (bool) {
        require(goods[goodId].exists, "货物不存在");
        uint256 companyId = companyOfAdmin[msg.sender];
        require(!shippingRecords[goodId].exists, "该货物已有运输记录");
        
        shippingRecords[goodId] = ShippingRecord(companyId, msg.sender, transportInfo, block.timestamp, true);
        emit Shipped(goodId, companyId, msg.sender, transportInfo, block.timestamp);
        return true;
    }

    // 港口登记验货
    function inspectGood(
        string memory goodId,
        string memory inspectionInfo
    ) public onlyCompany(CompanyType.Port) returns (bool) {
        require(goods[goodId].exists, "货物不存在");
        require(shippingRecords[goodId].exists, "该货物未有运输记录");
        uint256 companyId = companyOfAdmin[msg.sender];
        require(!inspectionRecords[goodId].exists, "该货物已有验货记录");
        
        inspectionRecords[goodId] = InspectionRecord(companyId, msg.sender, inspectionInfo, block.timestamp, true);
        emit Inspected(goodId, companyId, msg.sender, inspectionInfo, block.timestamp);
        return true;
    }

    // 经销商收货登记
    function deliverGood(
        string memory goodId,
        string memory deliveryInfo
    ) public onlyCompany(CompanyType.Dealer) returns (bool) {
        require(goods[goodId].exists, "货物不存在");
        require(shippingRecords[goodId].exists, "该货物未有运输记录");
        require(inspectionRecords[goodId].exists, "该货物未有验货记录");
        uint256 companyId = companyOfAdmin[msg.sender];
        require(!deliveryRecords[goodId].exists, "该货物已有收货记录");
        
        deliveryRecords[goodId] = DeliveryRecord(companyId, msg.sender, deliveryInfo, block.timestamp, true);
        emit Delivered(goodId, companyId, msg.sender, deliveryInfo, block.timestamp);
        return true;
    }

    // 查询货物信息
    function getGood(string memory goodId)
        public
        view
        returns (string memory, uint256, string memory, uint256, bool)
    {
        Good storage g = goods[goodId];
        return (g.goodId, g.ownerCompanyId, g.goodName, g.registerTime, g.exists);
    }

    // 查询运输记录
    function getShippingRecord(string memory goodId)
        public
        view
        returns (uint256, address, string memory, uint256, bool)
    {
        ShippingRecord storage s = shippingRecords[goodId];
        return (s.shipCompanyId, s.operatorAddr, s.transportInfo, s.time, s.exists);
    }

    // 查询验货记录
    function getInspectionRecord(string memory goodId)
        public
        view
        returns (uint256, address, string memory, uint256, bool)
    {
        InspectionRecord storage i = inspectionRecords[goodId];
        return (i.portCompanyId, i.operatorAddr, i.inspectionInfo, i.time, i.exists);
    }

    // 查询收货记录
    function getDeliveryRecord(string memory goodId)
        public
        view
        returns (uint256, address, string memory, uint256, bool)
    {
        DeliveryRecord storage d = deliveryRecords[goodId];
        return (d.dealerCompanyId, d.operatorAddr, d.deliveryInfo, d.time, d.exists);
    }
    
    // 获取完整溯源信息，使用结构体返回，解决参数过多问题
    function getFullTrace(string memory goodId) 
        public 
        view 
        returns (TraceRecord memory) 
    {
        Good storage g = goods[goodId];
        ShippingRecord storage s = shippingRecords[goodId];
        InspectionRecord storage i = inspectionRecords[goodId];
        DeliveryRecord storage d = deliveryRecords[goodId];
        
        return TraceRecord(
            // 货物信息
            g.goodId, g.ownerCompanyId, g.goodName, g.registerTime,
            // 运输信息
            s.shipCompanyId, s.operatorAddr, s.transportInfo, s.time, s.exists,
            // 验货信息
            i.portCompanyId, i.operatorAddr, i.inspectionInfo, i.time, i.exists,
            // 交付信息
            d.dealerCompanyId, d.operatorAddr, d.deliveryInfo, d.time, d.exists
        );
    }
    
    // 获取完整溯源信息，使用数组返回
function getFullTraceArray(string memory goodId) 
    public 
    view 
    returns (
        string memory,  // 0: goodId
        uint[4] memory, // 1: [ownerCompanyId, shipCompanyId, portCompanyId, dealerCompanyId]
        address[3] memory, // 2: [shipOperatorAddr, inspectOperatorAddr, deliverOperatorAddr]
        string[4] memory, // 3: [goodName, transportInfo, inspectionInfo, deliveryInfo]
        uint[4] memory,  // 4: [registerTime, shipTime, inspectTime, deliverTime]
        bool[3] memory   // 5: [shipExists, inspectExists, deliverExists]
    ) 
{
    Good storage g = goods[goodId];
    ShippingRecord storage s = shippingRecords[goodId];
    InspectionRecord storage i = inspectionRecords[goodId];
    DeliveryRecord storage d = deliveryRecords[goodId];
    
    uint[4] memory companies = [g.ownerCompanyId, s.shipCompanyId, i.portCompanyId, d.dealerCompanyId];
    address[3] memory operators = [s.operatorAddr, i.operatorAddr, d.operatorAddr];
    string[4] memory infos = [g.goodName, s.transportInfo, i.inspectionInfo, d.deliveryInfo];
    uint[4] memory times = [g.registerTime, s.time, i.time, d.time];
    bool[3] memory exists = [s.exists, i.exists, d.exists];
    
    return (goodId, companies, operators, infos, times, exists);
}     



    
    // 获取货物当前状态：0-已创建 1-已运输 2-已验货 3-已交付
    function getGoodStatus(string memory goodId) public view returns (uint8) {
        if (!goods[goodId].exists) {
            revert("货物不存在");
        }
        
        if (deliveryRecords[goodId].exists) {
            return 3; // 已交付
        } else if (inspectionRecords[goodId].exists) {
            return 2; // 已验货
        } else if (shippingRecords[goodId].exists) {
            return 1; // 已运输
        } else {
            return 0; // 已创建
        }
    }
}