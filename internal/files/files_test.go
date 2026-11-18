//go:build medium

package files

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	type file struct {
		path    string
		content string
		mode    os.FileMode
	}
	type dir struct {
		path string
		mode os.FileMode
	}
	tests := []struct {
		name    string
		files   []file
		dirs    []dir
		read    string
		suffix  string
		want    []File
		wantErr bool
	}{
		{
			name:   "when Read with files of the suffix among others then returns those files without the suffix",
			files:  []file{{path: "b.toml", content: "b = 2", mode: 0o600}, {path: "a.toml", content: "a = 1", mode: 0o600}, {path: "c.md", content: "c", mode: 0o600}},
			dirs:   []dir{{path: "d.toml", mode: 0o700}},
			suffix: ".toml",
			want:   []File{{Name: "a", Data: []byte("a = 1")}, {Name: "b", Data: []byte("b = 2")}},
		},
		{name: "when Read with an empty directory then returns no files", suffix: ".toml", want: nil},
		{name: "when Read with a missing directory then returns an error", read: "missing", suffix: ".toml", wantErr: true},
		{name: "when Read with an unreadable file of the suffix then returns an error", files: []file{{path: "a.toml", content: "a = 1", mode: 0o000}}, suffix: ".toml", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, f := range tt.files {
				if err := os.WriteFile(filepath.Join(root, f.path), []byte(f.content), f.mode); err != nil {
					t.Fatal(err)
				}
			}
			for _, d := range tt.dirs {
				if err := os.Mkdir(filepath.Join(root, d.path), d.mode); err != nil {
					t.Fatal(err)
				}
			}

			got, err := Read(filepath.Join(root, tt.read), tt.suffix)

			if (err != nil) != tt.wantErr || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, %v, want %v, error %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}
