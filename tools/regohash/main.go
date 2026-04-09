package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "read":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Usage: rh read <filepath>")
			os.Exit(1)
		}
		err := readFile(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "write":
		if len(os.Args) != 6 {
			fmt.Fprintln(os.Stderr, "Usage: rh write <filepath> <start_hash> <end_hash> <new_content_string>")
			os.Exit(1)
		}
		err := replaceLines(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "rh (regohash): A CLI tool for reading and editing files using deterministic line hashes.")
	fmt.Fprintln(os.Stderr, "\nUsage:")
	fmt.Fprintln(os.Stderr, "  rh read <filepath>")
	fmt.Fprintln(os.Stderr, "  rh write <filepath> <start_hash> <end_hash> <new_content_string>")
}

const (
	lcgModulus    = 456976 // 26^4
	lcgMultiplier = 27717
	lcgInverse    = 429261
	lcgIncrement  = 314159
)

// indexToHash converts a line index to a pseudo-random 4-letter string (a-z).
func indexToHash(index int) string {
	mapped := int((int64(index)*lcgMultiplier + lcgIncrement) % lcgModulus)
	if mapped < 0 {
		mapped += lcgModulus
	}
	chars := make([]byte, 4)
	for i := 3; i >= 0; i-- {
		chars[i] = byte('a' + (mapped % 26))
		mapped /= 26
	}
	return string(chars)
}

// hashToIndex converts a 4-letter string back to a line index.
func hashToIndex(hash string) (int, error) {
	if len(hash) != 4 {
		return 0, fmt.Errorf("invalid hash length, expected 4")
	}
	mapped := 0
	for i := 0; i < 4; i++ {
		c := hash[i]
		if c < 'a' || c > 'z' {
			return 0, fmt.Errorf("invalid character in hash '%s', must be a-z", hash)
		}
		mapped = mapped*26 + int(c-'a')
	}

	mapped = mapped - lcgIncrement
	if mapped < 0 {
		mapped = (mapped%lcgModulus + lcgModulus) % lcgModulus
	}

	index := int((int64(mapped) * lcgInverse) % lcgModulus)
	return index, nil
}

func readFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineIndex := 0
	for scanner.Scan() {
		hash := indexToHash(lineIndex)
		// Print out the 4 letter hash followed by the line content
		fmt.Printf("%s %s\n", hash, scanner.Text())
		lineIndex++
	}
	return scanner.Err()
}

func replaceLines(filepath, startHash, endHash, newContentStr string) error {
	startIdx, err := hashToIndex(startHash)
	if err != nil {
		return fmt.Errorf("invalid start hash: %v", err)
	}
	endIdx, err := hashToIndex(endHash)
	if err != nil {
		return fmt.Errorf("invalid end hash: %v", err)
	}
	if startIdx > endIdx {
		return fmt.Errorf("start hash (%s) represents a line after end hash (%s)", startHash, endHash)
	}

	// 1. Read existing file content
	f, err := os.Open(filepath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var lines []string
	if f != nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		f.Close()
		if err := scanner.Err(); err != nil {
			return err
		}
	}

	// 2. Process new content string
	var newLines []string
	if len(newContentStr) > 0 {
		// Prevent empty string at the end of slice if there's a trailing newline
		if strings.HasSuffix(newContentStr, "\n") {
			newContentStr = newContentStr[:len(newContentStr)-1]
		}
		// If input was ONLY a newline, it becomes empty, don't split.
		if len(newContentStr) > 0 {
			newLines = strings.Split(newContentStr, "\n")
		} else {
			newLines = []string{""}
		}
	}

	// 3. Construct the result by replacing lines[startIdx : endIdx+1] (inclusive)
	var result []string

	// Add lines before startIdx
	if startIdx < len(lines) {
		result = append(result, lines[:startIdx]...)
	} else {
		// If start index is beyond current lines, keep all existing and pad
		result = append(result, lines...)
		for i := len(lines); i < startIdx; i++ {
			result = append(result, "")
		}
	}

	// Insert the new lines
	result = append(result, newLines...)

	// Add lines after endIdx
	if endIdx+1 < len(lines) {
		result = append(result, lines[endIdx+1:]...)
	}

	// 4. Write back to file
	outFile, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, line := range result {
		// remove any trailing carriage returns (\r) often introduced by copy-pasting
		cleanLine := strings.TrimRight(line, "\r")
		if _, err := writer.WriteString(cleanLine + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}
