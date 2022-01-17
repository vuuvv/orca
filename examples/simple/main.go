package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vuuvv/orca"
	"github.com/vuuvv/orca/server"
	"net/http"
)

type indexController struct {
	server.BaseController
}

func (this *indexController) Name() string {
	return "index"
}

func (this *indexController) Path() string {
	return ""
}

func (this *indexController) Mount(router *gin.RouterGroup) {
	router.GET("", this.Index)
	router.GET("id", server.MiddlewareId, this.IndexWithMiddleware)
}

func (this *indexController) GetQuery(ctx *gin.Context) {

}

func (this *indexController) PostQuery(ctx *gin.Context) {

}

func (this *indexController) Index(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Hello World")
}

func (this *indexController) IndexWithMiddleware(ctx *gin.Context) {
	this.Send("Hello World")
	//ctx.String(http.StatusOK, "Hello World")
}

func main() {
	orca.Start(&indexController{})
}
