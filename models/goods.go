package models

import (
	"time"

	"github.com/beego/beego/v2/core/logs"
)

// Goods 货物模型
type Goods struct {
	ID             int       `orm:"pk;auto" json:"id"`
	GoodID         string    `orm:"size(64);unique" json:"good_id"`
	GoodName       string    `orm:"size(100)" json:"good_name"`
	OwnerCompanyId int       `orm:"column(owner_company_id)" json:"owner_company_id"` // 添加column映射
	Description    string    `orm:"type(text);null" json:"description"`
	CreatedAt      time.Time `orm:"auto_now_add" json:"created_at"`
	UpdatedAt      time.Time `orm:"auto_now" json:"updated_at"`
	// 区块链相关字段
	BlockchainTxHash string `orm:"size(66);null" json:"blockchain_tx_hash"` // 添加区块链交易哈希字段
}

// TableName 指定表名
func (g *Goods) TableName() string {
	return "goods"
}

// GetGoodByID 根据区块链ID获取货物
func GetGoodByID(goodID string) (*Goods, error) {
	o := GetOrm() // 使用公共ORM获取函数
	good := &Goods{}
	err := o.QueryTable(good).Filter("good_id", goodID).One(good)
	if err != nil {
		logs.Error("获取货物信息失败 [goodID=%s, error=%v, time=%s]",
			goodID, err, "2025-05-14 12:17:23")
	}
	return good, err
}

// SaveGood 保存货物信息
func SaveGood(goodID, goodName string, ownerCompanyID int, description string) (*Goods, error) {
	good := &Goods{
		GoodID:         goodID,
		GoodName:       goodName,
		OwnerCompanyId: ownerCompanyID,
		Description:    description,
	}

	o := GetOrm() // 使用公共ORM获取函数
	_, err := o.Insert(good)
	if err != nil {
		logs.Error("保存货物信息失败 [goodID=%s, goodName=%s, error=%v, time=%s]",
			goodID, goodName, err, "2025-05-14 12:17:23")
	} else {
		logs.Info("成功保存货物信息 [goodID=%s, goodName=%s, companyID=%d, time=%s]",
			goodID, goodName, ownerCompanyID, "2025-05-14 12:17:23")
	}
	return good, err
}

// CountGoodsByCompany 统计公司货物数量
func CountGoodsByCompany(companyID int) (int64, error) {
	o := GetOrm()
	count, err := o.QueryTable(new(Goods)).
		Filter("owner_company_id", companyID).
		Count()
	if err != nil {
		logs.Error("统计公司货物数量失败 [companyId=%d, error=%v, time=%s]",
			companyID, err, "2025-05-14 12:17:23")
	}
	return count, err
}

// GetGoodsList 获取货物列表
func GetGoodsList() ([]*Goods, error) {
	o := GetOrm()
	var goods []*Goods
	_, err := o.QueryTable(new(Goods)).OrderBy("-id").All(&goods)
	if err != nil {
		logs.Error("获取货物列表失败: %v [time=%s]", err, "2025-05-14 12:17:23")
	}
	return goods, err
}

// GetGoodsByCompany 获取公司的货物列表
func GetGoodsByCompany(companyID int) ([]*Goods, error) {
	o := GetOrm()
	var goods []*Goods
	_, err := o.QueryTable(new(Goods)).
		Filter("owner_company_id", companyID).
		OrderBy("-id").
		All(&goods)
	if err != nil {
		logs.Error("获取公司货物列表失败 [companyId=%d, error=%v, time=%s]",
			companyID, err, "2025-05-14 12:17:23")
	}
	return goods, err
}

// CountGoods 获取货物总数
func CountGoods() (int64, error) {
	o := GetOrm()
	count, err := o.QueryTable(new(Goods)).Count()
	if err != nil {
		logs.Error("统计货物总数失败: %v [time=%s]", err, "2025-05-14 12:17:23")
	}
	return count, err
}
