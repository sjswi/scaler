package util

import (
	"fmt"
	"testing"
)

func TestFile(t *testing.T) {
	str := "Log File: mysql-bin.000003, Log Position: 831;"
	fmt.Println(ParseFileAndPos(str))
}
