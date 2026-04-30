package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// readFileLines reads all lines from a file using a 1MB scanner buffer.
func readFileLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func readFile(path string) error {
	lines, err := readFileLines(path)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	c, err := loadCache()
	if err != nil {
		return err
	}

	state := c.Files[abs]

	// If the hash list doesn't match the file, rebuild it from scratch.
	if len(state.LineHashes) != len(lines) {
		state.LineHashes = buildHashList(c, len(lines))
	}

	for i, line := range lines {
		fmt.Printf("%s %s\n", state.LineHashes[i], line)
	}

	state.LastReadAt = time.Now()
	c.Files[abs] = state
	return saveCache(c)
}

func replaceLines(path, startHash, endHash, newContentStr string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	c, err := loadCache()
	if err != nil {
		return err
	}

	state := c.Files[abs]

	startIdx, err := findHashIndex(state.LineHashes, startHash)
	if err != nil {
		return fmt.Errorf("invalid start hash: %v", err)
	}
	endIdx, err := findHashIndex(state.LineHashes, endHash)
	if err != nil {
		return fmt.Errorf("invalid end hash: %v", err)
	}
	if startIdx > endIdx {
		return fmt.Errorf("start hash (%s) represents a line after end hash (%s)", startHash, endHash)
	}

	lines, err := readFileLines(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	newLines := splitContent(newContentStr)
	newHashes := buildHashList(c, len(newLines))

	// Build new file content.
	var result []string
	result = append(result, lines[:startIdx]...)
	result = append(result, newLines...)
	if endIdx+1 < len(lines) {
		result = append(result, lines[endIdx+1:]...)
	}

	// Build new hash list — keep hashes outside the replaced range, assign fresh ones inside.
	var newHashList []string
	newHashList = append(newHashList, state.LineHashes[:startIdx]...)
	newHashList = append(newHashList, newHashes...)
	if endIdx+1 < len(state.LineHashes) {
		newHashList = append(newHashList, state.LineHashes[endIdx+1:]...)
	}

	// Write file.
	outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()
	if err := writeLines(outFile, result); err != nil {
		return err
	}

	printWriteResult(lines, state.LineHashes, startIdx, endIdx, newLines, newHashes)

	// Update cache — new hash list, bump LastReadAt so the mtime guard doesn't fire.
	state.LineHashes = newHashList
	state.LastReadAt = time.Now()
	c.Files[abs] = state
	return saveCache(c)
}

func previewLines(path, startHash, endHash string) error {
	// Validate hash formats before any I/O so format errors surface immediately.
	if _, err := findHashIndex(nil, startHash); err != nil {
		if isFormatError(startHash) {
			return fmt.Errorf("invalid start hash: %v", err)
		}
	}
	if _, err := findHashIndex(nil, endHash); err != nil {
		if isFormatError(endHash) {
			return fmt.Errorf("invalid end hash: %v", err)
		}
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Read the full file and build/reuse the complete hash list — same as rh read.
	// This registers the file so a write can follow without re-reading.
	lines, err := readFileLines(path)
	if err != nil {
		return err
	}

	c, err := loadCache()
	if err != nil {
		return err
	}

	state := c.Files[abs]

	if len(state.LineHashes) != len(lines) {
		state.LineHashes = buildHashList(c, len(lines))
	}

	startIdx, err := findHashIndex(state.LineHashes, startHash)
	if err != nil {
		return fmt.Errorf("invalid start hash: %v", err)
	}
	endIdx, err := findHashIndex(state.LineHashes, endHash)
	if err != nil {
		return fmt.Errorf("invalid end hash: %v", err)
	}
	if startIdx > endIdx {
		return fmt.Errorf("start hash (%s) represents a line after end hash (%s)", startHash, endHash)
	}

	for i := startIdx; i <= endIdx && i < len(lines); i++ {
		fmt.Printf("%s %s\n", state.LineHashes[i], lines[i])
	}

	state.LastReadAt = time.Now()
	c.Files[abs] = state
	return saveCache(c)
}

// grepFile prints every line in path that matches pattern, prefixed with its hash.
// The full file is registered in the cache (identical to rh read), so a write can
// follow immediately using any of the returned hashes.
func grepFile(path, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %v", err)
	}

	lines, err := readFileLines(path)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	c, err := loadCache()
	if err != nil {
		return err
	}

	state := c.Files[abs]

	if len(state.LineHashes) != len(lines) {
		state.LineHashes = buildHashList(c, len(lines))
	}

	for i, line := range lines {
		if re.MatchString(line) {
			fmt.Printf("%s %s\n", state.LineHashes[i], line)
		}
	}

	state.LastReadAt = time.Now()
	c.Files[abs] = state
	return saveCache(c)
}

func appendToFile(path, content string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	existingData, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var lines []string
	if len(existingData) > 0 {
		raw := strings.TrimSuffix(string(existingData), "\n")
		lines = strings.Split(raw, "\n")
	}

	c, err := loadCache()
	if err != nil {
		return err
	}

	state := c.Files[abs]
	newLines := splitContent(content)
	newHashes := buildHashList(c, len(newLines))

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	if len(existingData) > 0 && !strings.HasSuffix(string(existingData), "\n") {
		if _, err := writer.WriteString("\n"); err != nil {
			return err
		}
	}
	for _, line := range newLines {
		if _, err := writer.WriteString(strings.TrimRight(line, "\r") + "\n"); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	// startIdx is after the last existing line; endIdx is before it (nothing replaced).
	printWriteResult(lines, state.LineHashes, len(lines), len(lines)-1, newLines, newHashes)

	state.LineHashes = append(state.LineHashes, newHashes...)
	state.LastReadAt = time.Now()
	c.Files[abs] = state
	return saveCache(c)
}

// splitContent normalises a raw content string into lines,
// stripping one trailing newline and trimming \r from each line.
func splitContent(s string) []string {
	if len(s) == 0 {
		return nil
	}
	s = strings.TrimSuffix(s, "\n")
	if len(s) == 0 {
		return []string{""}
	}
	parts := strings.Split(s, "\n")
	for i, p := range parts {
		parts[i] = strings.TrimRight(p, "\r")
	}
	return parts
}

// writeLines writes each line followed by a newline to w.
func writeLines(w *os.File, lines []string) error {
	writer := bufio.NewWriter(w)
	for _, line := range lines {
		if _, err := writer.WriteString(strings.TrimRight(line, "\r") + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}
