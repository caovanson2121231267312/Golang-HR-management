package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"hr-management-system/internal/config"

	_ "github.com/lib/pq"
)

type Database struct {
	*sql.DB
}

var db *Database

func NewConnection(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := cfg.GetDSN()
	
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(time.Minute * 5)
	
	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	db = &Database{sqlDB}
	return db, nil
}

func GetDB() *Database {
	return db
}

func (d *Database) Close() error {
	return d.DB.Close()
}

// Transaction helper
func (d *Database) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	
	return tx.Commit()
}

// Health check
func (d *Database) HealthCheck(ctx context.Context) error {
	return d.PingContext(ctx)
}

// Pagination helper
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Pages    int `json:"pages"`
}

func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

func (p *Pagination) GetLimit() int {
	return p.PageSize
}

func NewPagination(page, pageSize int) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p *Pagination) SetTotal(total int) {
	p.Total = total
	p.Pages = (total + p.PageSize - 1) / p.PageSize
}
