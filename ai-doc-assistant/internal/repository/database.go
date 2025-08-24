package repository

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ai-doc-assistant/internal/config"
	"ai-doc-assistant/internal/model"
)

// Database 数据库连接
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg config.DatabaseConfig) (*Database, error) {
	dsn := cfg.MySQL.GetDSN()

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层sql.DB以设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return &Database{DB: db}, nil
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Document{},
		&model.DocumentChunk{},
		&model.QueryHistory{},
		&model.DocumentShare{},
	)
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// HealthCheck 健康检查
func (d *Database) HealthCheck() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// SaveQueryHistory 保存查询历史
func (d *Database) SaveQueryHistory(history *model.QueryHistory) error {
	// 确保用户存在，如果不存在则创建
	var user model.User
	err := d.DB.Where("id = ?", history.UserID).First(&user).Error
	if err != nil {
		// 用户不存在，创建一个基本用户记录
		newUser := &model.User{
			ID:       history.UserID,
			Username: history.UserID, // 使用ID作为用户名
			Role:     "user",
			Status:   "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := d.DB.Create(newUser).Error; err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}
	}
	
	return d.DB.Create(history).Error
}

// GetQueryHistoryByUserID 获取用户的查询历史
func (d *Database) GetQueryHistoryByUserID(userID string, limit int) ([]model.QueryHistory, error) {
	var histories []model.QueryHistory
	query := d.DB.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&histories).Error
	return histories, err
}

// GetAllQueryHistory 获取所有查询历史（分页）
func (d *Database) GetAllQueryHistory(page, pageSize int) ([]model.QueryHistory, int64, error) {
	var histories []model.QueryHistory
	var total int64
	
	// 计算总数
	err := d.DB.Model(&model.QueryHistory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	err = d.DB.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&histories).Error
	return histories, total, err
}

// UpdateQueryFeedback 更新查询反馈
func (d *Database) UpdateQueryFeedback(queryID string, satisfactionScore int, feedback string) error {
	return d.DB.Model(&model.QueryHistory{}).
		Where("id = ?", queryID).
		Updates(map[string]interface{}{
			"satisfaction_score": satisfactionScore,
			"feedback":          feedback,
		}).Error
}

// GetQueryHistoryByID 根据ID获取查询历史
func (d *Database) GetQueryHistoryByID(queryID string) (*model.QueryHistory, error) {
	var history model.QueryHistory
	err := d.DB.Where("id = ?", queryID).First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}