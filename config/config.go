package config

import "gorm.io/gorm"

var (
	Device  string
	Address string
	Port    int
	Db      *gorm.DB
)

const (
	Tftp      = "./tftp/"
	IsoMount  = "./iso_mount"
	IsoSave   = "./iso/"
)
