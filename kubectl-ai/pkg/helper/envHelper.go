package helper

import "os"

func getEnvOrDefault(envName, defVal string) string {
	if env := os.Getenv(envName); env != "" {
		return env
	} else {
		return defVal
	}
}
