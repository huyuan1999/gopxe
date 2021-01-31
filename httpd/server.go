package httpd

import (
	"github.com/gin-gonic/gin"
	"gopxe/config"
	"net/http"
)

func router(engine *gin.Engine) {
	engine.StaticFS("/store", http.Dir(config.IsoMount))
	engine.GET("/ks/:type/:name/:version/kickstart/", kickstart)
	engine.POST("/create/", create)
	engine.GET("/installed/:uuid", installed) // uuid 访问 boot 生成的 uuid 信息
	engine.GET("/list/", list)
}

func Server(addr string) error {
	engine := gin.Default()
	router(engine)
	return engine.Run(addr)
}
