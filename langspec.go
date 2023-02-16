package teal

import (
	_ "embed"
	"encoding/json"
	"io"
	"strings"
)

//go:embed specs/langspec_v1.json
var BuiltInLangSpecV1 string

//go:embed specs/langspec_v2.json
var BuiltInLangSpecV2 string

//go:embed specs/langspec_v3.json
var BuiltInLangSpecV3 string

//go:embed specs/langspec_v4.json
var BuiltInLangSpecV4 string

//go:embed specs/langspec_v5.json
var BuiltInLangSpecV5 string

//go:embed specs/langspec_v6.json
var BuiltInLangSpecV6 string

//go:embed specs/langspec_v7.json
var BuiltInLangSpecV7 string

//go:embed specs/langspec_v8.json
var BuiltInLangSpecV8 string

//go:embed specs/langspec_v9.json
var BuiltInLangSpecV9 string

var BuiltInLangSpecs = []LangSpec{
	mustReadLangSpec(BuiltInLangSpecV1),
	mustReadLangSpec(BuiltInLangSpecV2),
	mustReadLangSpec(BuiltInLangSpecV3),
	mustReadLangSpec(BuiltInLangSpecV4),
	mustReadLangSpec(BuiltInLangSpecV5),
	mustReadLangSpec(BuiltInLangSpecV6),
	mustReadLangSpec(BuiltInLangSpecV7),
	mustReadLangSpec(BuiltInLangSpecV8),
	mustReadLangSpec(BuiltInLangSpecV9),
}

var LatestLangSpec = BuiltInLangSpecs[len(BuiltInLangSpecs)-1]

//go:embed langspec.json
var BuiltInLangSpecJson string

type LangSpecStackType struct {
	Type        string
	LengthBound [2]uint64
	ValueBound  [2]uint64
}

type LangSpecOp struct {
	Name     string
	Doc      string
	DocExtra string `json:",omitempty"`
}

type LangSpecFieldValue struct {
	Name string
	Type string
	Note string
}

type LangSpec struct {
	EvalMaxVersion  uint64
	LogicSigVersion uint64
	StackTypes      map[string]LangSpecStackType
	Ops             []LangSpecOp
	Fields          map[string][]LangSpecFieldValue
}

var BuiltInLangSpec = func() LangSpec {
	spec, err := readLangSpec(strings.NewReader(BuiltInLangSpecJson))
	if err != nil {
		panic(err)
	}

	return spec
}()

func readLangSpec(r io.Reader) (LangSpec, error) {
	var spec LangSpec

	d := json.NewDecoder(r)
	err := d.Decode(&spec)

	return spec, err
}

func mustReadLangSpec(specJson string) LangSpec {
	spec, err := readLangSpec(strings.NewReader(specJson))
	if err != nil {
		panic(err)
	}

	return spec
}

type immArgKind uint8

const (
	immNone = iota
	immField
	immArray
	immVar
)
