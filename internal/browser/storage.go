package browser

import "os"

// LoadStorage returns the contents of the storage state file.
func LoadStorage(path string) ([]byte, error) {
	return os.ReadFile(path)
}
