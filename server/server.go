package server

type Server interface {
	Start()
	GetConfig() *Config
	SetAuthorization(value Authorization) Server
	GetAuthorization() Authorization
	Mount(controllers ...interface{}) Server
	Use(handlers ...interface{}) Server
	Default() Server
}
