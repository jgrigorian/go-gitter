package cmd

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestEditorCommandArgsSupportsEditorFlags(t *testing.T) {
	t.Parallel()

	got := editorCommandArgs("code --wait", "/tmp/config.yaml")
	want := []string{"code", "--wait", "/tmp/config.yaml"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("editorCommandArgs() = %q, want %q", got, want)
	}
}

func TestEditorCommandArgsUsesViByDefault(t *testing.T) {
	t.Parallel()

	got := editorCommandArgs("", "/tmp/config.yaml")
	want := []string{"vi", "/tmp/config.yaml"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("editorCommandArgs() = %q, want %q", got, want)
	}
}

func TestWriteCompletionSupportsBash(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	if err := writeCompletion("bash", &output); err != nil {
		t.Fatalf("writeCompletion() error = %v", err)
	}
	if !strings.Contains(output.String(), "go-gitter") {
		t.Error("bash completion does not reference the command name")
	}
}

func TestWriteCompletionRejectsUnsupportedShell(t *testing.T) {
	t.Parallel()

	if err := writeCompletion("nushell", &bytes.Buffer{}); err == nil {
		t.Fatal("writeCompletion() error = nil, want an unsupported shell error")
	}
}
