package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BaseModelWithSoftDelete struct {
	BaseModel
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
