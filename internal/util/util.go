package util

import (
	"errors"
	"fmt"
	"math"
)

// SafeConvertToInt32 безопасное преобразование int -> int32.
func SafeConvertToInt32(x int) (int32, error) {
	if x > math.MaxInt32 || x < math.MinInt32 {
		return 0, errors.New(fmt.Sprintf("value %d out of range of int32", x))
	}

	return int32(x), nil
}
