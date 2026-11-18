package files

import (
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Name string
	Data []byte
}

func Read(dir, suffix string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []File
	for _, entry := range entries {
		name, ok := strings.CutSuffix(entry.Name(), suffix)
		if !ok || entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		files = append(files, File{Name: name, Data: data})
	}
	return files, nil
}
