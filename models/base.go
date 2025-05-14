package models

import (
	"github.com/beego/beego/v2/client/orm"
)

// GetOrm 获取ORM实例
func GetOrm() orm.Ormer {
	return orm.NewOrm()
}
