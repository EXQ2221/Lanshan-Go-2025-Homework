package walkdir

import (
	"os"
	"path/filepath"
)

func WalkDir(root string) ([]string, error) {
	var list []string
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		fullpath := filepath.Join(root, entry.Name())

		if entry.IsDir() {
			sublist, err := WalkDir(fullpath)
			if err != nil {
				return nil, err
			}
			list = append(list, sublist...)
		} else {
			list = append(list, fullpath)
		}
	}
	return list, nil
}
