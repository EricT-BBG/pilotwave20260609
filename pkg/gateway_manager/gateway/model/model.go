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

type Gateway struct {
	BaseModel
	Name           string           `gorm:"type:varchar(256);not null;index:idx_gateway"`
	Host           string           `gorm:"type:text;index:idx_gateway"`
	Description    string           `gorm:"type:text"`
	Namespace      string           `gorm:"type:varchar(256);not null;index:idx_gateway"`
	Ports          []Port           `gorm:"foreignkey:GatewayId"`
	BlackWhiteList []Blackwhitelist `gorm:"foreignkey:GatewayId"`
	IsDisabled     bool             `gorm:"default:false;not null"`
}

type Port struct {
	BaseModel
	GatewayId string `gorm:"type:varchar(36);not null;index:idx_port"`
	Port      int    `gorm:"type:int";not null`
	Protocol  string `gorm:"type:varchar(256);not null"`
	Cert      string `gorm:"type:text"`
	Pkey      string `gorm:"type:text"`
}

type Blackwhitelist struct {
	BaseModel
	GatewayId   string `gorm:"type:varchar(36);not null;index:idx_blackwhitelist"`
	Domain      string `gorm:"type:text;not null;index:idx_blackwhitelist"`
	Description string `gorm:"type:text"`
	Category    string `gorm:"type:varchar(36);not null;index:idx_blackwhitelist"`
}

type Gatewaymapping struct {
	BaseModel
	GatewayId string `gorm:"type:varchar(36);not null;index:idx_gatewaymapping"`
	RouterId  string `gorm:"type:varchar(36);not null"`
}
