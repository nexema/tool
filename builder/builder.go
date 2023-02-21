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
	"tomasweigenast.com/nexema/tool/parser"
)

const nexExtension = ".nex"
const builderVersion = 1

// Builder is responsible of parsing, linking and analysing a Nexema project
type Builder struct {
	inputPath string // the path to the input folder

	config   *NexemaConfig
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

// Build builds a Nexema snapshot. It does not generates files
func (self *Builder) Build() error {
	err := self.scanProject()
	if err != nil {
		return err
	}

	// now, start walking directories
	err = godirwalk.Walk(self.inputPath, &godirwalk.Options{
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
	files := analyzer.Files()
	snapshotHashcode, err := hashstructure.Hash(&files, hashstructure.FormatV2, &hashstructure.HashOptions{})
	if err != nil {
		return err
	}

	self.snapshot = &definition.NexemaSnapshot{
		Version:  builderVersion,
		Files:    files,
		Hashcode: snapshotHashcode,
	}

	return nil
}

// Snapshot generates and saves a snapshot file for a previously built definition.
// This method must be called after self.Build
func (self *Builder) Snapshot(outFolder string) error {
	if self.snapshot == nil {
		return errors.New("definition not build")
	}

	outPath := filepath.Join(outFolder, fmt.Sprintf("%d.nexs", self.snapshot.Hashcode))

	err := os.Mkdir(filepath.Dir(outPath), os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("could not save snapshot. %s", err.Error())
	}

	// todo: serialize to binary using nexemab (nexema binary)
	buf, err := jsoniter.Marshal(self.snapshot)
	if err != nil {
		return err
	}

	err = os.WriteFile(outPath, buf, os.ModePerm)

	if err != nil {
		return fmt.Errorf("could not save Nexema snapshot. %s", err.Error())
	}

	return nil
}

// scanProject looks up at self.inputPath for a nexema.yaml file
func (self *Builder) scanProject() error {
	buf, err := os.ReadFile(filepath.Join(self.inputPath, "nexema.yaml"))
	if err != nil {
		return fmt.Errorf("nexema.yaml could not be read. Error: %s", err.Error())
	}

	err = yaml.Unmarshal(buf, &self.config)
	if err != nil {
		return fmt.Errorf("invalid nexema.yaml file. Error: %s", err.Error())
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
	parser := parser.NewParser(bytes.NewBuffer(fileContents), &parser.File{
		FileName: path.Base(p),
		Path:     packagePath,
	})

	// parse
	ast := parser.Parse()
	self.parseTree.Insert(packagePath, ast)
	self.parserErrors = append(self.parserErrors, *parser.Errors()...)

	return nil
}
