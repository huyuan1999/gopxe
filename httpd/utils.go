package httpd

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/sirupsen/logrus"
	"gopxe/config"
	myos "gopxe/os"
	"gopxe/utils"
	"io"
	"mime/multipart"
	"os"
	"path"
)

// 相对原生的 gin.SaveUploadedFile 方法来说, 对大文件更加友好
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	fd, err := file.Open()
	if err != nil {
		return err
	}

	bufSize := 1024 * 1024 * 10
	buffer := make([]byte, bufSize)
	wf, err := os.OpenFile(dst, os.O_CREATE|os.O_APPEND|os.O_SYNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer func() { _ = wf.Close() }()

	for {
		bytesRead, err := fd.Read(buffer)
		if err != nil {
			if err == io.EOF {
				if _, err := wf.Write(buffer[:bytesRead]); err != nil {
					return err
				}
				break
			}
			return err
		}

		if _, err := wf.Write(buffer[:bytesRead]); err != nil {
			return err
		}

		_ = wf.Sync()
	}
	return nil
}

func upload(file *multipart.FileHeader) error {
	if err := utils.MkdirAll(config.IsoSave); err != nil {
		return err
	}
	fileName := path.Join(config.IsoSave, file.Filename)
	if utils.IsFile(fileName) {
		logrus.Warningf("the file %s already exists", file.Filename)
		return nil
	}
	if err := SaveUploadedFile(file, fileName); err != nil {
		return nil
	}
	return nil
}


// 将 label 模板合并到 default 模板中
func addDefaultLabel(pxe myos.PXEOsType, form OSForm) error {
	param := make(map[string]interface{})
	param["kernel"] = path.Join(form.Type, form.Name, "vmlinuz")
	param["initrd"] = path.Join(form.Type, form.Name, "initrd.img")
	param["ks"] = fmt.Sprintf("http://{{ address }}/ks/%s/%s/%s/kickstart/", form.Type, form.Name, form.Version)
	param["IS_DEFAULT"] = fmt.Sprintf("{{ %s }}", form.Name)
	param["name"] = form.Name

	tmpl, err := pongo2.FromString(pxe.Default())
	if err != nil {
		return err
	}

	booter, err := tmpl.Execute(param)
	if err != nil {
		return err
	}

	pxeLinux := path.Join(config.Tftp, "pxelinux.cfg", "default")
	fd, err := os.OpenFile(pxeLinux, os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	_, _ = fd.WriteString("\n\n\n")
	if _, err = fd.WriteString(booter); err != nil {
		return err
	}
	return nil
}

// 将用户上传的 iso 镜像挂载到本地
func mount(pxe myos.PXEOsType, form OSForm) error {
	mountPath := path.Join(config.IsoMount, form.Type, form.Name)
	if err := utils.MkdirAll(mountPath); err != nil {
		return err
	}
	fileName := path.Join(config.IsoSave, form.Iso.Filename)
	return pxe.Mount(fileName, mountPath, "iso9660")
}

func boot(pxe myos.PXEOsType, form OSForm) error {
	mountPath := path.Join(config.IsoMount, form.Type, form.Name)
	target := path.Join(config.Tftp, form.Type, form.Name)
	if err := utils.MkdirAll(target); err != nil {
		return err
	}
	return pxe.Boot(target, mountPath)
}
