package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	jsoniter "github.com/json-iterator/go"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/govalidator"
	"github.com/vuuvv/orca/orm"
	"github.com/vuuvv/orca/utils"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	pathLib "path"
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
	gin    *gin.Engine
	config *Config
	routes []*Route
	//middlewares []gin.HandlerFunc
}

func NewGinServer(config *Config) *GinServer {
	if config == nil {
		panic("Http server config is nil!")
	}

	if config.Port == 0 {
		config.Port = 4000
	}

	if config.AccessTokenMaxAge == 0 {
		config.AccessTokenMaxAge = 15
	}

	if config.AccessTokenHead == "" {
		config.AccessTokenHead = "Authorization"
	}

	if config.RefreshTokenMaxAge == 0 {
		config.RefreshTokenMaxAge = 60
	}

	if config.RefreshTokenHead == "" {
		config.RefreshTokenHead = "RefreshToken"
	}

	if config.JwtSecret == "" {
		config.JwtSecret = "eyJhbG.JIUzI1NiIsInR5cCI6IkpXVCJ9"
	}

	if config.JwtTokenPrefix == "" {
		config.JwtTokenPrefix = "Bearer"
	}

	if config.JwtIssuer == "" {
		config.JwtIssuer = "orca.vuuvv.com"
	}

	// 如果mode为空，gin会默认设置为debug
	gin.SetMode(config.Mode)

	s := &GinServer{
		gin:    gin.New(),
		config: config,
	}

	binding.Validator = &Validator{}

	return s
}

func (s *GinServer) addr() string {
	if s.config == nil {
		panic("Http server config is nil!")
	}
	return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
}

func (s *GinServer) GetConfig() *Config {
	return s.config
}

func (s *GinServer) Use(handlers ...interface{}) Server {
	for _, item := range handlers {
		switch h := item.(type) {
		case gin.HandlerFunc:
			s.gin.Use(h)
		case func(context2 *gin.Context):
			s.gin.Use(h)
		default:
			panic(fmt.Sprintf("Add gin middleware error: [%s] is not gin.HandleFunc", reflect.TypeOf(item)))
		}
	}
	return s
}

func (s *GinServer) Default() Server {
	return s.Use(MiddlewareId, gin.Logger(), gin.Recovery())
}

func (s *GinServer) AddRoute(route *Route) {
	s.routes = append(s.routes, route)
}

func (s *GinServer) Routes() []*Route {
	return s.routes
}

func (s *GinServer) Mount(controllers ...interface{}) Server {
	for _, c := range controllers {
		switch gc := c.(type) {
		case GinController:
			router := s.gin.Group(gc.Path(), gc.Middlewares()...)
			gc.SetName(gc.Name())
			gc.SetServer(s)
			gc.SetRouter(router)
			gc.Mount(router)
		default:
			panic(fmt.Sprintf("Mount gin controller error: [%s] is not GinController", reflect.TypeOf(c)))
		}
	}
	return s
}

func (s *GinServer) Start() {
	s.Mount(&ActuatorController{})
	//if len(s.middlewares) > 0 {
	//	s.gin.Use(s.middlewares...)
	//}

	srv := &http.Server{
		Addr:    s.addr(),
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
	SetName(name string)
	Path() string
	GetServer() *GinServer
	SetServer(server *GinServer)
	SetRouter(router *gin.RouterGroup)
	GetEngine() *gin.Engine
	Middlewares() []gin.HandlerFunc
	// Mount 挂载Endpoint
	Mount(router *gin.RouterGroup)
}

type BaseController struct {
	name   string
	server *GinServer
	router *gin.RouterGroup
	routes []*Route
}

func (this *BaseController) Name() string {
	panic("implement me")
}

func (this *BaseController) SetName(name string) {
	this.name = name
}

func (this *BaseController) Path() string {
	panic("implement me")
}

func (this *BaseController) GetServer() *GinServer {
	return this.server
}

func (this *BaseController) GetEngine() *gin.Engine {
	return this.server.gin
}

func (this *BaseController) SetServer(server *GinServer) {
	this.server = server
}

func (this *BaseController) SetRouter(router *gin.RouterGroup) {
	this.router = router
}

func (this *BaseController) Middlewares() []gin.HandlerFunc {
	return nil
}

func (this *BaseController) Mount(router *gin.RouterGroup) {
}

func (this *BaseController) Context() *gin.Context {
	return GetContext()
}

func (this *BaseController) Request(method string, path string, handler ...gin.HandlerFunc) *Route {
	if len(handler) == 0 {
		panic("handler should not be nil")
	}
	this.router.Handle(method, path, handler...)
	route := &Route{
		Group:      this.name,
		Path:       pathLib.Join(this.router.BasePath(), path),
		Method:     method,
		Handler:    utils.FunctionName(handler[len(handler)-1]),
		Permission: GuardAuthorization,
	}
	this.server.AddRoute(route)
	return route
}

func (this *BaseController) Get(path string, handler ...gin.HandlerFunc) *Route {
	return this.Request(http.MethodGet, path, handler...)
}

func (this *BaseController) Post(path string, handler ...gin.HandlerFunc) *Route {
	return this.Request(http.MethodPost, path, handler...)
}

func (this *BaseController) Put(path string, handler ...gin.HandlerFunc) *Route {
	return this.Request(http.MethodPut, path, handler...)
}

func (this *BaseController) Delete(path string, handler ...gin.HandlerFunc) *Route {
	return this.Request(http.MethodDelete, path, handler...)
}

func (this *BaseController) ValidForm(value interface{}) (err error) {
	return ParseForm(this.Context(), value)
}

func (this *BaseController) ParseFormIds() (ids []int64, err error) {
	return ParseIds(this.Context())
}

func (this *BaseController) GetQueryInt(key string, required bool) (value int, err error) {
	return ParseQueryInt(this.Context(), key, required)
}

func (this *BaseController) GetQueryInt64(key string, required bool) (value int64, err error) {
	return ParseQueryInt64(this.Context(), key, required)
}

func (this *BaseController) GetPaginator() *orm.Paginator {
	return orm.GetPaginator(this.Context().Query("q"))
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
	rawErr := err
	if errors.HasStack(rawErr) {
		rawErr = errors.Unwrap(rawErr)
		if rawErr == nil {
			rawErr = err
		}
	}
	switch e := rawErr.(type) {
	case govalidator.Error:
		this.SendJson(http.StatusBadRequest, &Response{
			Code:    http.StatusBadRequest,
			Message: e.Error(),
			Detail:  fmt.Sprintf("%+v", e),
		})
	case *govalidator.Error:
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
	msg := err.Error()
	if this.server.config.Mode == gin.DebugMode {
		msg = fmt.Sprintf("%+v", err)
	}
	zap.L().Error(msg, zap.Error(err))
}

type ActuatorController struct {
	BaseController
}

func (this *ActuatorController) Name() string {
	return "应用监控"
}

func (this *ActuatorController) Path() string {
	return "_m_"
}

func (this *ActuatorController) Mount(router *gin.RouterGroup) {
	this.Get("health", this.health).Anonymous().WithName("健康检测")
	this.Get("env", this.env).WithName("查看环境变量")
	this.Get("routes", this.routes).WithName("查看所有路由")
}

func (this *ActuatorController) health(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}

func (this *ActuatorController) routes(ctx *gin.Context) {
	this.Send(this.server.routes)
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

func WriteTokenToHead(ctx *gin.Context, config *Config, accessToken string, refreshToken string) {
	ctx.Writer.Header().Add(config.AccessTokenHead, config.JwtTokenPrefix+" "+accessToken)
	ctx.Writer.Header().Add(config.RefreshTokenHead, config.JwtTokenPrefix+" "+refreshToken)
}

func WriteTokenToCookies(ctx *gin.Context, config *Config, accessToken string, refreshToken string) {
	// max age of  access token and refresh token should be refresh token's max age
	ctx.SetCookie(
		config.AccessTokenHead,
		accessToken,
		config.RefreshTokenMaxAge*60,
		"/",
		"",
		false,
		true,
	)
	ctx.SetCookie(
		config.RefreshTokenHead,
		refreshToken,
		config.RefreshTokenMaxAge*60,
		"/",
		"",
		false,
		true,
	)
}

func GenTokens(config *Config, userId int64, username string, roles []int64) (accessToken string, refreshToken string, err error) {
	accessToken, err = GenAccessToken(
		config.JwtIssuer, time.Duration(config.AccessTokenMaxAge)*time.Minute, config.JwtSecret,
		userId, username, roles,
	)
	if err != nil {
		return "", "", err
		//ctx.JSON(http.StatusInternalServerError, NewError(http.StatusInternalServerError, err.Error()))
		//ctx.Abort()
		//return
	}
	refreshToken, err = GenRefreshToken(
		config.JwtIssuer, time.Duration(config.RefreshTokenMaxAge)*time.Minute, config.JwtSecret, userId,
	)
	if err != nil {
		return "", "", err
	}
	return
}
