// Command protoc-gen-orm is a protoc plugin that reads proto descriptors
// annotated with google.api.* and orm.v1.* options, then generates database
// schema artifacts for the requested backend: gorm, sql, or prisma. The generic
// IR engine lives in the protokit module; the on-chain (solidity/subgraph)
// targets ship as the separate protoc-gen-web3 plugin.
//
// # Install
//
//	go install github.com/the-protobuf-project/orm/plugin/cmd/protoc-gen-orm@latest
//
// # Usage via buf.gen.yaml
//
//	plugins:
//	  - local: protoc-gen-orm
//	    out: generated/
//	    opt:
//	      - target=prisma   # prisma | gorm | sql
//
// # Inference priority
//
//  1. google.api.* annotations — table, column, FK inference (~80%)
//  2. orm.v1.* options         — overrides: type, name, skip, unique, index
//  3. buf.gen.yaml opt:         — global defaults (target backend)
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/the-protobuf-project/orm/plugin/generator"
	"github.com/the-protobuf-project/orm/plugin/generator/backend"
	core "github.com/the-protobuf-project/protokit"
	"github.com/the-protobuf-project/protokit/header"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// version is the build version, injected at release time via
// -ldflags "-X main.version=...".
var version = "dev"

// resolveVersion returns the build version to stamp into generated files.
// A release sets `version` via ldflags and wins outright. Otherwise we recover
// it from the build info the Go toolchain embeds: `go install …@v0.1.2` records
// the tag as the main module version, and when orm is consumed as a
// dependency its module entry carries the version. Only genuine local builds
// (`go build`/`go run`, which report "(devel)") fall back to "dev".
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}
	if v := bi.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	for _, dep := range bi.Deps {
		if dep.Path == "github.com/the-protobuf-project/orm" && dep.Version != "" {
			return dep.Version
		}
	}
	return version
}

func main() {
	v := resolveVersion()
	// When invoked directly with -version (not by protoc), print and exit before
	// protogen tries to read a CodeGeneratorRequest from stdin.
	if len(os.Args) == 2 && (os.Args[1] == "-version" || os.Args[1] == "--version") {
		fmt.Printf("protoc-gen-orm %s\n", v)
		return
	}

	// Every generated file's banner names the tool that produced it.
	header.SetTool("protoc-gen-orm")

	// flags are populated by protogen (ParamFunc maps each buf.gen.yaml opt:
	// "key=value" to flags.Set) before the Run closure reads them.
	var flags flag.FlagSet
	target := flags.String("target", "", "output backend: gorm | sql | prisma")
	strict := flags.String("strict", "",
		"per-rule severity for schema problems: \"\"=all warn, \"true\"=all error, "+
			"or \"ref:error,collision:warn,index:error,lint:warn\"")
	config := flags.String("config", "",
		"path to an orm.yaml mapping proto packages to databases/schemas")
	goModule := flags.String("go_module", "",
		"Go import path of the output directory (e.g. github.com/me/gen); the gorm "+
			"target needs it to generate the migration aggregator that imports each schema package")
	stores := flags.Bool("stores", false,
		"gorm target only: also generate a typed CRUD store per resource "+
			"(introduces a gorm.io/gorm dependency in each models package)")
	converters := flags.Bool("converters", false,
		"gorm target only: also generate proto↔model converters per schema "+
			"(<Model>ToProto / <Model>FromProto plus per-enum value mappers; introduces "+
			"a dependency on the generated proto packages)")
	otel := flags.Bool("otel", true,
		"gorm target only: fold an OpenTelemetry tracing helper into the migration "+
			"Registry (Instrument); on by default, takes effect with go_module. "+
			"Set otel=false to omit it. orm.yaml otel: tunes it further")

	protogen.Options{ParamFunc: flags.Set}.Run(func(p *protogen.Plugin) error {
		// Proto3 `optional` is fully supported (presence is read via field_behavior,
		// not synthetic oneofs); declare it so buf/protoc don't warn.
		p.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		// orm owns its layout config (protokit has none): load it here and hand the
		// backend a fully-resolved view of grouping + the gorm render knobs.
		cfg, err := backend.LoadConfig(*config)
		if err != nil {
			return err
		}
		return core.Run(p, core.Options{
			Target:  *target,
			Strict:  *strict,
			Version: v,
		}, generator.Targets(), backend.New(cfg, *goModule, *stores, *otel, *converters))
	})
}
