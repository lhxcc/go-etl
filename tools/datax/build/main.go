// Copyright 2020 the go-etl Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	mylog "github.com/Breeze0806/go/log"
)

//go:generate go run main.go
var log mylog.Logger = mylog.NewDefaultLogger(os.Stdout, mylog.ErrorLevel, "[datax]")

const (
	readerCode = `package %v

import (
	"github.com/Breeze0806/go-etl/config"
	"github.com/Breeze0806/go-etl/datax/plugin/reader"
)

func init() {
	var err error
	maker := &maker{}
	if err = reader.RegisterReader(maker); err != nil {
		panic(err)
	}
}

var pluginConfig = %v

//NewReaderFromString 创建读取器
func NewReaderFromString(plugin string) (rd reader.Reader, err error) {
	r := &Reader{}
	if r.pluginConf, err = config.NewJSONFromString(plugin); err != nil {
		return nil, err
	}
	rd = r
	return
}

type maker struct{}

func (m *maker) Default() (reader.Reader, error) {
	return NewReaderFromString(pluginConfig)
}
`
	writerCode = `package %v

import (
	"github.com/Breeze0806/go-etl/config"
	"github.com/Breeze0806/go-etl/datax/plugin/writer"
)

func init() {
	var err error
	maker := &maker{}
	if err = writer.RegisterWriter(maker); err != nil {
		panic(err)
	}
}

var pluginConfig = %v

//NewWriterFromString 创建写入器
func NewWriterFromString(plugin string) (wr writer.Writer, err error) {
	w := &Writer{}
	if w.pluginConf, err = config.NewJSONFromString(plugin); err != nil {
		return nil, err
	}
	wr = w
	return
}

type maker struct{}

func (m *maker) Default() (writer.Writer, error) {
	return NewWriterFromString(pluginConfig)
}`
	versionCode = `package main

import (
	"os"
	"fmt"
	"strings"
)

const version = "%v (git commit: %v) complied by %v"

func init() {
	if len(os.Args) > 1 {
		if strings.ToLower(os.Args[1]) == "version" {
			fmt.Println(version)
			os.Exit(0)
		}
	}
}
`
	sourcePath  = "../../../datax/"
	programPath = "../../../cmd/datax/version.go"
)

var (
	testFile = `
package datax

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Breeze0806/go-etl/config"
	"github.com/Breeze0806/go-etl/datax/common/plugin"
	"github.com/Breeze0806/go-etl/datax/common/plugin/loader"
	"github.com/Breeze0806/go-etl/datax/common/spi"
	"github.com/Breeze0806/go-etl/datax/common/spi/reader"
	"github.com/Breeze0806/go-etl/datax/common/spi/writer"
)

type mockTaskPlugin struct {
}

func (m *mockTaskPlugin) Init(ctx context.Context) error {
	return nil
}

func (m *mockTaskPlugin) Destroy(ctx context.Context) error {
	return nil
}

type mockReaderJob struct {
	*plugin.BaseJob
	*mockTaskPlugin
}

func (m *mockReaderJob) Split(ctx context.Context, number int) ([]*config.JSON, error) {
	return nil, nil
}

type mockReaderTask struct {
	*plugin.BaseTask
	*mockTaskPlugin
}

func (m *mockReaderTask) StartRead(ctx context.Context, sender plugin.RecordSender) error {
	return nil
}

type mockWriterJob struct {
	*plugin.BaseJob
	*mockTaskPlugin
}

func (m *mockWriterJob) Split(ctx context.Context, number int) ([]*config.JSON, error) {
	return nil, nil
}

type mockWriterTask struct {
	*mockTaskPlugin
	*writer.BaseTask
}

func (m *mockWriterTask) StartWrite(ctx context.Context, receiver plugin.RecordReceiver) error {
	return nil
}

type mockReader struct{}

func (m *mockReader) Job() reader.Job {
	return &mockReaderJob{}
}

func (m *mockReader) Task() reader.Task {
	return &mockReaderTask{}
}

type mockWriter struct{}

func (m *mockWriter) Job() writer.Job {
	return &mockWriterJob{}
}

func (m *mockWriter) Task() writer.Task {
	return &mockWriterTask{}
}

func TestRegisterReader(t *testing.T) {
	type args struct {
		name   string
		reader spi.Reader
		filename string
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		%s
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Rename(tt.args.filename, tt.args.filename+".tmp")
			defer func() {
				os.Rename(tt.args.filename+".tmp", tt.args.filename)
				tt.err = fmt.Errorf("datax: reader %%v has already registered", tt.args.name)
				if err := recover(); !strings.Contains(err.(error).Error(), tt.err.Error()) {
					t.Errorf("panic err: %%v wantErr: %%v", err, tt.err)
				}
			}()

			loader.RegisterReader(tt.args.name, tt.args.reader)
		})
	}
}

func TestRegisterWriter(t *testing.T) {
	type args struct {
		name   string
		writer spi.Writer
		filename string
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		%s
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Rename(tt.args.filename, tt.args.filename+".tmp")
			defer func() {
				os.Rename(tt.args.filename+".tmp", tt.args.filename)
				tt.err = fmt.Errorf("datax: writer %%v has already registered", tt.args.name)
				if err := recover(); !strings.Contains(err.(error).Error(), tt.err.Error()) {
					t.Errorf("panic err: %%v wantErr: %%v", err, tt.err)
				}
			}()
			loader.RegisterWriter(tt.args.name, tt.args.writer)
		})
	}
}
`
	readerCase = `
		{
			name: "%s",
			args: args{
				name:   "%s",
				reader: &mockReader{},
				filename: "%s",
			},
		},
`

	writerCase = `
		{
			name: "%s",
			args: args{
				name:   "%s",
				writer: &mockWriter{},
				filename: "%s",
			},
		},
`
)

func main() {
	var imports []string
	parser := pluginParser{}
	if err := parser.readPackages(sourcePath + "plugin/reader"); err != nil {
		log.Errorf("readPackages %v", err)
		os.Exit(1)
	}
	br := &strings.Builder{}
	for _, info := range parser.infos {
		if err := info.genFile(sourcePath+"plugin/reader", readerCode); err != nil {
			log.Errorf("genFile %v", err)
			os.Exit(1)
		}
		imports = append(imports, info.genImport("reader"))
		br.WriteString(fmt.Sprintf(readerCase, info.Name, info.Name, "plugin/reader/"+info.shotPackage+"/resources/plugin.json"))
	}

	imports = append(imports, "")
	parser.infos = nil

	bw := &strings.Builder{}
	if err := parser.readPackages(sourcePath + "plugin/writer"); err != nil {
		log.Errorf("readPackages %v", err)
		os.Exit(1)
	}
	for _, info := range parser.infos {
		if err := info.genFile(sourcePath+"plugin/writer", writerCode); err != nil {
			log.Errorf("genFile %v", err)
			os.Exit(1)
		}
		imports = append(imports, info.genImport("writer"))
		bw.WriteString(fmt.Sprintf(writerCase, info.Name, info.Name, "plugin/writer/"+info.shotPackage+"/resources/plugin.json"))
	}

	if err := writeAllPlugins(imports); err != nil {
		log.Errorf("writeAllPlugins fail. err: %v", err)
		os.Exit(1)
	}
	if err := writeAllPluginsTest(br.String(), bw.String()); err != nil {
		log.Errorf("writeAllPlugins fail. err: %v", err)
		os.Exit(1)
	}

	if err := writeVersionCode(); err != nil {
		log.Errorf("writeAllPlugins fail. err: %v", err)
		os.Exit(1)
	}
}

//生成plugin的reader/writer插件文件
type pluginParser struct {
	infos []pluginInfo
}

func (p *pluginParser) readPackages(path string) (err error) {
	var list []os.FileInfo
	list, err = ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	var data []byte
	for _, v := range list {
		if v.IsDir() {
			data, err = ioutil.ReadFile(filepath.Join(path, v.Name(), "resources", "plugin.json"))
			if err != nil {
				err = nil
				continue
			}
			info := pluginInfo{
				shotPackage:  v.Name(),
				pluginConfig: "`" + string(data) + "`",
			}
			err = json.Unmarshal(data, &info)
			if err != nil {
				err = nil
				continue
			}
			p.infos = append(p.infos, info)
		}
	}
	return
}

type pluginInfo struct {
	Name         string `json:"name"`
	shotPackage  string
	pluginConfig string
}

func (p *pluginInfo) genImport(typ string) string {
	return fmt.Sprintf(`	_ "github.com/Breeze0806/go-etl/datax/plugin/%s/%s"`, typ, p.shotPackage)
}

func (p *pluginInfo) genFile(path string, code string) (err error) {
	var f *os.File
	f, err = os.Create(filepath.Join(path, p.shotPackage, "plugin.go"))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, code, p.shotPackage, p.pluginConfig)
	return
}

func writeAllPlugins(imports []string) (err error) {
	var f *os.File
	f, err = os.Create(sourcePath + "plugin.go")
	if err != nil {
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()
	f.WriteString(`package datax

import (
`)
	for _, v := range imports {
		f.WriteString(v)
		f.WriteString("\n")
	}
	_, err = f.WriteString(")\n")
	return
}

func writeAllPluginsTest(readCases, writerCases string) (err error) {
	var f *os.File
	f, err = os.Create(sourcePath + "plugin_test.go")
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf(testFile, readCases, writerCases))

	return
}

func writeVersionCode() (err error) {
	var f *os.File
	f, err = os.Create(programPath)
	if err != nil {
		return
	}
	defer f.Close()
	version := ""
	if version, err = getVersion(); err != nil {
		return
	}
	_, err = f.WriteString(version)
	return
}

//通过git获取git版本号 `tag`` (git commit: `git version`) complied by gp version `go version`
//例如 v0.1.2 (git commit: c26eb4e15751e41d32402cbf3c7f1ea8af4e3e47) complied by go version go1.16.14 windows/amd64
func getVersion() (version string, err error) {
	output := ""
	if output, err = cmdOutput("git", "describe", "--abbrev=0", "--tags"); err != nil {
		err = fmt.Errorf("use git to tag version fail. error: %w", err)
		return
	}
	tagVersion := strings.ReplaceAll(output, "\r", "")
	tagVersion = strings.ReplaceAll(tagVersion, "\n", "")
	if output, err = cmdOutput("git", "log", "-1", `--pretty=format:%H`); err != nil {
		err = fmt.Errorf("use git to get version fail. error: %w", err)
		return
	}
	gitVersion := output

	//now := time.Now().Format("2006-01-02 15:04:05Z07:00")

	if output, err = cmdOutput("go", "version"); err != nil {
		err = fmt.Errorf("use git to get version fail. error: %w", err)
		return
	}
	goVersion := strings.ReplaceAll(output, "\r", "")
	goVersion = strings.ReplaceAll(goVersion, "\n", "")
	version = fmt.Sprintf(versionCode, tagVersion, gitVersion, goVersion)
	return
}

func cmdOutput(cmd string, arg ...string) (output string, err error) {
	c := exec.Command(cmd, arg...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	if err = c.Run(); err != nil {
		err = fmt.Errorf("%v(%s)", err, stderr.String())
		return
	}
	return stdout.String(), nil
}
