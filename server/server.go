package server

type Server interface {
	Start()
	GetConfig() *Config
	Mount(controllers ...interface{}) Server
	Use(handlers ...interface{}) Server
	Default() Server
}
