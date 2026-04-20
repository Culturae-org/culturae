// backend/internal/model/manifest.go

package model

type DatasetManifest struct {
	SchemaVersion string            `json:"schema_version"`
	Dataset       string            `json:"dataset"`
	Version       string            `json:"version"`
	Type          string            `json:"type,omitempty"`
	License       string            `json:"license,omitempty"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
	Includes      []string          `json:"includes"`
	Sources       []ManifestSource  `json:"sources,omitempty"`
	Assets        map[string]string `json:"assets,omitempty"`
	Counts        ManifestCounts    `json:"counts"`
	Checksums     map[string]string `json:"checksums"`
}

type ManifestSource struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	License string `json:"license"`
}

type ManifestCounts struct {
	Questions int `json:"questions,omitempty"`
	Subthemes int `json:"subthemes,omitempty"`
	Tags      int `json:"tags,omitempty"`
	Themes    int `json:"themes,omitempty"`

	Countries  int `json:"countries,omitempty"`
	Continents int `json:"continents,omitempty"`
	Regions    int `json:"regions,omitempty"`
	Flags      int `json:"flags,omitempty"`
}

func (m *DatasetManifest) IsGeographyDataset() bool {
	return m.Type == "geography" || m.Dataset == "geography"
}

func (m *DatasetManifest) IsQuestionDataset() bool {
	return m.Type == "questions" || m.Dataset == "questions"
}

type ImportResult struct {
	Success          bool     `json:"success"`
	Message          string   `json:"message"`
	QuestionsAdded   int      `json:"questions_added"`
	QuestionsUpdated int      `json:"questions_updated"`
	QuestionsSkipped int      `json:"questions_skipped"`
	ThemesAdded      int      `json:"themes_added"`
	SubthemesAdded   int      `json:"subthemes_added"`
	TagsAdded        int      `json:"tags_added"`
	Errors           []string `json:"errors,omitempty"`
}
