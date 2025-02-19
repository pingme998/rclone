package rc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pingme998/rclone/fs"
	"github.com/pingme998/rclone/fs/config/obscure"
)

func TestMain(m *testing.M) {
	// Pretend to be rclone version if we have a version string parameter
	if os.Args[len(os.Args)-1] == "version" {
		fmt.Printf("rclone %s\n", fs.Version)
		os.Exit(0)
	}
	// Pretend to error if we have an unknown command
	if os.Args[len(os.Args)-1] == "unknown_command" {
		fmt.Printf("rclone %s\n", fs.Version)
		fmt.Fprintf(os.Stderr, "Unknown command\n")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestInternalNoop(t *testing.T) {
	call := Calls.Get("rc/noop")
	assert.NotNil(t, call)
	in := Params{
		"String": "hello",
		"Int":    42,
	}
	out, err := call.Fn(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, in, out)
}

func TestInternalError(t *testing.T) {
	call := Calls.Get("rc/error")
	assert.NotNil(t, call)
	in := Params{}
	out, err := call.Fn(context.Background(), in)
	require.Error(t, err)
	require.Nil(t, out)
}

func TestInternalList(t *testing.T) {
	call := Calls.Get("rc/list")
	assert.NotNil(t, call)
	in := Params{}
	out, err := call.Fn(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, Params{"commands": Calls.List()}, out)
}

func TestCorePid(t *testing.T) {
	call := Calls.Get("core/pid")
	assert.NotNil(t, call)
	in := Params{}
	out, err := call.Fn(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, out)
	pid := out["pid"]
	assert.NotEqual(t, nil, pid)
	_, ok := pid.(int)
	assert.Equal(t, true, ok)
}

func TestCoreMemstats(t *testing.T) {
	call := Calls.Get("core/memstats")
	assert.NotNil(t, call)
	in := Params{}
	out, err := call.Fn(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, out)
	sys := out["Sys"]
	assert.NotEqual(t, nil, sys)
	_, ok := sys.(uint64)
	assert.Equal(t, true, ok)
}

func TestCoreGC(t *testing.T) {
	call := Calls.Get("core/gc")
	assert.NotNil(t, call)
	in := Params{}
	out, err := call.Fn(context.Background(), in)
	require.NoError(t, err)
	require.Nil(t, out)
	assert.Equal(t, Params(nil), out)
}

func TestCoreVersion(t *testing.T) {
	call := Calls.Get("core/version")
	assert.NotNil(t, call)
	in := Params{}
	out, err := call.Fn(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, fs.Version, out["version"])
	assert.Equal(t, runtime.GOOS, out["os"])
	assert.Equal(t, runtime.GOARCH, out["arch"])
	assert.Equal(t, runtime.Version(), out["goVersion"])
	_ = out["isGit"].(bool)
	v := out["decomposed"].([]int64)
	assert.True(t, len(v) >= 2)
}

func TestCoreObscure(t *testing.T) {
	call := Calls.Get("core/obscure")
	assert.NotNil(t, call)
	in := Params{
		"clear": "potato",
	}
	out, err := call.Fn(context.Background(), in)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, in["clear"], obscure.MustReveal(out["obscured"].(string)))
}

func TestCoreQuit(t *testing.T) {
	//The call should return an error if param exitCode is not parsed to int
	call := Calls.Get("core/quit")
	assert.NotNil(t, call)
	in := Params{
		"exitCode": "potato",
	}
	_, err := call.Fn(context.Background(), in)
	require.Error(t, err)
}

// core/command: Runs a raw rclone command
func TestCoreCommand(t *testing.T) {
	call := Calls.Get("core/command")

	test := func(command string, returnType string, wantOutput string, fail bool) {
		var rec = httptest.NewRecorder()
		var w http.ResponseWriter = rec

		in := Params{
			"command":   command,
			"opt":       map[string]string{},
			"arg":       []string{},
			"_response": w,
		}
		if returnType != "" {
			in["returnType"] = returnType
		} else {
			returnType = "COMBINED_OUTPUT"
		}
		stream := strings.HasPrefix(returnType, "STREAM")
		got, err := call.Fn(context.Background(), in)
		if stream && fail {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		if !stream {
			assert.Equal(t, wantOutput, got["result"])
			assert.Equal(t, fail, got["error"])
		} else {
			assert.Equal(t, wantOutput, rec.Body.String())
		}
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	}

	version := fmt.Sprintf("rclone %s\n", fs.Version)
	errorString := "Unknown command\n"
	t.Run("OK", func(t *testing.T) {
		test("version", "", version, false)
	})
	t.Run("Fail", func(t *testing.T) {
		test("unknown_command", "", version+errorString, true)
	})
	t.Run("Combined", func(t *testing.T) {
		test("unknown_command", "COMBINED_OUTPUT", version+errorString, true)
	})
	t.Run("Stderr", func(t *testing.T) {
		test("unknown_command", "STREAM_ONLY_STDERR", errorString, true)
	})
	t.Run("Stdout", func(t *testing.T) {
		test("unknown_command", "STREAM_ONLY_STDOUT", version, true)
	})
	t.Run("Stream", func(t *testing.T) {
		test("unknown_command", "STREAM", version+errorString, true)
	})
}
