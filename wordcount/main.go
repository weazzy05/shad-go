//go:build !solution

package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {

	lineCounts := make(map[string]int)

	// Обрабатываем каждый файл, переданный в качестве аргумента
	for _, filename := range os.Args[1:] {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", filename, err)
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			lineCounts[line]++
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
		}
	}

	// Выводим строки, которые встречаются хотя бы дважды
	for line, count := range lineCounts {
		if count >= 2 {
			fmt.Printf("%d\t%s\n", count, line)
		}
	}
}
