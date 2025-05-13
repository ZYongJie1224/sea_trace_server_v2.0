package models

import (
	"errors"
	"strconv"
	"time"

	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/client/orm"
)

// User 用户模型
type User struct {
	ID        int       `orm:"pk;auto" json:"id"`
	Username  string    `orm:"size(50);unique" json:"username"`
	Password  string    `orm:"size(255)" json:"-"`
	RealName  string    `orm:"size(50);null" json:"real_name"`
	AvatarURL string    `orm:"size(255);null" json:"avatar_url"`
	Role      string    `orm:"size(20)" json:"role"` // super_admin, company_admin, operator
	CompanyID int       `orm:"null" json:"company_id"`
	Email     string    `orm:"size(100);null" json:"email"`
	Phone     string    `orm:"size(20);null" json:"phone"`
	Status    int       `orm:"default(1)" json:"status"` // 1=active, 0=inactive
	CreatedAt time.Time `orm:"auto_now_add" json:"created_at"`
	UpdatedAt time.Time `orm:"auto_now" json:"updated_at"`
	LastLogin time.Time `orm:"null" json:"last_login"`
}

// TableName 指定表名
func (u *User) TableName() string {
	return "users"
}

// AddUser 添加用户 (接收完整的用户对象)
func AddUser(user User) string {
	// 密码哈希处理
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return ""
	}

	// 设置哈希后的密码
	user.Password = hashedPassword

	o := orm.NewOrm()
	id, err := o.Insert(&user)
	if err != nil {
		return ""
	}

	// 返回字符串形式的ID
	return strconv.FormatInt(id, 10)
}

// CreateUser 创建用户 (使用分离的参数)
func CreateUser(username, password, realName, role string, companyID int) (*User, error) {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:  username,
		Password:  hashedPassword,
		RealName:  realName,
		Role:      role,
		CompanyID: companyID,
		Status:    1,
	}

	o := orm.NewOrm()
	_, err = o.Insert(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetAllUsers 获取所有用户
func GetAllUsers() []User {
	var users []User
	o := orm.NewOrm()
	_, err := o.QueryTable(new(User)).All(&users)
	if err != nil {
		return []User{}
	}
	return users
}

// GetUser 根据ID获取用户
func GetUser(uid string) (*User, error) {
	id, err := strconv.Atoi(uid)
	if err != nil {
		return nil, err
	}

	o := orm.NewOrm()
	user := &User{ID: id}
	err = o.Read(user)
	return user, err
}

// GetUserByID 根据ID获取用户 (int版本)
func GetUserByID(id int) (*User, error) {
	o := orm.NewOrm()
	user := &User{ID: id}

	if err := o.Read(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	o := orm.NewOrm()
	user := &User{}

	err := o.QueryTable(new(User)).Filter("username", username).One(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser 更新用户
func UpdateUser(uid string, userUpdate *User) (*User, error) {
	id, err := strconv.Atoi(uid)
	if err != nil {
		return nil, err
	}

	o := orm.NewOrm()
	existingUser := &User{ID: id}
	if err = o.Read(existingUser); err != nil {
		return nil, err
	}

	// 更新字段，保留ID不变
	if userUpdate.Username != "" {
		existingUser.Username = userUpdate.Username
	}
	if userUpdate.RealName != "" {
		existingUser.RealName = userUpdate.RealName
	}
	if userUpdate.Role != "" {
		existingUser.Role = userUpdate.Role
	}
	if userUpdate.Email != "" {
		existingUser.Email = userUpdate.Email
	}
	if userUpdate.Phone != "" {
		existingUser.Phone = userUpdate.Phone
	}
	if userUpdate.AvatarURL != "" {
		existingUser.AvatarURL = userUpdate.AvatarURL
	}
	if userUpdate.CompanyID != 0 {
		existingUser.CompanyID = userUpdate.CompanyID
	}
	if userUpdate.Status != 0 {
		existingUser.Status = userUpdate.Status
	}

	// 单独处理密码
	if userUpdate.Password != "" {
		hashedPassword, err := utils.HashPassword(userUpdate.Password)
		if err != nil {
			return nil, err
		}
		existingUser.Password = hashedPassword
	}

	_, err = o.Update(existingUser)
	return existingUser, err
}

// DeleteUser 根据ID删除用户 (字符串版本)
func DeleteUser(uid string) error {
	id, err := strconv.Atoi(uid)
	if err != nil {
		return err
	}

	o := orm.NewOrm()
	_, err = o.Delete(&User{ID: id})
	return err
}

// DeleteUserByID 根据ID删除用户 (int版本)
func DeleteUserByID(id int) error {
	o := orm.NewOrm()
	_, err := o.Delete(&User{ID: id})
	return err
}

// Login 验证用户登录 (布尔版本)
func Login(username, password string) bool {
	o := orm.NewOrm()
	user := &User{Username: username}

	err := o.QueryTable(new(User)).Filter("username", username).One(user)
	if err != nil {
		return false
	}

	// 检查用户状态
	if user.Status != 1 {
		return false
	}

	// 验证密码
	if !utils.CheckPasswordHash(password, user.Password) {
		return false
	}

	// 更新最后登录时间
	user.LastLogin = time.Now()
	o.Update(user, "LastLogin")

	return true
}

// CheckLogin 用户登录 (返回用户对象版本)
func CheckLogin(username, password string) (*User, error) {
	o := orm.NewOrm()
	user := &User{}

	err := o.QueryTable(new(User)).Filter("username", username).One(user)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("用户已禁用")
	}

	// 验证密码
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("密码错误")
	}

	// 更新最后登录时间
	user.LastLogin = time.Now()
	o.Update(user, "LastLogin")

	return user, nil
}

// GetUserInfo 获取用户信息 (包含公司信息)
func GetUserInfo(user *User) map[string]interface{} {
	info := map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"role":      user.Role,
		"real_name": user.RealName,
		"avatar":    user.AvatarURL,
		"email":     user.Email,
		"phone":     user.Phone,
		"status":    user.Status,
	}

	if user.Role != "super_admin" && user.CompanyID > 0 {
		o := orm.NewOrm()
		company := &Company{ID: user.CompanyID}
		if err := o.Read(company); err == nil {
			info["company_id"] = company.ID
			info["company_name"] = company.CompanyName
			info["company_type"] = company.CompanyType
		}
	}

	return info
}

// GetUsersByCompanyID 获取公司所有用户
func GetUsersByCompanyID(companyID int) ([]*User, error) {
	var users []*User
	o := orm.NewOrm()
	_, err := o.QueryTable(new(User)).
		Filter("company_id", companyID).
		All(&users)
	return users, err
}

// GetCompanyAdmins 获取公司管理员
func GetCompanyAdmins(companyID int) ([]*User, error) {
	var admins []*User
	o := orm.NewOrm()
	_, err := o.QueryTable(new(User)).
		Filter("company_id", companyID).
		Filter("role", "company_admin").
		All(&admins)
	return admins, err
}

// GetCompanyOperators 获取公司操作员
func GetCompanyOperators(companyID int) ([]*User, error) {
	var operators []*User
	o := orm.NewOrm()
	_, err := o.QueryTable(new(User)).
		Filter("company_id", companyID).
		Filter("role", "operator").
		All(&operators)
	return operators, err
}

// CountUsers 统计用户数量
func CountUsers() (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable(new(User)).Count()
}
