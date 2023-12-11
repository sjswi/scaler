package mysql

import (
	"fmt"
	"testing"
)

func TestDefer(t *testing.T) {
	fmt.Println(De())
}

func De() int {
	pools := make([]int, 0, 10)
	pools = append(pools, []int{1, 2, 3}...)
	defer func() {
		pools = pools[1:]
	}()
	return pools[0]
}
