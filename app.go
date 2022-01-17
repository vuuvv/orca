package orca

import (
	redis2 "github.com/go-redis/redis/v8"
	"github.com/vuuvv/orca/config"
	"github.com/vuuvv/orca/database"
	"github.com/vuuvv/orca/logger"
	"github.com/vuuvv/orca/redis"
	"github.com/vuuvv/orca/serialize"
	"github.com/vuuvv/orca/server"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Application struct {
	configPath   string
	configLoader config.Loader
	httpServer   server.Server
	logger       *zap.Logger
	redisClient  *redis2.Client
	db           *gorm.DB
}

type ApplicationOption func(app *Application)

func WithConfigLoader(loader config.Loader) ApplicationOption {
	return func(app *Application) {
		app.configLoader = loader
	}
}

func WithConfigPath(path string) ApplicationOption {
	return func(app *Application) {
		app.configPath = path
	}
}

func WithHttpServer(server *server.GinServer) ApplicationOption {
	return func(app *Application) {
		app.httpServer = server
	}
}

func WithLogger(logger *zap.Logger) ApplicationOption {
	return func(app *Application) {
		app.logger = logger
	}
}

func NewApplication(opts ...ApplicationOption) *Application {
	app := &Application{
		configPath:   "resources/application.yaml",
		configLoader: config.NewViperConfigLoader(),
	}
	for _, opt := range opts {
		opt(app)
	}

	// jsoniter序列化初始化
	serialize.InitializeJsoniter()

	// 配置加载器初始化
	err := app.configLoader.Load(app.configPath)
	if err != nil {
		panic(err)
	}

	// 日志初始化
	zapConfig := &logger.Config{}
	err = app.UnmarshalConfig(zapConfig, "zap")
	if err != nil {
		panic(err)
	}
	app.logger = logger.NewLogger(zapConfig)

	// redis初始化
	if app.configLoader.IsSet("redis") {
		redisConfig := &redis.Config{}
		app.redisClient, err = redis.NewRedisClient(redisConfig)
		if err != nil {
			panic(err)
		}
	}

	// gorm初始化
	if app.configLoader.IsSet("database") {
		databaseConfig := &database.Config{}
		app.db, err = database.NewGorm(databaseConfig)
		if err != nil {
			panic(err)
		}
	}

	// http服务器初始化
	httpConfig := &server.Config{}
	err = app.UnmarshalConfig(httpConfig, "http")
	if err != nil {
		panic(err)
	}
	app.httpServer = server.NewGinServer(httpConfig)

	if defaultApplication == nil {
		ReplaceDefaultApplication(app)
	}

	return app
}

func (this *Application) GetConfig(name string) interface{} {
	return this.configLoader.Get(name)
}

func (this *Application) UnmarshalConfig(output interface{}, name ...string) error {
	return this.configLoader.Unmarshal(output, name...)
}

func (this *Application) Mount(controllers ...interface{}) *Application {
	this.httpServer.Mount(controllers...)
	return this
}

func (this *Application) SetDefault() {
	ReplaceDefaultApplication(this)
}

func (this *Application) Start() {
	if this.httpServer == nil {
		panic("Application start error: http server is nil")
	}

	this.httpServer.Start()
}

var defaultApplication *Application = nil

func ReplaceDefaultApplication(app *Application) {
	defaultApplication = app
}

func GetConfig(name string) interface{} {
	return defaultApplication.GetConfig(name)
}

func UnmarshalConfig(output interface{}, name ...string) error {
	return defaultApplication.UnmarshalConfig(output, name...)
}

func Redis() *redis2.Client {
	return defaultApplication.redisClient
}

func Database() *gorm.DB {
	return defaultApplication.db
}

func Start(controllers ...interface{}) {
	if defaultApplication == nil {
		defaultApplication = NewApplication()
	}
	defaultApplication.Mount(controllers...).Start()
}
