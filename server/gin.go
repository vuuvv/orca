package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	jsoniter "github.com/json-iterator/go"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/govalidator"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"
)

var httpMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

type GinServer struct {
	gin  *gin.Engine
	addr string
}

func NewGinServer(config *Config) *GinServer {
	if config == nil {
		panic("Http server config is nil!")
	}

	if config.Port == 0 {
		config.Port = 4000
	}

	// 如果mode为空，gin会默认设置为debug
	gin.SetMode(config.Mode)

	s := &GinServer{
		gin:  gin.New(),
		addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
	}

	binding.Validator = &Validator{}

	s.gin.Use(MiddlewareId, gin.Logger(), gin.Recovery())
	s.Mount(&ActuatorController{})

	return s
}

func (s *GinServer) Mount(controllers ...interface{}) Server {
	for _, c := range controllers {
		switch gc := c.(type) {
		case GinController:
			router := s.gin.Group(gc.Path(), gc.Middlewares()...)
			gc.SetEngine(s.gin)
			gc.Mount(router)
		default:
			panic(fmt.Sprintf("Mount gin controller error: [%s] is not GinController", reflect.TypeOf(c)))
		}
	}
	return s
}

func (s *GinServer) Start() {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.gin,
	}

	go func() {
		zap.L().Info("启动http服务", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Panic("启动http服务失败", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server forced to shutdown", zap.Error(err))
	}
	zap.L().Info("Server exited")
}

type GinController interface {
	Name() string
	Path() string
	SetEngine(engine *gin.Engine)
	GetEngine() *gin.Engine
	Middlewares() []gin.HandlerFunc
	// Mount 挂载Endpoint
	Mount(router *gin.RouterGroup)
}

type BaseController struct {
	engine *gin.Engine
}

func (this *BaseController) Name() string {
	panic("implement me")
}

func (this *BaseController) Path() string {
	panic("implement me")
}

func (this *BaseController) GetEngine() *gin.Engine {
	return this.engine
}
func (this *BaseController) SetEngine(engine *gin.Engine) {
	this.engine = engine
}

func (this *BaseController) Middlewares() []gin.HandlerFunc {
	return nil
}

func (this *BaseController) Mount(router *gin.RouterGroup) {
}

func (this *BaseController) Context() *gin.Context {
	return GetContext()
}

func (this *BaseController) Send(value interface{}) {
	switch val := value.(type) {
	case error:
		this.SendError(val)
	default:
		this.SendJson(http.StatusOK, &Response{
			Code: 0,
			Data: value,
		})
	}
}

func (this *BaseController) SendW(value interface{}, err error) {
	if err == nil {
		this.Send(value)
	} else {
		this.SendError(err)
	}
}

func (this *BaseController) SendData(statusCode int, contentType string, data []byte) {
	this.Context().Data(statusCode, contentType, data)
}

func (this *BaseController) SendJson(statusCode int, value interface{}) {
	bytes, err := jsoniter.Marshal(value)
	if err != nil {
		panic(err)
	}
	this.SendData(statusCode, "application/json; charset=utf-8", bytes)
}

func (this *BaseController) SendError(err error) {
	switch e := err.(type) {
	case govalidator.Error:
		this.SendJson(http.StatusBadRequest, &Response{
			Code:    http.StatusBadRequest,
			Message: e.Error(),
			Detail:  fmt.Sprintf("%+v", e),
		})
	case *Error:
		this.SendJson(http.StatusInternalServerError, &Response{
			Code:    e.Code,
			Message: e.Error(),
			Data:    e.Data,
			Detail:  fmt.Sprintf("%+v", e),
		})
	default:
		this.SendJson(http.StatusInternalServerError, &Response{
			Code:    1,
			Message: e.Error(),
			Detail:  fmt.Sprintf("%+v", e),
		})
	}
}

type ActuatorController struct {
	BaseController
}

func (this *ActuatorController) Name() string {
	return "actuator"
}

func (this *ActuatorController) Path() string {
	return "_m_"
}

func (this *ActuatorController) Mount(router *gin.RouterGroup) {
	router.GET("health", this.health)
	router.GET("env", this.env)
	router.GET("routes", this.routes)
}

func (this *ActuatorController) health(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}

func (this *ActuatorController) routes(ctx *gin.Context) {
	this.Send(Routes(this.GetEngine()))
	ctx.String(http.StatusOK, "OK")
}

func (this *ActuatorController) env(ctx *gin.Context) {
	ret := make(map[string]string)
	envList := os.Environ()
	for _, key := range envList {
		i := strings.Index(key, "=")
		if i <= 0 {
			continue
		}
		ret[key[0:i]] = key[i+1:]
	}
	this.Send(ret)
}
