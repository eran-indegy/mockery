package mockery

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"bufio"
)

type Cleanup func() error

type OutputStreamProvider interface {
	GetWriter(iface *Interface, pkg string) (io.Writer, error, Cleanup)
}

type StdoutStreamProvider struct {
}

func (this *StdoutStreamProvider) GetWriter(iface *Interface, pkg string) (io.Writer, error, Cleanup) {
	return os.Stdout, nil, func() error { return nil }
}

type FileOutputStreamProvider struct {
	BaseDir   string
	InPackage bool
	TestOnly  bool
	Case      string
}

var (
	hasher        = regexp.MustCompile(`hash\:([0-9a-f]*);`)
	NotChangedErr = errors.New("interface didn't change skipping")
)

func (this *FileOutputStreamProvider) GetWriter(iface *Interface, pkg string) (io.Writer, error, Cleanup) {
	var path string

	caseName := iface.Name
	if this.Case == "underscore" {
		caseName = this.underscoreCaseName(caseName)
	}

	if this.InPackage {
		path = filepath.Join(filepath.Dir(iface.Path), this.filename(caseName))
	} else {
		path = filepath.Join(this.BaseDir, this.filename(caseName))
		os.MkdirAll(filepath.Dir(path), 0755)
		pkg = filepath.Base(filepath.Dir(path))
	}

	f, err := os.Open(path)
	if err == nil {
		r := bufio.NewReader(f)
		line, _, err := r.ReadLine()
		if err == nil {
			match := hasher.FindSubmatch(line)
			if match != nil {
				if string(match[1]) ==  iface.Hash {
					f.Close()
					return nil, NotChangedErr, func() error { return nil }
				}
			}
		}
		f.Close()
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, err, func() error { return nil }
	}

	fmt.Printf("Generating mock for: %s.%s\n", iface.Pkg.Name(), iface.Name)
	return f, nil, func() error {
		return f.Close()
	}
}

func (this *FileOutputStreamProvider) filename(name string) string {
	if this.InPackage && this.TestOnly {
		return "mock_" + name + "_test.go"
	} else if this.InPackage {
		return "mock_" + name + ".go"
	} else if this.TestOnly {
		return name + "_test.go"
	}
	return name + ".go"
}

// shamelessly taken from http://stackoverflow.com/questions/1175208/elegant-python-function-to-convert-camelcase-to-camel-caseo
func (this *FileOutputStreamProvider) underscoreCaseName(caseName string) string {
	rxp1 := regexp.MustCompile("(.)([A-Z][a-z]+)")
	s1 := rxp1.ReplaceAllString(caseName, "${1}_${2}")
	rxp2 := regexp.MustCompile("([a-z0-9])([A-Z])")
	return strings.ToLower(rxp2.ReplaceAllString(s1, "${1}_${2}"))
}
