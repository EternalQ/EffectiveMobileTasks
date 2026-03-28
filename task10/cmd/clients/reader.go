package clients

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	scanner = bufio.NewScanner(os.Stdin)
)

func ReadInput(prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func ReadInt(prompt string) int32 {
	for {
		input := ReadInput(prompt)
		if input == "" {
			return 0
		}
		val, err := strconv.ParseInt(input, 10, 32)
		if err == nil {
			return int32(val)
		}
		fmt.Println("Введите число")
	}
}
