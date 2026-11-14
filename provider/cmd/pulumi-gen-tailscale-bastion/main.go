package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	dotnetgen "github.com/pulumi/pulumi-dotnet/pulumi-language-dotnet/v3/codegen"
	gogen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	nodejsgen "github.com/pulumi/pulumi/pkg/v3/codegen/nodejs"
	pygen "github.com/pulumi/pulumi/pkg/v3/codegen/python"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	pkghost "github.com/pulumi/pulumi/pkg/v3/host"
	"github.com/pulumi/pulumi/pkg/v3/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag/colors"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s <language> <out-dir> <schema-file>\n", os.Args[0])
		os.Exit(1)
	}

	language, outdir, schemaPath := os.Args[1], os.Args[2], os.Args[3]

	err := emitSDK(language, outdir, schemaPath)
	if err != nil {
		fmt.Printf("Failed: %s", err.Error())
	}
}

func emitSDK(language, outdir, schemaPath string) error {
	pkg, err := readSchema(schemaPath)
	if err != nil {
		return err
	}

	tool := "Pulumi SDK Generator"
	extraFiles := map[string][]byte{}

	var generator func() (map[string][]byte, error)
	switch language {
	case "dotnet":
		generator = func() (map[string][]byte, error) {
			return dotnetgen.GeneratePackage(tool, pkg, extraFiles, map[string]string{})
		}
	case "go":
		generator = func() (map[string][]byte, error) { return gogen.GeneratePackage(tool, pkg, map[string]string{}) }
	case "nodejs":
		generator = func() (map[string][]byte, error) {
			return nodejsgen.GeneratePackage(tool, pkg, extraFiles, map[string]string{}, true, nil)
		}
	case "python":
		generator = func() (map[string][]byte, error) { return pygen.GeneratePackage(tool, pkg, extraFiles, nil) }
	default:
		return errors.Errorf("Unrecognized language %q", language)
	}

	files, err := generator()
	if err != nil {
		return errors.Wrapf(err, "generating %s package", language)
	}
	if language == "nodejs" {
		if err := postProcessNodeJSFiles(files); err != nil {
			return err
		}
	}
	if language == "python" {
		fixPythonImportPaths(files)
	}
	if language == "dotnet" {
		for filename, contents := range files {
			if strings.HasSuffix(filename, ".csproj") {
				files[filename] = []byte(strings.ReplaceAll(string(contents),
					`Include="Microsoft.SourceLink.GitHub" Version="1.0.0"`,
					`Include="Microsoft.SourceLink.GitHub" Version="8.0.0"`))
			}
		}
	}

	for f, contents := range files {
		if err := emitFile(outdir, f, contents); err != nil {
			return errors.Wrapf(err, "emitting file %v", f)
		}
	}

	return nil
}

func postProcessNodeJSFiles(files map[string][]byte) error {
	for filename, contents := range files {
		switch {
		case filename == "package.json":
			packageJSON, err := addNodePackageManager(contents)
			if err != nil {
				return errors.Wrap(err, "adding packageManager to package.json")
			}
			files[filename] = packageJSON
		case strings.HasSuffix(filename, ".ts") && strings.Contains(filename, "/"):
			files[filename] = fixNodeJSImportPaths(contents)
		}
	}

	return nil
}

func addNodePackageManager(contents []byte) ([]byte, error) {
	var packageJSON map[string]interface{}
	if err := json.Unmarshal(contents, &packageJSON); err != nil {
		return nil, err
	}
	if _, ok := packageJSON["packageManager"]; !ok {
		packageJSON["packageManager"] = "yarn@1.22.22"
	}
	dependencies, ok := packageJSON["dependencies"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("generated Node.js package has no dependencies")
	}
	dependencies["@pulumi/pulumi"] = "^3.276.0"
	dependencies["@types/node"] = "^24"
	dependencies["typescript"] = "^6.0.3"
	packageJSON["repository"] = map[string]interface{}{
		"type": "git",
		"url":  "git+https://github.com/lbrlabs/pulumi-tailscale-bastion.git",
	}

	updated, err := json.MarshalIndent(packageJSON, "", "    ")
	if err != nil {
		return nil, err
	}
	return append(updated, '\n'), nil
}

func fixNodeJSImportPaths(contents []byte) []byte {
	updated := strings.ReplaceAll(string(contents), "from \"./types/", "from \"../types/")
	updated = strings.ReplaceAll(updated, "from \"./utilities\"", "from \"../utilities\"")
	return []byte(updated)
}

func fixPythonImportPaths(files map[string][]byte) {
	for filename, contents := range files {
		if !strings.HasSuffix(filename, ".py") || strings.Count(filename, "/") < 2 {
			continue
		}

		files[filename] = []byte(strings.ReplaceAll(string(contents), "from . import _utilities", "from .. import _utilities"))
	}
}

func readSchema(schemaPath string) (*schema.Package, error) {
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading schema")
	}

	if strings.HasSuffix(schemaPath, ".yaml") {
		schemaBytes, err = yaml.YAMLToJSON(schemaBytes)
		if err != nil {
			return nil, errors.Wrap(err, "reading YAML schema")
		}
	}

	var spec schema.PackageSpec
	if err = json.Unmarshal(schemaBytes, &spec); err != nil {
		return nil, errors.Wrap(err, "unmarshalling schema")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "getting working directory")
	}
	sink := diag.DefaultSink(os.Stdout, os.Stderr, diag.FormatOptions{Color: colors.Never})
	host, err := pkghost.New(context.Background(), sink, sink, nil, nil, schema.NewLoaderServerFromContext, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating plugin host")
	}
	defer host.Close()
	pctx, err := plugin.NewContext(context.Background(), sink, sink, host, nil, cwd, nil, false, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating plugin context")
	}
	defer pctx.Close()
	pkg, err := schema.ImportSpec(spec, nil, schema.NewPluginLoader(pctx), schema.ValidationOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "importing schema")
	}
	return pkg, nil
}

func emitFile(rootDir, filename string, contents []byte) error {
	outPath := filepath.Join(rootDir, filename)
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, contents, 0600); err != nil {
		return err
	}
	return nil
}
