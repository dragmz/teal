package sarif

type Results struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []Run  `json:"runs"`
}

type Run struct {
	Tool      Tool       `json:"tool"`
	Artifacts []Artifact `json:"artifacts"`
	Results   []Result   `json:"results"`
}

type Result struct {
	Level     string           `json:"level"`
	Message   Message          `json:"message"`
	Locations []ResultLocation `json:"locations"`
	RuleId    string           `json:"ruleId"`
	RuleIndex int              `json:"ruleIndex"`
}

type ResultLocation struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

type ArtifactLocation struct {
	Uri   string `json:"uri"`
	Index int    `json:"index"`
}

type Region struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
}

type Message struct {
	Text string `json:"text"`
}

type Artifact struct {
	Location Location `json:"location"`
}

type Location struct {
	Uri string `json:"uri"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name           string `json:"name"`
	InformationUri string `json:"informationUri"`
	Rules          []Rule `json:"rules"`
}

type Rule struct {
	Id               string      `json:"id"`
	ShortDescription Description `json:"shortDescription"`
	HelpUri          string      `json:"helpUri"`
	Properties       []Property
}

type Description struct {
	Text string `json:"text"`
}

type Property struct {
	Category string `json:"category"`
}
