package main

import (
	"bytes"
	"flag"
	"fmt"
	"go-example/hot_switch/utils"
	"go/parser"
	"go/token"
	"golang.org/x/mod/modfile"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const DateTimeFORMAt = "20060102150405"

var pluginDir, outDir string

func init() {
	flag.StringVar(&pluginDir, "pluginDir", "", "The plugin module directory")
	flag.StringVar(&outDir, "outDir", "", "The plugin .so output directory")
}

func main() {
	flag.Parse()
	if err := utils.IsDirectory(pluginDir, "pluginDir"); err != nil {
		panic(err)
	}
	pluginDir, _ = filepath.Abs(pluginDir)

	if pkgs, err := parser.ParseDir(token.NewFileSet(), pluginDir, nil, parser.PackageClauseOnly); err != nil {
		panic(err)
	} else {
		if len(pkgs) != 1 {
			panic(fmt.Errorf("%s contains %d packages", pluginDir, len(pkgs)))
		}
		for _, pkg := range pkgs {
			if pkg.Name != "main" {
				panic(fmt.Errorf("%s not contains main package", pluginDir))
			}
		}
	}

	var a []string
	modPath := pluginDir
	for modPath != "" {
		f := filepath.Join(modPath, "go.mod")
		if _, err := os.Stat(f); err != nil {
			a = append(a, filepath.Base(modPath))
			modPath = filepath.Dir(modPath)
		} else {
			break
		}
	}

	modules := make([]string, len(a)+1)
	if data, err := os.ReadFile(filepath.Join(modPath, "go.mod")); err != nil {
		panic(err)
	} else {
		mod, err := modfile.ParseLax("go.mod", data, nil)
		if err != nil {
			panic(err)
		}
		if mod.Module == nil {
			panic("failed to parse go.mod [1]")
		}
		if mod.Module.Mod.Path == "" {
			panic("failed to parse go.mod [2]")
		}
		modules[0] = mod.Module.Mod.Path
	}

	for i, j := len(a)-1, 1; i > 0; i, j = i-1, j+1 {
		modules[j] = a[i]
	}
	modulePath := filepath.Join(modules...)

	tmpDir := fmt.Sprintf("%s-%s-%s", pluginDir, time.Now().Format(DateTimeFORMAt), utils.RandomString(6))
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	replaceModulePath := filepath.Join(filepath.Dir(modulePath), filepath.Base(tmpDir))

	var files []string
	if err := filepath.WalkDir(pluginDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if pluginDir == path {
			return err
		}
		rel, err := filepath.Rel(pluginDir, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(tmpDir, rel), 0744)
		}

		switch {
		case strings.HasSuffix(path, "_test.go"):
			return nil
		case strings.HasSuffix(path, ".go"):
		default:
			return nil
		}

		files = append(files, rel)
		return nil
	}); err != nil {
		panic(err)
	}

	if err := copyFiles(pluginDir, tmpDir, files, modulePath, replaceModulePath); err != nil {
		panic(err)
	}

	outDir, _ := filepath.Abs(outDir)
	_ = os.MkdirAll(outDir, 0744)
	outputFile := filepath.Join(outDir, fmt.Sprintf("%s.so", filepath.Base(pluginDir)))
	var args []string
	args = append(args, []string{"build", "-trimpath", "-buildmode=plugin", "-o", outputFile}...)

	goBuild := exec.Command("go", args...)
	goBuild.Dir = tmpDir
	goBuild.Stdout = os.Stdout
	goBuild.Stderr = os.Stderr
	goBuild.Env = append(os.Environ(), "GO111MODULE=on")
	if err := goBuild.Run(); err != nil {
		panic(err)
	}
}

func copyFiles(fromDir, toDir string, files []string, oldModulePath, newModulePath string) (err error) {
	oldImportPath := []byte(oldModulePath)
	newImportPath := []byte(newModulePath)
	for _, file := range files {
		fromF := filepath.Join(fromDir, file)
		toF := filepath.Join(toDir, file)
		data, e := os.ReadFile(fromF)
		if e != nil {
			err = e
			break
		}
		data = bytes.ReplaceAll(data, oldImportPath, newImportPath)
		err = os.WriteFile(toF, data, 0744)
	}
	return
}
