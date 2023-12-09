package error_handling

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gogitfs/pkg/daemon/internal/environment"
	"testing"
)

func TestGetDaemonEnv(t *testing.T) {
	environment.Init("foo")
	envInfo, err := GetDaemonEnv()
	defer CleanupDeamonEnv(envInfo)

	assert.NoError(t, err)
	assert.FileExists(t, envInfo.NamedPipeName, "Named pipe file doesn't exist")
	expected := []string{
		fmt.Sprintf("_DAEMON_NAMED_PIPE=%v", envInfo.NamedPipeName),
	}
	assert.Equal(t, expected, envInfo.Env, "incorrect environment")
	assert.Contains(t, envInfo.NamedPipeName, "foo", "program name should be contained in the pipe name")
}
