package database

import (
	"fmt"
	"order-payment-system/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitMySQL 初始化MySQL连接
func InitMySQL(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	var err error
	DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(100)       // 最大打开连接数
	sqlDB.SetMaxIdleConns(10)        // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(3600e9) // 连接最大存活时间 1小时（纳秒）

	return DB, nil
}
