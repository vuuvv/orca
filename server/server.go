package server

type Server interface {
	Start()
	Mount(controllers ...interface{}) Server
}
