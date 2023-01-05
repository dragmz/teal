package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/dragmz/teal"
	"github.com/dragmz/teal/internal/sarif"
	"github.com/pkg/errors"
)

type args struct {
	Path string
}

func ToSarifLevel(s teal.DiagnosticSeverity) string {
	switch s {
	case teal.DiagInfo:
		return "note"
	case teal.DiagHint:
		return "note"
	case teal.DiagWarn:
		return "warning"
	case teal.DiagErr:
		return "error"
	default:
		panic("unexpected severity")
	}
}

func run(a args) error {
	sr := sarif.Results{
		Version: "2.1.0",
		Schema:  "http://json.schemastore.org/sarif-2.1.0-rtm.4",
		Runs:    []sarif.Run{},
	}

	rules := []sarif.Rule{
		{
			Id: "SYNTAX",
			ShortDescription: sarif.Description{
				Text: "Syntax checks",
			},
		},
		{
			Id: "PARSE",
			ShortDescription: sarif.Description{
				Text: "Parser checks",
			},
		},
	}

	for _, r := range teal.LintRules {
		rules = append(rules, sarif.Rule{
			Id: r.Id(),
			ShortDescription: sarif.Description{
				Text: r.Desc(),
			},
		})
	}

	run := sarif.Run{
		Tool: sarif.Tool{
			Driver: sarif.Driver{
				Name:           "tealscan",
				InformationUri: "https://github.com/dragmz/teal",
				Rules:          rules,
			},
		},
		Artifacts: []sarif.Artifact{},
		Results:   []sarif.Result{},
	}

	var paths []string

	fi, err := os.Stat(a.Path)
	if err != nil {
		return errors.Wrap(err, "failed to read path")
	}

	if fi.IsDir() {
		err := filepath.WalkDir(a.Path, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				if filepath.Ext(path) == ".teal" {
					paths = append(paths, path)
				}
			}
			return nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to walk dir")
		}
	} else {
		paths = append(paths, a.Path)
	}

	for _, path := range paths {
		s, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		res := teal.Process(string(s))

		ab, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		u := url.URL{
			Scheme: "file",
			Path:   ab,
		}

		up := u.String()

		run.Artifacts = append(run.Artifacts, sarif.Artifact{
			Location: sarif.Location{
				Uri: up,
			},
		})

		for i, d := range res.Diagnostics {
			run.Results = append(run.Results, sarif.Result{
				RuleId: d.Rule(),
				Level:  ToSarifLevel(d.Severity()),
				Message: sarif.Message{
					Text: d.String(),
				},
				Locations: []sarif.ResultLocation{
					{
						PhysicalLocation: sarif.PhysicalLocation{
							ArtifactLocation: sarif.ArtifactLocation{
								Uri:   up,
								Index: i,
							},
							Region: sarif.Region{
								StartLine:   d.Line() + 1,
								StartColumn: d.Begin() + 1,
							},
						},
					},
				},
			})
		}
	}

	sr.Runs = append(sr.Runs, run)

	rb, err := json.MarshalIndent(sr, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(rb))

	return err
}

func main() {
	var a args

	flag.StringVar(&a.Path, "path", "", "path to scan")
	flag.Parse()

	err := run(a)
	if err != nil {
		panic(err)
	}
}
