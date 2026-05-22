package repository

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// TransactionManager 统一的事务管理器和 Repository 工厂
type TransactionManager struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewTransactionManager(db *gorm.DB, rdb *redis.Client) *TransactionManager {
	return &TransactionManager{
		db:  db,
		rdb: rdb,
	}
}

// Transaction 执行事务操作，将包含了 tx 作为底层的 TransactionManager 交给 fn 使用
func (tm *TransactionManager) Transaction(fn func(txManager *TransactionManager) error) error {
	return tm.db.Transaction(func(tx *gorm.DB) error {
		txManager := NewTransactionManager(tx, tm.rdb)
		return fn(txManager)
	})
}

func (tm *TransactionManager) OrderRepo() *OrderRepo {
	return NewOrderRepo(tm.db, tm.rdb)
}

func (tm *TransactionManager) GoodsRepo() *GoodsRepo {
	return NewGoodsRepo(tm.db, tm.rdb)
}

func (tm *TransactionManager) UserRepo() *UserRepo {
	return NewUserRepo(tm.db, tm.rdb)
}
