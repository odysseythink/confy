package confy

import (
	"fmt"
	"os"
	"path/filepath"
)

func (v *Confy) findConfigFile() (string, error) {
	v.logger.Info("searching for config in paths", "paths", v.configPaths)

	for _, cp := range v.configPaths {
		if file := v.searchInPath(cp); file != "" {
			return file, nil
		}
	}
	return "", ConfigFileNotFoundError{v.configName, fmt.Sprintf("%v", v.configPaths)}
}

func (v *Confy) searchInPath(in string) string {
	v.logger.Debug("searching for config in path", "path", in)
	for _, ext := range SupportedExts {
		fullPath := filepath.Join(in, v.configName+"."+ext)
		v.logger.Debug("checking if file exists", "file", fullPath)
		if b, _ := exists(v.fs, fullPath); b {
			v.logger.Debug("found file", "file", fullPath)
			return fullPath
		}
	}

	if v.configType != "" {
		fullPath := filepath.Join(in, v.configName)
		if b, _ := exists(v.fs, fullPath); b {
			return fullPath
		}
	}
	return ""
}

func exists(fs FS, path string) (bool, error) {
	stat, err := fs.Stat(path)
	if err == nil {
		return !stat.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
