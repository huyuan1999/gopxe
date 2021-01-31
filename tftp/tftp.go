package tftp

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/pin/tftp"
	"github.com/sirupsen/logrus"
	"gopxe/config"
	"gopxe/model"
	"gopxe/utils"
	"io"
	"os"
	"path"
)

type Handler interface {
	Read(string, io.ReaderFrom) error
	Write(string, io.WriterTo) error
	RootDir(string) error
}

type DefaultHandler struct {
	rootDir string
}

func (t *DefaultHandler) RootDir(path string) error {
	if !utils.IsDir(path) {
		if err := os.MkdirAll(path, 0644); err != nil {
			return err
		}
	}
	t.rootDir = path
	return nil
}

func rendering(filepath string) (string, error) {
	var pos model.OS
	param := make(map[string]interface{})

	tmpl, err := pongo2.FromFile(filepath)
	if err != nil {
		logrus.Errorf("Rendering from file: %s", err.Error())
		return "", err
	}

	if err := config.Db.Where("default_menu = ?", "yes").First(&pos).Error; err == nil {
		param[pos.Name] = "MENU default"
	}

	param["address"] = fmt.Sprintf("%s:%d", config.Address, config.Port)

	context, err := tmpl.Execute(param)
	if err != nil {
		logrus.Errorf("Rendering template execute: %s", err.Error())
		return "", err
	}

	return context, nil
}

func (t *DefaultHandler) Read(filename string, rf io.ReaderFrom) error {
	filepath := path.Join(t.rootDir, filename)
	if !utils.IsFile(filepath) {
		logrus.Debugf("%s no such file or directory", filepath)
		return errors.New("no such file or directory")
	}

	if filename == path.Join("pxelinux.cfg", "default") {
		c, err := rendering(filepath)
		if err != nil {
			return err
		}
		buf := bytes.NewBuffer([]byte(c))
		bit, err := rf.ReadFrom(buf)
		if err != nil {
			logrus.Errorf(err.Error())
			return err
		}
		logrus.Debugf("Download %s success %d bytes sent", filename, bit)
		return nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		logrus.Errorf(err.Error())
		return err
	}

	defer func() { _ = file.Close() }()
	bit, err := rf.ReadFrom(file)
	if err != nil {
		logrus.Errorf(err.Error())
		return err
	}
	logrus.Debugf("Download %s success %d bytes sent", filename, bit)
	return nil
}

func (t *DefaultHandler) Write(filename string, wt io.WriterTo) error {
	filepath := path.Join(t.rootDir, filename)
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		logrus.Errorf(err.Error())
		return err
	}

	defer func() { _ = file.Close() }()

	bit, err := wt.WriteTo(file)
	if err != nil {
		logrus.Errorf(err.Error())
		return err
	}
	logrus.Debugf("Write %s success %d bytes received", filename, bit)
	return nil
}

func NewDefaultTFTP(rootDir string) *tftp.Server {
	handler := DefaultHandler{rootDir: rootDir}
	if err := handler.RootDir(rootDir); err != nil {
		logrus.Fatalln("init tftp server error: ", err.Error())
	}
	return tftp.NewServer(handler.Read, handler.Write)
}
