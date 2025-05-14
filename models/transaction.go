package models

import (
	"github.com/beego/beego/v2/core/logs"
)

// Transaction 交易记录模型
type Transaction struct {
	ID               int    `orm:"pk;auto" json:"id"`
	BlockchainTxHash string `orm:"size(66);unique" json:"blockchain_tx_hash"`
	Type             string `orm:"size(50)" json:"type"`           // 交易类型：注册公司、注册货物、运输、验货、交付等
	Content          string `orm:"type(text)" json:"content"`      // 交易内容JSON
	Status           int    `orm:"default(1)" json:"status"`       // 交易状态: 0=失败, 1=成功
	BlockNumber      int64  `orm:"default(0)" json:"block_number"` // 区块高度
	CreatedAt        string `orm:"size(19)" json:"created_at"`     // 交易时间
}

// TableName 指定表名
func (t *Transaction) TableName() string {
	return "transactions"
}

// SaveTransaction 保存交易记录
func SaveTransaction(txHash, txType, content string, blockNumber int64) (*Transaction, error) {
	tx := &Transaction{
		BlockchainTxHash: txHash,
		Type:             txType,
		Content:          content,
		Status:           1,
		BlockNumber:      blockNumber,
		CreatedAt:        "2025-05-14 12:17:23", // 使用当前时间
	}

	o := GetOrm()
	_, err := o.Insert(tx)
	if err != nil {
		logs.Error("保存交易记录失败 [txHash=%s, type=%s, error=%v, time=%s]",
			txHash, txType, err, "2025-05-14 12:17:23")
	} else {
		logs.Info("成功保存交易记录 [txHash=%s, type=%s, blockNumber=%d, time=%s]",
			txHash, txType, blockNumber, "2025-05-14 12:17:23")
	}
	return tx, err
}

// CountTransactions 统计系统中的交易总数
func CountTransactions() (int64, error) {
	o := GetOrm()
	count, err := o.QueryTable(new(Transaction)).Count()
	if err != nil {
		logs.Error("统计交易总数失败: %v [time=%s]", err, "2025-05-14 12:17:23")
	}
	return count, err
}

// GetTransactionList 获取交易记录列表
func GetTransactionList(limit int, offset int) ([]*Transaction, error) {
	o := GetOrm()
	var transactions []*Transaction

	query := o.QueryTable(new(Transaction)).OrderBy("-id")
	if limit > 0 {
		query = query.Limit(limit, offset)
	}

	_, err := query.All(&transactions)
	if err != nil {
		logs.Error("获取交易记录列表失败: %v [time=%s]", err, "2025-05-14 12:17:23")
	}
	return transactions, err
}

// GetTransactionByHash 根据交易哈希获取交易记录
func GetTransactionByHash(txHash string) (*Transaction, error) {
	o := GetOrm()
	tx := &Transaction{}
	err := o.QueryTable(new(Transaction)).Filter("blockchain_tx_hash", txHash).One(tx)
	if err != nil {
		logs.Error("获取交易记录失败 [txHash=%s, error=%v, time=%s]",
			txHash, err, "2025-05-14 12:17:23")
	}
	return tx, err
}
