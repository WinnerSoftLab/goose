package goose

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestSequential(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "tmptest")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)               // clean up
	defer os.Remove("./bin/create-goose") // clean up

	commands := []string{
		"go build -o ./bin/create-goose ./cmd/goose",
		fmt.Sprintf("./bin/create-goose -s -dir=%s create create_table", dir),
		fmt.Sprintf("./bin/create-goose -s -dir=%s create add_users", dir),
		fmt.Sprintf("./bin/create-goose -s -dir=%s create add_indices", dir),
		fmt.Sprintf("./bin/create-goose -s -dir=%s create update_users", dir),
	}

	for _, cmd := range commands {
		args := strings.Split(cmd, " ")
		time.Sleep(1 * time.Second)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = os.Environ()
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s:\n%v\n\n%s", err, cmd, out)
		}
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	// check that the files are in order
	for i, f := range files {
		expected := fmt.Sprintf("%05v", i+1)
		if !strings.HasPrefix(f.Name(), expected) {
			t.Errorf("failed to find %s prefix in %s", expected, f.Name())
		}
	}
}
func TestCustomTemplate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	defer os.RemoveAll(dir) // clean up

	tmpl := template.Must(template.New("goose.go-migration").Parse(`
{{index .Values "comment"}}
Name {{.CamelName}}
Version {{.Version}}
`))

	SetSequential(true)
	if err := Create(nil, dir, "foo", "go", map[string]string{"comment": "// hello world"}, WithTemplate(tmpl)); err != nil {
		t.Fatal(err)
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		b := strings.Builder{}
		for _, file := range files {
			b.WriteString(file.Name())
			b.WriteString("\n")
		}
		t.Fatalf("should be only one file, got %d: %s", len(files), b.String())
	}

	content, err := ioutil.ReadFile(path.Join(dir, files[0].Name()))
	if err != nil {
		t.Fatal(err)
	}

	expected := `
// hello world
Name Foo
Version 00001
`

	if expected != string(content) {
		t.Errorf("mismatched text: got %v, wanted %v\n", string(content), expected)
	}
}
