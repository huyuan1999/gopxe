package model

import "gorm.io/gorm"

type OS struct {
	gorm.Model
	Name        string `gorm:"unique" form:"name" binding:"required"`
	Type        string `form:"type" binding:"required"`
	ISO         string
	DefaultMenu string `gorm:"default_menu:no" form:"default"`
	KsParam     string
	Version     string `form:"version" binding:"required"`
}
