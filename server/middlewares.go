package server

import (
	"github.com/gin-gonic/gin"
	"github.com/vuuvv/goid"
	"sync"
)

var contexts = sync.Map{} //map[int64]*gin.Context{}

func MiddlewareId(ctx *gin.Context) {
	id := goid.Get()
	contexts.Store(id, ctx)
	ctx.Next()
	contexts.Delete(id)
}

func GetContext() *gin.Context {
	id := goid.Get()
	ctx, ok := contexts.Load(id)
	if !ok {
		panic("Context set incorrect, are you use middleware [MiddlewareId]")
	}
	return ctx.(*gin.Context)
}
