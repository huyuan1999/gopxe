package os

import (
	"errors"
	"strings"
)

type PXEOsType interface {
	Default() string                    // 此方法返回引导菜单的 LABEL 信息
	Template(string) (string, error)    // 此方法返回 ks 模板
	Mount(string, string, string) error // 此方法用于将镜像挂载到指定目录(主要用于提供 package 源)
	Boot(string, string) error          // 此方法负责创建对应的 tftp 目录结构并从 Mount 点将需要的文件拷贝放对应目录
}

func NewOsType(osType string) (PXEOsType, error) {
	switch strings.ToUpper(osType) {
	case strings.ToUpper("CentOS"), strings.ToUpper("RedHat"):
		return &CentOS{}, nil
	}
	return nil, errors.New("unsupported system type")
}
