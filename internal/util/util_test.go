package util

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeConvertToInt32(t *testing.T) {
	convertable := 156
	unConvertable := math.MaxUint32 + 1
	minUnConvertable := math.MinInt32 - 1

	testCases := []struct {
		name     string
		value    int
		wantErr  bool
		expected int32
	}{
		{
			name:     "success",
			value:    convertable,
			wantErr:  false,
			expected: int32(convertable),
		},
		{
			name:     "tooBig",
			value:    unConvertable,
			wantErr:  true,
			expected: int32(0),
		},
		{
			name:     "tooShort",
			value:    minUnConvertable,
			wantErr:  true,
			expected: int32(0),
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := SafeConvertToInt32(test.value)

			if !test.wantErr {
				require.NoError(t, err)
				assert.IsType(t, int32(0), res)
			} else {
				require.Error(t, err)
			}

			assert.Equal(t, test.expected, res)
		})
	}
}

func ExampleSafeConvertToInt32() {
	valueInt := 1
	valueInt32, err := SafeConvertToInt32(valueInt)
	fmt.Println("1: ", valueInt32, err)

	bigValueInt := math.MaxUint32 + 1
	bigValueInt32, bigErr := SafeConvertToInt32(bigValueInt)
	fmt.Println("2: ", bigValueInt32, bigErr)

	// Output:
	// 1: 1 nil
	// 2: 0, value is out of range int32
}
