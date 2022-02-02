package server

type Config struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Mode               string `json:"mode"`
	JwtIssuer          string `json:"JwtIssuer"`
	JwtSecret          string `json:"JwtSecret"`
	JwtTokenPrefix     string `json:"JwtTokenPrefix"`
	AccessTokenMaxAge  int    `json:"accessTokenMaxAge"`
	AccessTokenHead    string `json:"AccessTokenHead"`
	RefreshTokenMaxAge int    `json:"refreshTokenMaxAge"`
	RefreshTokenHead   string `json:"refreshTokenHead"`
}
