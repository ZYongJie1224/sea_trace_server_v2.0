package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// CompanyType 公司类型枚举
type CompanyType int

const (
	Producer CompanyType = iota // 货主/生产商
	Shipper                     // 船东/运输商
	Port                        // 港口/验货商
	Dealer                      // 经销商
)

// CompanyTypeMap 公司类型映射
var CompanyTypeMap = map[CompanyType]string{
	Producer: "生产商",
	Shipper:  "运输商",
	Port:     "验货商",
	Dealer:   "经销商",
}

// Company 公司模型
type Company struct {
	ID               int         `orm:"pk;auto" json:"id"`
	CompanyName      string      `orm:"size(100);unique" json:"company_name"`
	CompanyType      CompanyType `orm:"default(0)" json:"company_type"` // 0=生产商，1=运输商，2=验货商，3=经销商
	Address          string      `orm:"size(255)" json:"address"`
	Contact          string      `orm:"size(50)" json:"contact"`
	Phone            string      `orm:"size(20)" json:"phone"`
	CreatedAt        time.Time   `orm:"auto_now_add" json:"created_at"`
	UpdatedAt        time.Time   `orm:"auto_now" json:"updated_at"`
	BlockchainTxHash string      `orm:"size(66);null" json:"blockchain_tx_hash"` // 区块链交易哈希
}

// TableName 指定表名
func (c *Company) TableName() string {
	return "companies"
}

// GetCompanyTypeName 获取公司类型名称
func (c *Company) GetCompanyTypeName() CompanyType {
	// typeNames := map[int]string{
	// 0: "生产商",
	// 1: "运输商",
	// 2: "验货商",
	// 3: "经销商",
	// }
	// if name, ok := typeNames[c.CompanyType]; ok {
	// 	return name
	// }
	return c.CompanyType
}

// GetCompanyOperators 获取公司操作员列表，支持分页和搜索
func GetCompanyOperators(companyID, page, pageSize int, search string) ([]*User, int64, error) {
	o := orm.NewOrm()
	query := o.QueryTable("users")

	// 基本条件
	baseQuery := query.Filter("role", "operator")

	// 如果指定了公司ID（非0），则过滤该公司的操作员
	if companyID > 0 {
		baseQuery = baseQuery.Filter("company_id", companyID)
	}

	// 如果有搜索关键词，添加搜索条件（但保持role筛选）
	if search != "" {
		// 创建一个条件组合
		searchCond := orm.NewCondition()
		// 搜索条件（OR关系）
		searchFields := searchCond.Or("username__icontains", search).
			Or("real_name__icontains", search).
			Or("email__icontains", search).
			Or("phone__icontains", search)

		// 创建主条件：role=operator AND (search conditions)
		mainCond := orm.NewCondition()
		mainCond = mainCond.And("role", "operator")
		if companyID > 0 {
			mainCond = mainCond.And("company_id", companyID)
		}
		mainCond = mainCond.AndCond(searchFields)

		// 应用组合条件
		query = query.SetCond(mainCond)
	} else {
		// 如果没有搜索条件，就使用基本查询
		query = baseQuery
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 计算分页
	offset := (page - 1) * pageSize

	// 查询操作员数据
	var operators []*User
	_, err = query.OrderBy("-id").Limit(pageSize, offset).All(&operators)
	if err != nil {
		return nil, 0, err
	}

	return operators, total, nil
}

// GetUserInfo 获取用户信息
func GetUserInfo(user *User) map[string]interface{} {
	info := map[string]interface{}{
		"id":         user.Id,
		"username":   user.Username,
		"role":       user.Role,
		"real_name":  user.RealName,
		"avatar":     user.AvatarUrl,
		"email":      user.Email,
		"phone":      user.Phone,
		"status":     user.Status,
		"created_at": user.CreatedAt.Format("2006-01-02 15:04:05"), // 格式化创建时间

	}

	if user.Role != "super_admin" && user.CompanyId > 0 {
		o := orm.NewOrm()
		company := &Company{ID: user.CompanyId}
		if err := o.Read(company); err == nil {
			info["company_id"] = company.ID
			info["company_name"] = company.CompanyName
			info["company_type"] = company.CompanyType
		}
	}

	return info

}

// ***********************************
// GetCompanyList 获取公司列表，支持分页、关键词搜索和类型筛选
func GetCompanyList(page, pageSize int, keyword string, companyType int) ([]*Company, int64, error) {
	o := orm.NewOrm()
	query := o.QueryTable(new(Company))

	// 添加公司类型筛选条件（如果提供且不是-1）
	if companyType >= 0 {
		query = query.Filter("company_type", companyType)
	}

	// 添加关键词搜索条件（如果提供）
	if keyword != "" {
		// 使用OR条件搜索多个字段
		cond := orm.NewCondition()
		keywordCondition := cond.Or("company_name__icontains", keyword).
			Or("company_name", keyword).
			Or("phone", keyword).
			Or("address", keyword)
		query = query.SetCond(keywordCondition)
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 计算分页
	offset := (page - 1) * pageSize

	// 查询公司数据
	var companies []*Company
	_, err = query.OrderBy("-id").Limit(pageSize, offset).All(&companies)
	if err != nil {
		return nil, 0, err
	}

	// 对于高级查询需求，可以在这里添加更多处理，比如获取每个公司的用户数量等
	err = enrichCompanyData(companies)
	if err != nil {
		logs.Warning("丰富公司数据失败: %v", err)
		// 继续执行，不影响主流程
	}

	return companies, total, nil
}

// enrichCompanyData 丰富公司数据，如添加统计信息等
func enrichCompanyData(companies []*Company) error {
	// 此处可以添加一些额外的处理，如获取每个公司的用户数量、交易数量等
	// 目前版本暂未实现具体功能，仅为示例

	return nil
}

// GetCompanyByID 根据ID获取公司信息
func GetCompanyByID(id int) (*Company, error) {
	o := orm.NewOrm()
	company := &Company{ID: id}
	err := o.Read(company)
	return company, err
}

//***********************

// CreateCompany 创建公司
func CreateCompany(name string, companyType CompanyType, address, contact, phone string) (*Company, error) {
	company := &Company{
		CompanyName: name,
		CompanyType: companyType,
		Address:     address,
		Contact:     contact,
		Phone:       phone,
	}

	o := GetOrm()
	id, err := o.Insert(company)
	if err != nil {
		logs.Error("创建公司失败 [name=%s, error=%v]", name, err)
		return nil, err
	}

	company.ID = int(id)
	logs.Info("成功创建公司 [id=%d, name=%s, type=%s, time=%s]",
		company.ID, company.CompanyName, company.GetCompanyTypeName(), "2025-05-14 12:08:22")
	return company, nil
}

// UpdateCompany 更新公司
func UpdateCompany(company *Company) error {
	o := GetOrm()
	company.UpdatedAt = time.Now()
	_, err := o.Update(company)
	if err != nil {
		logs.Error("更新公司失败 [id=%d, name=%s, error=%v]", company.ID, company.CompanyName, err)
		return err
	}
	logs.Info("成功更新公司 [id=%d, name=%s, time=%s]",
		company.ID, company.CompanyName, "2025-05-14 12:08:22")
	return nil
}

// DeleteCompany 删除公司
func DeleteCompany(id int) error {
	o := GetOrm()
	company := &Company{ID: id}
	// 先获取公司信息，用于日志记录
	if err := o.Read(company); err != nil {
		logs.Error("删除公司前查询失败 [id=%d, error=%v]", id, err)
		return err
	}

	_, err := o.Delete(company)
	if err != nil {
		logs.Error("删除公司失败 [id=%d, name=%s, error=%v]", id, company.CompanyName, err)
		return err
	}

	logs.Info("成功删除公司 [id=%d, name=%s, time=%s]",
		id, company.CompanyName, "2025-05-14 12:08:22")
	return nil
}

// CheckCompanyNameExists 检查公司名称是否已存在
func CheckCompanyNameExists(name string) (bool, error) {
	o := GetOrm()
	exists := o.QueryTable(new(Company)).Filter("company_name", name).Exist()
	return exists, nil
}

// CountCompanyOperators 统计公司操作员数量
func CountCompanyOperators(companyID int) (int64, error) {
	o := GetOrm()
	count, err := o.QueryTable(new(User)).
		Filter("company_id", companyID).
		Filter("role", "operator").
		Count()
	if err != nil {
		logs.Error("统计公司操作员数量失败 [companyId=%d, error=%v]", companyID, err)
	}
	return count, err
}

// CountCompanyGoods 统计公司货物数量
func CountCompanyGoods(companyID int) (int64, error) {
	o := GetOrm()
	count, err := o.QueryTable(new(Goods)).
		Filter("owner_company_id", companyID).
		Count()
	if err != nil {
		logs.Error("统计公司货物数量失败 [companyId=%d, error=%v]", companyID, err)
	}
	return count, err
}

// CountCompanies 获取公司总数
func CountCompanies() (int64, error) {
	o := GetOrm()
	count, err := o.QueryTable(new(Company)).Count()
	if err != nil {
		logs.Error("统计公司总数失败: %v", err)
	}
	return count, err
}

// GetCompaniesByType 获取指定类型的公司列表
func GetCompaniesByType(companyType CompanyType) ([]*Company, error) {
	o := GetOrm()
	var companies []*Company
	_, err := o.QueryTable(new(Company)).
		Filter("company_type", companyType).
		OrderBy("-id").
		All(&companies)
	if err != nil {
		logs.Error("获取指定类型公司列表失败 [type=%d, error=%v]", companyType, err)
	}
	return companies, err
}

// GetCompanyByName 根据名称获取公司
func GetCompanyByName(name string) (*Company, error) {
	o := GetOrm()
	company := &Company{}
	err := o.QueryTable(new(Company)).
		Filter("company_name", name).
		One(company)
	if err != nil {
		logs.Error("根据名称获取公司失败 [name=%s, error=%v]", name, err)
		return nil, err
	}
	return company, nil
}

// GetBlockchainRegisteredCompanies 获取已在区块链注册的公司
func GetBlockchainRegisteredCompanies() ([]*Company, error) {
	o := GetOrm()
	var companies []*Company
	_, err := o.QueryTable(new(Company)).
		Filter("blockchain_tx_hash__isnull", false).
		Filter("blockchain_tx_hash__gt", "").
		OrderBy("-id").
		All(&companies)
	if err != nil {
		logs.Error("获取已在区块链注册的公司列表失败: %v", err)
	}
	return companies, err
}
