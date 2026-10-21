package cmd

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func profileTestCommand(run func(*cobra.Command, []string) error) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	root := &cobra.Command{Use: "test", RunE: run, SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("pprof", "", "CPU profile output")
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	return root, &stdout, &stderr
}

// A complete stream checks our stop/flush/close ordering, not pprof's encoding.
func checkProfileFlushed(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path) // #nosec G304 -- Test fixture path under t.TempDir().
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("profile was not flushed: %v", err)
	}
	payload, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("incomplete profile: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	if len(payload) == 0 {
		t.Fatal("empty profile")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("profile permissions = %o, want 600", info.Mode().Perm())
	}
}

func TestPprofLifecycle(t *testing.T) {
	failure := errors.New("command failed")
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "success"},
		{name: "command failure", err: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "cpu.pprof")
			root, stdout, stderr := profileTestCommand(func(command *cobra.Command, _ []string) error {
				_, _ = io.WriteString(command.OutOrStdout(), `{"ok":true}`)
				return tc.err
			})
			root.SetArgs([]string{"--pprof", path})
			if err := executeCommand(root); !errors.Is(err, tc.err) {
				t.Fatalf("execute error = %v, want %v", err, tc.err)
			}
			if stdout.String() != `{"ok":true}` || stderr.Len() != 0 {
				t.Fatalf("profiling changed output: stdout=%q stderr=%q", stdout, stderr)
			}
			checkProfileFlushed(t, path)
		})
	}
}

func TestPprofDisabled(t *testing.T) {
	root, stdout, stderr := profileTestCommand(func(command *cobra.Command, _ []string) error {
		_, err := io.WriteString(command.OutOrStdout(), "normal output")
		return err
	})
	root.SetArgs([]string{})
	if err := executeCommand(root); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "normal output" || stderr.Len() != 0 {
		t.Fatalf("default execution changed: stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestPprofRefusesOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cpu.pprof")
	if err := os.WriteFile(path, []byte("keep this file"), 0600); err != nil {
		t.Fatal(err)
	}
	called := false
	root, _, _ := profileTestCommand(func(_ *cobra.Command, _ []string) error {
		called = true
		return nil
	})
	root.SetArgs([]string{"--pprof", path})
	if err := executeCommand(root); err == nil || !strings.Contains(err.Error(), "profile") {
		t.Fatalf("expected profile output error, got %v", err)
	}
	if called {
		t.Fatal("command ran despite failure to start profiling")
	}
	data, err := os.ReadFile(path) // #nosec G304 -- Test fixture path under t.TempDir().
	if err != nil || string(data) != "keep this file" {
		t.Fatalf("existing file changed: data=%q err=%v", data, err)
	}
}
