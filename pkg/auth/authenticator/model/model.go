package model

import (
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type BaseModel struct {
	ID        string     `gorm:"type:char(36);primary_key"`
	CreatedAt time.Time  `gorm:"index:idx_time"`
	UpdatedAt time.Time  `gorm:"index:idx_time"`
	DeletedAt *time.Time `gorm:"index:idx_time"`
}

func (model *BaseModel) BeforeCreate(scope *gorm.Scope) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	return scope.SetColumn("Id", id.String())
}

type User struct {
	BaseModel
	Username    string `gorm:"type:varchar(128);not null;unique_index:username"`
	Password    string `gorm:"type:varchar(128);not null"`
	Email       string `gorm:"type:varchar(256);index:idx_user"`
	Name        string `gorm:"type:varchar(256)"`
	Permissions string `gorm:"type:varchar(1024);index:idx_user"`
	IsDisabled  bool   `gorm:"default:false;not null"`
}

type Domain struct {
	BaseModel
	Domain string `gorm:"type:varchar(128);not null;unique_index:idx_domain"`
}
