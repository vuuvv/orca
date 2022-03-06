package request

import (
	"fmt"
	"os"
)

const EnvServiceHost = "SERVICE_NAME"

func ServiceHost() string {
	return os.Getenv(EnvServiceHost)
}

func ServiceUrl(path string) string {
	return fmt.Sprintf("%s%s", ServiceHost(), path)
}
