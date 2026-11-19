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
	writing := len(os.Args) > 1 && os.Args[1] == "--write"
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
		var tagLines []int
		stale := map[int]bool{}
		inserts := map[int]string{}
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "@") {
				tags = append(tags, strings.Fields(line)...)
				tagLines = append(tagLines, i)
				continue
			}
			if !slices.ContainsFunc(keywords, func(k string) bool { return strings.HasPrefix(line, k) }) {
				if line != "" && !strings.HasPrefix(line, "#") {
					tags, tagLines = nil, nil
				}
				continue
			}
			scenario := i
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
			if !slices.Contains(tags, "@sha:"+hash) {
				for _, at := range tagLines {
					stale[at] = true
				}
				indent := lines[scenario][:len(lines[scenario])-len(strings.TrimLeft(lines[scenario], " \t"))]
				inserts[scenario] = indent + "@sha:" + hash
			}
			tags, tagLines = nil, nil
		}
		if !writing || len(inserts) == 0 {
			continue
		}
		var out []string
		for i, line := range lines {
			if stale[i] {
				kept := slices.DeleteFunc(strings.Fields(line), func(tag string) bool { return strings.HasPrefix(tag, "@sha:") })
				if len(kept) > 0 {
					indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
					out = append(out, indent+strings.Join(kept, " "))
				}
			} else {
				out = append(out, line)
			}
			if tag, ok := inserts[i]; ok {
				out = append(out[:len(out)-1], tag, line)
			}
		}
		if err := os.WriteFile(paths.Text(), []byte(strings.Join(out, "\n")), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	if writing {
		return
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
