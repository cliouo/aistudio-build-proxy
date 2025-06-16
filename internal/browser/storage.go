package browser

import (
	"os"
)

func LoadStorage(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}
