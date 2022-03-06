package request

import "os"

const EnvServiceName = "SERVICE_NAME"

func ServiceName() string {
	return os.Getenv(EnvServiceName)
}
