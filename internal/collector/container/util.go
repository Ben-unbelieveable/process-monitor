package container

import "os"

func readEnv(key string) string {
	return os.Getenv(key)
}
