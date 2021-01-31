package httpd

import (
	"encoding/json"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gin-gonic/gin"
	"gopxe/config"
	"gopxe/model"
	myos "gopxe/os"
	"mime/multipart"
	"net/http"
	"path"
	"regexp"
)

type ResMsg struct {
	Success bool
	Msg     interface{}
	Err     interface{}
}

type OSForm struct {
	model.OS
	Iso     *multipart.FileHeader  `form:"iso" binding:"required"`
	KsParam map[string]interface{} `form:"ks_param"`
}

func kickstart(c *gin.Context) {
	var from model.OS
	osType := c.Param("type")
	osName := c.Param("name")
	version := c.Param("version")

	if err := config.Db.Where("name = ?", osName).First(&from).Error; !checkError(c, err) {
		return
	}

	pxe, err := myos.NewOsType(osType)
	if !checkError(c, err) {
		return
	}

	body, err := pxe.Template(version)
	if !checkError(c, err) {
		return
	}

	tmpl, err := pongo2.FromString(body)
	if !checkError(c, err) {
		return
	}

	param := make(map[string]interface{})

	if from.KsParam == "" {
		if err := json.Unmarshal([]byte(from.KsParam), &param); !checkError(c, err) {
			return
		}
	}

	param["url"] = fmt.Sprintf("http://%s:%d/store/%s/%s/", config.Address, config.Port, osType, osName)
	ks, err := tmpl.Execute(param)
	if !checkError(c, err) {
		return
	}
	c.String(http.StatusOK, ks)
}

func checkError(c *gin.Context, err error) bool {
	if err != nil {
		var resMsg ResMsg
		resMsg.Err = err.Error()
		c.JSON(http.StatusOK, resMsg)
		return false
	}
	return true
}

func checkForValidIdentifiers(name string) error {
	var reIdentifiers = regexp.MustCompile("^[a-zA-Z0-9_]+$")
	if !reIdentifiers.MatchString(name) {
		return fmt.Errorf("context-key '%s' is not a valid identifier", name)
	}
	return nil
}

func create(c *gin.Context) {
	var resMsg ResMsg
	var form OSForm

	if err := c.ShouldBind(&form); !checkError(c, err) {
		return
	}

	if err := checkForValidIdentifiers(form.Name); !checkError(c, err) {
		return
	}

	if err := checkForValidIdentifiers(form.Type); !checkError(c, err) {
		return
	}

	pxe, err := myos.NewOsType(form.Type)
	if !checkError(c, err) {
		return
	}

	if err := upload(form.Iso); !checkError(c, err) {
		return
	}

	if err := mount(pxe, form); !checkError(c, err) {
		return
	}

	ksParam, err := json.Marshal(form.KsParam)
	if !checkError(c, err) {
		return
	}

	form.OS.ISO = path.Join(config.IsoSave, form.Iso.Filename)
	form.OS.KsParam = string(ksParam)

	if err := config.Db.Save(&form.OS).Error; !checkError(c, err) {
		return
	}

	if err := boot(pxe, form); !checkError(c, err) {
		return
	}

	if err := addDefaultLabel(pxe, form); !checkError(c, err) {
		return
	}
	resMsg.Success = true
	c.JSON(http.StatusOK, resMsg)
}

func updateParam(c *gin.Context) {

}

func installed(c *gin.Context) {

}

func list(c *gin.Context) {

}
