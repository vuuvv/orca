package logger

type Config struct {
	Level    string `json:"level"`
	Encoding string `json:"encoding"`
	Name     string `json:"name"`
}
