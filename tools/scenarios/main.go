package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"slices"
	"strings"
)

func main() {
	keywords := []string{"Scenario:", "Scenario Outline:", "Scenario Template:", "Example:"}
	stops := append(slices.Clone(keywords), "Feature:", "Rule:", "Background:")
	printing := len(os.Args) > 1 && os.Args[1] == "--sha"
	var found, hashes []string
	paths := bufio.NewScanner(os.Stdin)
	for paths.Scan() {
		data, err := os.ReadFile(paths.Text())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		lines := strings.Split(string(data), "\n")
		var tags []string
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "@") {
				tags = append(tags, strings.Fields(line)...)
				continue
			}
			if !slices.ContainsFunc(keywords, func(k string) bool { return strings.HasPrefix(line, k) }) {
				if line != "" && !strings.HasPrefix(line, "#") {
					tags = nil
				}
				continue
			}
			start := i + 1
			content := []string{line}
			for i+1 < len(lines) {
				next := strings.TrimSpace(lines[i+1])
				if strings.HasPrefix(next, "@") || slices.ContainsFunc(stops, func(k string) bool { return strings.HasPrefix(next, k) }) {
					break
				}
				i++
				if next != "" && !strings.HasPrefix(next, "#") {
					content = append(content, next)
				}
			}
			sum := sha256.Sum256([]byte(strings.Join(content, "\n")))
			hash := hex.EncodeToString(sum[:])[:12]
			hashes = append(hashes, fmt.Sprintf("%s:%d: @sha:%s", paths.Text(), start, hash))
			if !slices.Contains(tags, "@sha:"+hash) && !slices.Contains(tags, "@skip") {
				found = append(found, fmt.Sprintf("%s:%d: %s", paths.Text(), start, line))
			}
			tags = nil
		}
	}
	if printing {
		fmt.Println(strings.Join(hashes, "\n"))
		return
	}
	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "a scenario without a matching @sha carries @skip (.shampoo/rules/scenario-annotations.toml):")
		fmt.Fprintln(os.Stderr, strings.Join(found, "\n"))
		os.Exit(1)
	}
}
