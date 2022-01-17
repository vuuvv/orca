package server

type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Mode string `json:"mode"`
}
