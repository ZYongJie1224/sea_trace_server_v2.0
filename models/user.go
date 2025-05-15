package models

import (
	"errors"
	"strconv"
	"time"

	"sea_trace_server_V2.0/utils"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// User 用户模型
// User 用户模型
type User struct {
	Id             int       `orm:"pk;auto" json:"id"`
	Username       string    `orm:"size(50);unique" json:"username"`
	Password       string    `orm:"size(255)" json:"-"`
	RealName       string    `orm:"size(50);null" json:"real_name"`
	AvatarUrl      string    `orm:"size(255);null" json:"avatar_url"`
	Role           string    `orm:"size(20)" json:"role"`                      // super_admin, company_admin, operator
	CompanyId      int       `orm:"column(company_id);null" json:"company_id"` // 确保字段映射正确
	Email          string    `orm:"size(100);null" json:"email"`
	Phone          string    `orm:"size(20);null" json:"phone"`
	Status         int       `orm:"default(1)" json:"status"` // 1=active, 0=inactive
	CreatedAt      time.Time `orm:"auto_now_add" json:"created_at"`
	UpdatedAt      time.Time `orm:"auto_now" json:"updated_at"`
	LastLogin      time.Time `orm:"null" json:"last_login"`
	BlockchainAddr string    `orm:"size(42);null" json:"blockchain_addr"` // 区块链钱包地址
	SignUserId     string    `orm:"size(64);null" json:"sign_user_id"`    // WeBASE-Sign用户ID
	BlockchainType int       `orm:"default(0)" json:"blockchain_type"`    // 区块链用户类型
	CompanyName    string    `orm:"-" json:"company_name"`                // 非数据库字段，仅用于API返回
}

// GetUserList 获取用户列表，支持分页、关键词搜索和角色筛选
// 如果提供了 companyID，则只返回该公司的用户
func GetUserList(page, pageSize int, keyword, roleFilter string, companyID ...int) ([]*User, int64, error) {
	o := orm.NewOrm()
	query := o.QueryTable(new(User))

	// 添加公司ID筛选条件（如果提供）
	if len(companyID) > 0 && companyID[0] > 0 {
		query = query.Filter("company_id", companyID[0])
	}

	// 添加角色筛选条件（如果提供）
	if roleFilter != "" {
		query = query.Filter("role", roleFilter)
	}

	// 添加关键词搜索条件（如果提供）
	if keyword != "" {
		// 使用OR条件搜索多个字段
		cond := orm.NewCondition()
		keywordCondition := cond.Or("username__icontains", keyword).
			Or("real_name__icontains", keyword).
			Or("email__icontains", keyword).
			Or("phone__icontains", keyword)
		query = query.SetCond(keywordCondition)
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 计算分页
	offset := (page - 1) * pageSize

	// 查询用户数据
	var users []*User
	_, err = query.OrderBy("-id").Limit(pageSize, offset).All(&users)
	if err != nil {
		return nil, 0, err
	}

	// 填充公司名称
	err = fillCompanyNames(users)
	if err != nil {
		logs.Warning("填充公司名称失败: %v", err)
		// 继续执行，不影响主流程
	}

	return users, total, nil
}

// fillCompanyNames 为用户数据填充公司名称
func fillCompanyNames(users []*User) error {
	// 收集所有公司ID
	companyIDs := make(map[int]bool)
	for _, user := range users {
		if user.CompanyId > 0 {
			companyIDs[user.CompanyId] = true
		}
	}

	// 如果没有公司ID，直接返回
	if len(companyIDs) == 0 {
		return nil
	}

	// 将公司ID转为切片
	ids := make([]int, 0, len(companyIDs))
	for id := range companyIDs {
		ids = append(ids, id)
	}

	// 查询所有相关公司
	o := orm.NewOrm()
	var companies []*Company
	_, err := o.QueryTable(new(Company)).Filter("id__in", ids).All(&companies)
	if err != nil {
		return err
	}

	// 创建公司ID到名称的映射
	companyMap := make(map[int]string)
	for _, company := range companies {
		companyMap[company.ID] = company.CompanyName
	}

	// 设置用户的公司名称
	for _, user := range users {
		if name, ok := companyMap[user.CompanyId]; ok {
			user.CompanyName = name
		}
	}

	return nil
}

// CountAdmins 统计超级管理员数量
func CountAdmins() (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable(new(User)).Filter("role", "super_admin").Count()
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
func CreateUser(username, password, realName, role string, companyID int, email string, phone string) (*User, error) {
	logs.Emergency(password)
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:  username,
		Password:  hashedPassword,
		RealName:  realName,
		Role:      role,
		Email:     email,
		CompanyId: companyID,
		Status:    1,
		Phone:     phone,
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
	user := &User{Id: id}
	err = o.Read(user)
	return user, err
}

// GetUserByID 根据ID获取用户 (int版本)
func GetUserByID(id int) (*User, error) {
	o := orm.NewOrm()
	user := &User{Id: id}

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

// UpdateLastLogin 更新用户最后登录时间
func UpdateLastLogin(userID int) error {
	o := orm.NewOrm()
	user := &User{Id: userID}
	if err := o.Read(user); err != nil {
		return err
	}

	user.LastLogin = time.Now()
	_, err := o.Update(user, "LastLogin")
	return err
}

// UpdateUser 更新用户
func UpdateUser(uid string, userUpdate *User) (*User, error) {
	id, err := strconv.Atoi(uid)
	if err != nil {
		return nil, err
	}

	o := orm.NewOrm()
	existingUser := &User{Id: id}
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
	if userUpdate.AvatarUrl != "" {
		existingUser.AvatarUrl = userUpdate.AvatarUrl
	}
	if userUpdate.CompanyId != 0 {
		existingUser.CompanyId = userUpdate.CompanyId
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
	_, err = o.Delete(&User{Id: id})
	return err
}

// DeleteUserByID 根据ID删除用户 (int版本)
func DeleteUserByID(id int) error {
	o := orm.NewOrm()
	_, err := o.Delete(&User{Id: id})
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
// func GetUserInfo(user *User) map[string]interface{} {
// 	info := map[string]interface{}{
// 		"id":         user.ID,
// 		"username":   user.Username,
// 		"role":       user.Role,
// 		"real_name":  user.RealName,
// 		"avatar":     user.AvatarURL,
// 		"email":      user.Email,
// 		"phone":      user.Phone,
// 		"status":     user.Status,
// 		"created_at": user.CreatedAt.Format("2006-01-02 15:04:05"), // 格式化创建时间

// 	}

// 	if user.Role != "super_admin" && user.CompanyId > 0 {
// 		o := orm.NewOrm()
// 		company := &Company{ID: user.CompanyId}
// 		if err := o.Read(company); err == nil {
// 			info["company_id"] = company.ID
// 			info["company_name"] = company.CompanyName
// 			info["company_type"] = company.CompanyType
// 		}
// 	}

// 	return info
// }

// GetUsersByCompanyID 获取公司所有用户
func GetUsersByCompanyID(companyID int) ([]*User, error) {
	var users []*User
	o := orm.NewOrm()
	_, err := o.QueryTable(new(User)).
		Filter("company_id", companyID).
		All(&users)
	return users, err
}

// GetCompanyAdmins 获取公司管理员列表
func GetCompanyAdmins(companyID int) ([]*User, error) {
	o := GetOrm()
	var admins []*User
	_, err := o.QueryTable(new(User)).
		Filter("company_id", companyID).
		Filter("role", "company_admin").
		All(&admins)
	if err != nil {
		logs.Error("获取公司管理员列表失败 [companyId=%d, error=%v]", companyID, err)
	}
	return admins, err
}

// GetCompanyOperators 获取公司操作员列表
func GetCompanyOperatorsNoPage(companyID int) ([]*User, error) {
	o := GetOrm()
	var operators []*User
	_, err := o.QueryTable(new(User)).
		Filter("company_id", companyID).
		Filter("role", "operator").
		All(&operators)
	if err != nil {
		logs.Error("获取公司操作员列表失败 [companyId=%d, error=%v]", companyID, err)
	}
	return operators, err
}

// CountUsers 统计用户数量
func CountUsers() (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable(new(User)).Count()
}
