package project

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	jsoniter "github.com/json-iterator/go"
	"github.com/karrick/godirwalk"
	"gopkg.in/yaml.v3"
	"tomasweigenast.com/nexema/tool/internal/analyzer"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/linker"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/schema"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

var (
	ErrEmptyParseTree error = errors.New("parsetree empty")
)

// ProjectBuilder builds a Nexema project.
//
// The steps to build a project are the following:
// 1. Finds in the working directory the nexema.yaml file.
// 2. Walk the entire input directory and its subdirectories, relative to the nexema.yaml file.
// 3. Parse every .nex file.
// 4. Link files with imports
// 5. Analyzes the files
// 6. Constructs a Nexema snapshot
// 7. Decide what to do next.
type ProjectBuilder struct {
	inputPath string

	config *ProjectConfig

	parseTree     *parser.ParseTree
	linkedScopes  []*scope.Scope
	builtSnapshot *definition.NexemaSnapshot
	parseErrs     parser.ParserErrorCollection
}

const builderVersion = 1
const nexExtension = ".nex"
const nexemaSnapshotExtension = ".nexs"

func NewProjectBuilder(inputPath string) *ProjectBuilder {
	return &ProjectBuilder{
		inputPath: inputPath,
		parseTree: parser.NewParseTree(),
	}
}

// Discover looks up for a nexema.yaml file in the input path
func (self *ProjectBuilder) Discover() error {
	buf, err := os.ReadFile(filepath.Join(self.inputPath, "nexema.yaml"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("not in a Nexema project directory")
		}
		return fmt.Errorf("nexema.yaml could not be read. Error: %v", err)
	}

	self.config = &ProjectConfig{}
	err = yaml.Unmarshal(buf, &self.config)
	if err != nil {
		return fmt.Errorf("invalid nexema.yaml file. Error: %v", err)
	}

	if self.config.Version != builderVersion {
		return fmt.Errorf("unknown Nexema builder version %d", self.config.Version)
	}

	if len(self.config.Generators) == 0 {
		return fmt.Errorf("you must specify at least one generator")
	}

	return nil
}

// Build builds the current discovered Nexema project. It will fail if Discover was not run or failed.
func (self *ProjectBuilder) Build() error {
	if self.config == nil {
		return fmt.Errorf("project was not discovered or Discover failed")
	}

	err := godirwalk.Walk(self.inputPath, &godirwalk.Options{
		Callback: func(osPathname string, directoryEntry *godirwalk.Dirent) error {
			if directoryEntry.IsDir() {
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
	if len(self.parseErrs) > 0 {
		return self.parseErrs.AsError()
	}

	if self.parseTree.IsEmpty() {
		return ErrEmptyParseTree
	}

	// link everything
	linker := linker.NewLinker(self.parseTree)
	linker.Link()

	if linker.HasLinkErrors() {
		return linker.Errors().AsError()
	}

	self.linkedScopes = linker.LinkedScopes()

	// run analyzer
	analyzer := analyzer.NewAnalyzer(self.linkedScopes)
	analyzer.Analyze()

	if analyzer.HasAnalysisErrors() {
		return analyzer.Errors().AsError()
	}

	return nil
}

// GetConfig returns the discovered project config
func (self *ProjectBuilder) GetConfig() *ProjectConfig {
	return self.config
}

// HasSnapshot returns true if builder built something
func (self *ProjectBuilder) HasSnapshot() bool {
	return self.builtSnapshot != nil && self.builtSnapshot.Hashcode != "0"
}

// BuildSnapshot creates a Nexema snapshot of a built project that can be saved to a file
func (self *ProjectBuilder) BuildSnapshot() error {
	schemaBuilder := schema.NewSchemaBuilder(self.linkedScopes)
	self.builtSnapshot = schemaBuilder.BuildSnapshot()

	return nil
}

// SaveSnapshot saves a built snapshot to a folder.
// If succeed, returns the full path to the file.
func (self *ProjectBuilder) SaveSnapshot(outputFolder string) (filename string, err error) {
	if self.builtSnapshot == nil {
		return "", errors.New("no snapshot has been built")
	}

	outPath := filepath.Join(outputFolder, fmt.Sprintf("%s%s", self.builtSnapshot.Hashcode, nexemaSnapshotExtension))
	err = os.MkdirAll(outputFolder, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("could not create output directory for the snapshot. Error: %v", err)
	}

	// todo: maybe serialize using nexemab
	buf, err := jsoniter.Marshal(self.builtSnapshot)
	if err != nil {
		return "", fmt.Errorf("could not serialize the snapshot. Error: %v", err)
	}

	err = os.WriteFile(outPath, buf, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("could not save Nexema snapshot to %s. Error: %v", outPath, err)
	}

	return outPath, nil
}

// GetSnapshot returns a previously built snapshot. Returns nil of no snapshot has been built.
func (self *ProjectBuilder) GetSnapshot() *definition.NexemaSnapshot {
	return self.builtSnapshot
}

// Format formats a project based on the currently built snapshot
func (self *ProjectBuilder) Format() error {
	context := FormatterContext{parseTree: self.parseTree}

	self.parseTree.Root().Iter(func(pkgName string, node *parser.ParseNode) {
		context.formatNode(pkgName, node)
	})

	return nil
}

func (self *ProjectBuilder) parseFile(p string) error {
	fileContents, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("could not parse file %s. Error: %v", p, err)
	}

	// todo: may reuse parser
	packagePath, _ := filepath.Rel(self.inputPath, path.Dir(p))
	filePath, _ := filepath.Rel(self.inputPath, p)
	if packagePath == "." {
		packagePath = "root"
	}

	parser := parser.NewParser(bytes.NewBuffer(fileContents), &parser.File{Path: filePath})
	parser.Reset()

	ast := parser.Parse()
	self.parseTree.Insert(packagePath, ast)
	self.parseErrs = append(self.parseErrs, *parser.Errors()...)

	return nil
}
