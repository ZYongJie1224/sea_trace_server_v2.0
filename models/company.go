package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
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
	ID          int         `orm:"pk;auto" json:"id"`
	CompanyName string      `orm:"size(100);unique" json:"company_name"`
	CompanyType CompanyType `orm:"" json:"company_type"`
	Address     string      `orm:"size(255);null" json:"address"`
	Contact     string      `orm:"size(50);null" json:"contact"`
	Phone       string      `orm:"size(20);null" json:"phone"`
	CreatedAt   time.Time   `orm:"auto_now_add" json:"created_at"`
	UpdatedAt   time.Time   `orm:"auto_now" json:"updated_at"`
}

// TableName 指定表名
func (c *Company) TableName() string {
	return "companies"
}

// GetCompanyList 获取公司列表
func GetCompanyList() ([]*Company, error) {
	var companies []*Company
	o := orm.NewOrm()
	_, err := o.QueryTable(new(Company)).All(&companies)
	return companies, err
}

// GetCompanyByID 根据ID获取公司
func GetCompanyByID(id int) (*Company, error) {
	o := orm.NewOrm()
	company := &Company{ID: id}
	if err := o.Read(company); err != nil {
		return nil, err
	}
	return company, nil
}

// CreateCompany 创建公司
func CreateCompany(name string, companyType CompanyType, address, contact, phone string) (*Company, error) {
	company := &Company{
		CompanyName: name,
		CompanyType: companyType,
		Address:     address,
		Contact:     contact,
		Phone:       phone,
	}

	o := orm.NewOrm()
	_, err := o.Insert(company)
	if err != nil {
		return nil, err
	}

	return company, nil
}

// UpdateCompany 更新公司
func UpdateCompany(company *Company) error {
	o := orm.NewOrm()
	_, err := o.Update(company)
	return err
}

// DeleteCompany 删除公司
func DeleteCompany(id int) error {
	o := orm.NewOrm()
	_, err := o.Delete(&Company{ID: id})
	return err
}
