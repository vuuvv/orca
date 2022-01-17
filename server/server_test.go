package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"testing"
)

type indexController struct {
}

func (this *indexController) Name() string {
	return "index"
}

func (this *indexController) Path() string {
	return "index"
}

func (this *indexController) Middlewares() []gin.HandlerFunc {
	return nil
}

func (this *indexController) Mount(router *gin.RouterGroup) {
	router.GET("index", this.Index)
}

func (this *indexController) Index(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Hello World")
}

func tt(i interface{}) {
	reflect.ValueOf(i)
}

func TestServer(t *testing.T) {
	c := &indexController{}
	fmt.Println(fmt.Sprintf("%p, %p, %s", c, c.Index, reflect.ValueOf(c).MethodByName("Index").Interface()))
	tt((&indexController{}).Index)
	NewGinServer(&Config{
		Host: "0.0.0.0",
		Port: 4001,
		Mode: gin.DebugMode,
	}).Mount(&indexController{}).Start()
}
