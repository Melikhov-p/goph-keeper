package external_storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveFileData(t *testing.T) {
	testCases := []struct {
		name             string
		path             string
		userID           int
		content          []byte
		expectedFilePath string
		wantErr          bool
	}{
		{
			name:             "success",
			path:             "C:\\Users\\melik\\GolandProjects\\goph-keeper\\external_store",
			userID:           1,
			content:          []byte("hello world"),
			expectedFilePath: "C:\\Users\\melik\\GolandProjects\\goph-keeper\\external_store\\user_1\\",
			wantErr:          false,
		},
		{
			name:             "pathDoesNotExist",
			path:             "C:\\Users\\melik\\GolandProjects\\goph-keeper\\external_store1234",
			userID:           1,
			content:          []byte("hello world"),
			expectedFilePath: "C:\\Users\\melik\\GolandProjects\\goph-keeper\\external_store\\user_1\\",
			wantErr:          true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			_, filePath, err := SaveFileData(context.Background(), test.userID, test.path, test.content)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, filePath, test.expectedFilePath)
			}
		})
	}
}
