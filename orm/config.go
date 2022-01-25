package orm

type Config struct {
	// Dsn data source name eg. username:passwd@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	Dsn string
	// 'postgres' or 'mysql'
	Type  string
	Debug bool
}
