package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"os"
	"regexp"
)

// 判断路径是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 判断路径是否存在, 且为目录
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断路径是否存在, 且为文件
func IsFile(path string) bool {
	if Exists(path) {
		return !IsDir(path)
	}
	return false
}

func MkdirAll(path string) error {
	if IsDir(path) {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

func Mount(source string, target string, fstype string) error {
	mounts, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return err
	}

	re := regexp.MustCompile(target)
	if re.MatchString(string(mounts)) {
		logrus.Warningf("%s is already mounted", target)
		return nil
	}

	result := Shell("mount", "-t", fstype, source, target)
	if result.Err != nil {
		return result.Err
	}
	if result.Code != 0 {
		return fmt.Errorf("exit %d: stdout: %s stderr: %s", result.Code, result.Stdout, result.Stderr)
	}
	return nil
}

func GetAddress(device string) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Name == device {
			addrs, err := iface.Addrs()
			if err != nil {
				return "", err
			}

			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip != nil {
					if ipAddr := ip.To4(); ipAddr != nil {
						return ipAddr.String(), nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("device %s not found", device)
}

func CopyFile(src, dst string) error {
	srcFd, err := os.Open(src)
	if err != nil {
		return err
	}

	defer func() { _ = srcFd.Close() }()

	dstFd, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = dstFd.Close() }()

	_, err = io.Copy(dstFd, srcFd)
	if err != nil {
		return err
	}
	return nil
}
