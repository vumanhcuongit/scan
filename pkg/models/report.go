package models

type FindingReport struct {
	Findings []Finding `json:"findings"`
}

type Finding struct {
	Type     string   `json:"type"`
	RuleID   string   `json:"ruleId"`
	Location Location `json:"location"`
	Metadata Metadata `json:"metadata"`
}

type Location struct {
	Path     string   `json:"path"`
	Position Position `json:"positions"`
}

type Position struct {
	Begin Begin `json:"begin"`
}

type Begin struct {
	Line int `json:"line"`
}

type Metadata struct {
	Description string `json:"description"`
	Severity    string `json:"severity"`
}
