package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// Goods 货物模型
type Goods struct {
	ID             int       `orm:"pk;auto" json:"id"`
	GoodID         string    `orm:"size(64);unique" json:"good_id"`
	GoodName       string    `orm:"size(100)" json:"good_name"`
	OwnerCompanyID int       `orm:"" json:"owner_company_id"`
	Description    string    `orm:"type(text);null" json:"description"`
	CreatedAt      time.Time `orm:"auto_now_add" json:"created_at"`
	UpdatedAt      time.Time `orm:"auto_now" json:"updated_at"`
}

// TableName 指定表名
func (g *Goods) TableName() string {
	return "goods"
}

// GetGoodByID 根据区块链ID获取货物
func GetGoodByID(goodID string) (*Goods, error) {
	o := orm.NewOrm()
	good := &Goods{}
	err := o.QueryTable(good).Filter("good_id", goodID).One(good)
	return good, err
}

// SaveGood 保存货物信息
func SaveGood(goodID, goodName string, ownerCompanyID int, description string) (*Goods, error) {
	good := &Goods{
		GoodID:         goodID,
		GoodName:       goodName,
		OwnerCompanyID: ownerCompanyID,
		Description:    description,
	}

	o := orm.NewOrm()
	_, err := o.Insert(good)
	return good, err
}
