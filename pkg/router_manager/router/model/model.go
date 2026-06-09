package model

import (
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type BaseModel struct {
	ID        string `gorm:"type:char(36);primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (model *BaseModel) BeforeCreate(scope *gorm.Scope) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	return scope.SetColumn("Id", id.String())
}

type Grafana struct {
	BaseModel
	Provider      string `gorm:"type:varchar(32);default:'grafana';not null"`
	Host          string `gorm:"type:varchar(256);not null;unique_index:ghost"`
	Port          string `gorm:"type:varchar(10);not null"`
	Token         string `gorm:"type:varchar(256);index:idx_token"`
	DatasourceID  string `gorm:"type:varchar(32);default:'1';not null"`
	Tls           bool   `gorm:"default:false;not null"`
	SkipTLSVerify bool   `gorm:"default:false;not null"`
}
