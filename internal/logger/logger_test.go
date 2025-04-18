package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildLogger(t *testing.T) {
	debug := "debug"
	universe := "universe"

	testCases := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "success",
			level:   debug,
			wantErr: false,
		},
		{
			name:    "fail",
			level:   universe,
			wantErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			_, err := BuildLogger(test.level)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func ExampleBuildLogger() {
	logLevel := "info"
	log, err := BuildLogger(logLevel)
	if err != nil {
		panic("fail")
	}

	log.Debug("debug msg: don't shown cause of log level", zap.String("level", logLevel))
	log.Info("info msg: show cause of log level", zap.String("logLevel", logLevel))

	// Output:
	// {"level":"info","msg":"info msg: show cause of log level", "logLevel":"info"}
}
