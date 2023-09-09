package config

import (
	"os"
)

func GetDir(app string) (dir string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir = home + "/." + app
	err = os.MkdirAll(dir, 0777)
	return
}
