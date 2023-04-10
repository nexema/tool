package builder

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	jsoniter "github.com/json-iterator/go"
	"github.com/karrick/godirwalk"
	"github.com/mitchellh/hashstructure/v2"
	"gopkg.in/yaml.v3"
	"tomasweigenast.com/nexema/tool/analyzer"
	"tomasweigenast.com/nexema/tool/definition"
	"tomasweigenast.com/nexema/tool/linker"
	"tomasweigenast.com/nexema/tool/nexema"
	"tomasweigenast.com/nexema/tool/parser"
)

const nexExtension = ".nex"
const builderVersion = 1

// Builder is responsible of parsing, linking and analysing a Nexema project
type Builder struct {
	inputPath string // the path to the input folder

	config   *nexema.NexemaProjectConfig
	snapshot *definition.NexemaSnapshot // the generated snapshot

	parserErrors parser.ParserErrorCollection
	parseTree    *parser.ParseTree
}

func NewBuilder(inputPath string) *Builder {
	return &Builder{
		inputPath: inputPath,
		parseTree: parser.NewParseTree(),
	}
}

// Config returns the discovered nexema.yaml file
func (self *Builder) Config() *nexema.NexemaProjectConfig {
	return self.config
}

// Discover looks up for a Nexema project in input directory
func (self *Builder) Discover() error {
	err := self.scanProject()
	if err != nil {
		return err
	}

	return nil
}

// Build builds a Nexema snapshot. It does not generates files
func (self *Builder) Build() error {
	if self.config == nil {
		panic("this must be called after Discover method")
	}

	// now, start walking directories
	err := godirwalk.Walk(self.inputPath, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				return nil // skip
			}

			if filepath.Ext(osPathname) != nexExtension {
				return godirwalk.SkipThis
			}

			// scan file
			return self.parseFile(osPathname)
		},
		Unsorted:            true,
		FollowSymbolicLinks: false,
		AllowNonDirectory:   false,
	})

	if err != nil {
		return err
	}

	// at this moment, if any parse error is encountered, return as error
	if len(self.parserErrors) > 0 {
		return self.parserErrors.AsError()
	}

	// link
	linker := linker.NewLinker(self.parseTree)
	linker.Link()

	if linker.HasLinkErrors() {
		return linker.Errors().AsError()
	}

	// run analyzer
	analyzer := analyzer.NewAnalyzer(linker.LinkedScopes())
	analyzer.Analyze()

	if analyzer.HasAnalysisErrors() {
		return analyzer.Errors().AsError()
	}

	// build snapshot
	files := make([]definition.NexemaFile, 0)
	ids := make([]string, 0)
	for _, file := range analyzer.Files() {
		files = append(files, file)
		ids = append(ids, file.Id)
	}

	snapshotHashcode, err := hashstructure.Hash(&ids, hashstructure.FormatV2, &hashstructure.HashOptions{})
	if err != nil {
		return err
	}

	self.snapshot = &definition.NexemaSnapshot{
		Version:  builderVersion,
		Files:    files,
		Hashcode: fmt.Sprint(snapshotHashcode),
	}

	return nil
}

// SaveSnapshot generates and saves a snapshot file for a previously built definition.
// This method must be called after self.Build
func (self *Builder) SaveSnapshot(outFolder string) (filename string, err error) {
	if self.snapshot == nil {
		return "", errors.New("definition not build")
	}

	outPath := filepath.Join(outFolder, fmt.Sprintf("%s.nexs", self.snapshot.Hashcode))

	err = os.Mkdir(outFolder, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf("could not save snapshot. %s", err.Error())
	}

	// todo: serialize to binary using nexemab (nexema binary)
	buf, err := jsoniter.Marshal(self.snapshot)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(outPath, buf, os.ModePerm)

	if err != nil {
		return "", fmt.Errorf("could not save Nexema snapshot. %s", err.Error())
	}

	return outPath, nil
}

// Snapshot returns the built NexemaSnapshot
func (self *Builder) Snapshot() *definition.NexemaSnapshot {
	return self.snapshot
}

// scanProject looks up at self.inputPath for a nexema.yaml file
func (self *Builder) scanProject() error {
	buf, err := os.ReadFile(filepath.Join(self.inputPath, "nexema.yaml"))
	if err != nil {
		return fmt.Errorf("nexema.yaml could not be read. Error: %s", err.Error())
	}

	self.config = &nexema.NexemaProjectConfig{}
	err = yaml.Unmarshal(buf, &self.config)
	if err != nil {
		return fmt.Errorf("invalid nexema.yaml file. Error: %s", err.Error())
	}

	if self.config.Version != builderVersion {
		return fmt.Errorf("invalid Nexema builder version %d", self.config.Version)
	}

	if len(self.config.Generators) == 0 {
		return fmt.Errorf("you must specify at least one generator in nexema.yaml")
	}

	return nil
}

// parseFile parses a file
func (self *Builder) parseFile(p string) error {
	fileContents, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("could not parse file %s. Error: %s", p, err)
	}

	// todo: maybe re-use the parser
	packagePath, _ := filepath.Rel(self.inputPath, path.Dir(p))
	if packagePath == "." {
		packagePath = "root"
	}

	parser := parser.NewParser(bytes.NewBuffer(fileContents), &parser.File{
		FileName: path.Base(p),
		Path:     packagePath,
	})

	parser.Begin()

	// parse
	ast := parser.Parse()
	self.parseTree.Insert(packagePath, ast)
	self.parserErrors = append(self.parserErrors, *parser.Errors()...)

	return nil
}
