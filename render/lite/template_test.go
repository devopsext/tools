package lite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTextTemplateRender(t *testing.T) {

	const envKey = "LITE_TEMPLATE_TEST_VAR"
	if err := os.Setenv(envKey, "hello world"); err != nil {
		t.Fatalf("Setenv() error = %v", err)
	}
	defer os.Unsetenv(envKey)

	content := strings.Join([]string{
		`{{ env "` + envKey + `" | toUpper }}`,
		`{{ regexReplaceAll "o" "0" "foo" }}`,
		`{{ toJson (dict "a" 1) }}`,
	}, "|")

	options := TemplateOptions{
		Name:    "test",
		Content: content,
	}

	tpl, err := NewTextTemplate(options, nil)
	if err != nil {
		t.Fatalf("NewTextTemplate() error = %v", err)
	}

	b, err := tpl.RenderObject(nil)
	if err != nil {
		t.Fatalf("RenderObject() error = %v", err)
	}

	got := string(b)
	want := `HELLO WORLD|f00|{"a":1}`
	if got != want {
		t.Errorf("RenderObject() = %q, want %q", got, want)
	}
}

func TestTemplateFuncMap(t *testing.T) {

	tpl := &Template{}
	funcs := tpl.FuncMap()

	for _, name := range []string{"env", "toUpper", "regexReplaceAll", "toJson", "templateRender", "dirCreate"} {
		if _, ok := funcs[name]; !ok {
			t.Errorf("FuncMap() missing function %q", name)
		}
	}

	for _, name := range []string{"jiraCreateIssue", "awsS3GetObject", "k8sResourceDescribe", "vmReset"} {
		if _, ok := funcs[name]; ok {
			t.Errorf("FuncMap() unexpectedly contains heavy function %q", name)
		}
	}
}

func TestDirCreateFileCreateDefaultMode(t *testing.T) {

	dir := filepath.Join(t.TempDir(), "sub")
	file := filepath.Join(dir, "file.txt")

	content := strings.Join([]string{
		`{{ dirCreate "` + dir + `" 0 }}`,
		`{{ fileCreate "` + file + `" "hello world" 0 }}`,
	}, "")

	options := TemplateOptions{
		Name:    "test",
		Content: content,
	}

	tpl, err := NewTextTemplate(options, nil)
	if err != nil {
		t.Fatalf("NewTextTemplate() error = %v", err)
	}

	if _, err := tpl.RenderObject(nil); err != nil {
		t.Fatalf("RenderObject() error = %v", err)
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat(dir) error = %v", err)
	}
	if got, want := dirInfo.Mode()&os.ModePerm, os.FileMode(0755); got != want {
		t.Errorf("dir mode = %v, want %v", got, want)
	}

	fileInfo, err := os.Stat(file)
	if err != nil {
		t.Fatalf("Stat(file) error = %v", err)
	}
	if got, want := fileInfo.Mode()&os.ModePerm, os.FileMode(0644); got != want {
		t.Errorf("file mode = %v, want %v", got, want)
	}
}
