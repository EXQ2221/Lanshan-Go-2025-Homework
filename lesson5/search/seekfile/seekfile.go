package seekfile

import (
	"bufio"
	"os"
	"strings"
)

type SearchResult struct {
	Path string
	Line []string
}

func Searchfile(path string, keyword string) SearchResult {
	file, err := os.Open(path)
	if err != nil {
		return SearchResult{Path: path}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var matches []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, keyword) {
			matches = append(matches, line)
		}
	}

	return SearchResult{
		Path: path,
		Line: matches,
	}
}
