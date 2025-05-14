package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID          int       `orm:"pk;auto" json:"id"`
	Key         string    `orm:"column(config_key);size(50);unique" json:"key"`
	Value       string    `orm:"column(config_value);type(text)" json:"value"`
	Description string    `orm:"size(255);null" json:"description"`
	UpdatedAt   time.Time `orm:"auto_now" json:"updated_at"`
}

// TableName 设置表名
func (s *SystemConfig) TableName() string {
	return "system_config"
}

// GetSystemConfig 获取系统配置
func GetSystemConfig(key string) (*SystemConfig, error) {
	o := orm.NewOrm()
	config := &SystemConfig{Key: key}

	err := o.Read(config, "Key")
	if err != nil {
		return nil, err
	}

	return config, nil
}

// SetSystemConfig 设置系统配置
func SetSystemConfig(key, value, description string) error {
	o := orm.NewOrm()

	// 检查配置是否存在
	config := &SystemConfig{Key: key}
	err := o.Read(config, "Key")

	if err == orm.ErrNoRows {
		// 创建新配置
		config.Value = value
		config.Description = description
		_, err = o.Insert(config)
	} else if err == nil {
		// 更新现有配置
		config.Value = value
		if description != "" {
			config.Description = description
		}
		_, err = o.Update(config, "Value", "Description", "UpdatedAt")
	}

	return err
}
