package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/zricethezav/gitleaks/v8/config"
)

func writeSarif(cfg config.Config, findings []*Finding, w io.WriteCloser) error {
	sarif := Sarif{
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json",
		Version: version,
		Runs:    getRuns(cfg, findings),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", " ")
	return encoder.Encode(sarif)
}

func getRuns(cfg config.Config, findings []*Finding) []Runs {
	return []Runs{
		{
			Tool:    getTool(cfg),
			Results: getResults(findings),
		},
	}
}

func getTool(cfg config.Config) Tool {
	return Tool{
		Driver: Driver{
			Name:            driver,
			SemanticVersion: version,
			Rules:           getRules(cfg),
		},
	}
}

func getRules(cfg config.Config) []Rules {
	// TODO	for _, rule := range cfg.Rules {
	var rules []Rules
	for _, rule := range cfg.Rules {
		rules = append(rules, Rules{
			ID:   rule.RuleID,
			Name: rule.Description,
			Description: ShortDescription{
				Text: rule.Regex.String(),
			},
		})
	}
	return rules
}

func messageText(f *Finding) string {
	if f.Commit == "" {
		return fmt.Sprintf("%s has detected secret for file %s.", f.RuleID, f.File)
	}

	return fmt.Sprintf("%s has detected secret for file %s at commit %s.", f.RuleID, f.File, f.Commit)

}

func getResults(findings []*Finding) []Results {
	var results []Results
	for _, f := range findings {
		r := Results{
			Message: Message{
				Text: messageText(f),
			},
			RuleId:    f.RuleID,
			Locations: getLocation(f),
			// This information goes in partial fingerprings until revision
			// data can be added somewhere else
			PartialFingerPrints: PartialFingerPrints{
				CommitSha:     f.Commit,
				Email:         f.Email,
				CommitMessage: f.Message,
				Date:          f.Date,
				Author:        f.Author,
			},
		}
		results = append(results, r)
	}
	return results
}

func getLocation(f *Finding) []Locations {
	return []Locations{
		{
			PhysicalLocation: PhysicalLocation{
				ArtifactLocation: ArtifactLocation{
					URI: f.File,
				},
				Region: Region{
					StartLine:   f.StartLine,
					EndLine:     f.EndLine,
					StartColumn: f.StartColumn,
					EndColumn:   f.EndColumn,
					Snippet: Snippet{
						Text: f.Secret,
					},
				},
			},
		},
	}
}

type PartialFingerPrints struct {
	CommitSha     string `json:"commitSha"`
	Email         string `json:"email"`
	Author        string `json:"author"`
	Date          string `json:"date"`
	CommitMessage string `json:"commitMessage"`
}

type Sarif struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Runs `json:"runs"`
}

type ShortDescription struct {
	Text string `json:"text"`
}

type FullDescription struct {
	Text string `json:"text"`
}

type Rules struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description ShortDescription `json:"shortDescription"`
}

type Driver struct {
	Name            string  `json:"name"`
	SemanticVersion string  `json:"semanticVersion"`
	Rules           []Rules `json:"rules"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Message struct {
	Text string `json:"text"`
}

type ArtifactLocation struct {
	URI string `json:"uri"`
}

type Region struct {
	StartLine   int     `json:"startLine"`
	StartColumn int     `json:"startColumn"`
	EndLine     int     `json:"endLine"`
	EndColumn   int     `json:"endColumn"`
	Snippet     Snippet `json:"snippet"`
}

type Snippet struct {
	Text string `json:"text"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

type Locations struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type Results struct {
	Message             Message     `json:"message"`
	RuleId              string      `json:"ruleId"`
	Locations           []Locations `json:"locations"`
	PartialFingerPrints `json:"partialFingerprints"`
}

type Runs struct {
	Tool    Tool      `json:"tool"`
	Results []Results `json:"results"`
}

func configToRules(cfg config.Config) []Rules {
	var rules []Rules
	for _, rule := range cfg.Rules {
		rules = append(rules, Rules{
			ID:   rule.RuleID,
			Name: rule.Description,
		})
	}
	return rules
}