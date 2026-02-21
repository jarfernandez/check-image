package output

// CheckResult is the common envelope for every validation check.
type CheckResult struct {
	Check   string `json:"check"`
	Image   string `json:"image"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AgeDetails holds details for the age check.
type AgeDetails struct {
	CreatedAt string  `json:"created-at"`
	AgeDays   float64 `json:"age-days"`
	MaxAge    uint    `json:"max-age"`
}

// SizeDetails holds details for the size check.
type SizeDetails struct {
	TotalBytes int64       `json:"total-bytes"`
	TotalMB    float64     `json:"total-mb"`
	MaxSizeMB  uint        `json:"max-size-mb"`
	LayerCount int         `json:"layer-count"`
	MaxLayers  uint        `json:"max-layers"`
	Layers     []LayerInfo `json:"layers"`
}

// LayerInfo holds size information for a single layer.
type LayerInfo struct {
	Index int   `json:"index"`
	Bytes int64 `json:"bytes"`
}

// PortsDetails holds details for the ports check.
type PortsDetails struct {
	ExposedPorts      []int `json:"exposed-ports"`
	AllowedPorts      []int `json:"allowed-ports,omitempty"`
	UnauthorizedPorts []int `json:"unauthorized-ports,omitempty"`
}

// RegistryDetails holds details for the registry check.
type RegistryDetails struct {
	Registry string `json:"registry"`
	Skipped  bool   `json:"skipped,omitempty"`
}

// RootUserDetails holds details for the root-user check.
type RootUserDetails struct {
	User string `json:"user"`
}

// SecretsDetails holds details for the secrets check.
type SecretsDetails struct {
	EnvVarFindings []EnvVarFinding `json:"env-var-findings,omitempty"`
	FileFindings   []FileFinding   `json:"file-findings,omitempty"`
	TotalFindings  int             `json:"total-findings"`
	EnvVarCount    int             `json:"env-var-count"`
	FileCount      int             `json:"file-count"`
}

// EnvVarFinding represents a sensitive environment variable finding.
type EnvVarFinding struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FileFinding represents a sensitive file finding.
type FileFinding struct {
	Path        string `json:"path"`
	LayerIndex  int    `json:"layer-index"`
	Description string `json:"description"`
}

// HealthcheckDetails holds details for the healthcheck check.
type HealthcheckDetails struct {
	HasHealthcheck bool `json:"has-healthcheck"`
}

// LabelsDetails holds details for the labels check.
type LabelsDetails struct {
	RequiredLabels []RequiredLabelCheck `json:"required-labels"`
	ActualLabels   map[string]string    `json:"actual-labels,omitempty"`
	MissingLabels  []string             `json:"missing-labels,omitempty"`
	InvalidLabels  []InvalidLabelDetail `json:"invalid-labels,omitempty"`
}

// RequiredLabelCheck represents a required label with its validation mode.
type RequiredLabelCheck struct {
	Name    string `json:"name"`
	Value   string `json:"value,omitempty"`
	Pattern string `json:"pattern,omitempty"`
}

// InvalidLabelDetail represents a label that exists but doesn't meet requirements.
type InvalidLabelDetail struct {
	Name            string `json:"name"`
	ActualValue     string `json:"actual-value"`
	ExpectedValue   string `json:"expected-value,omitempty"`
	ExpectedPattern string `json:"expected-pattern,omitempty"`
	Reason          string `json:"reason"`
}

// PlatformDetails holds details for the platform check.
type PlatformDetails struct {
	Platform         string   `json:"platform"`
	AllowedPlatforms []string `json:"allowed-platforms"`
}

// EntrypointDetails holds details for the entrypoint check.
type EntrypointDetails struct {
	HasEntrypoint    bool     `json:"has-entrypoint"`
	ExecForm         bool     `json:"exec-form,omitempty"`
	ShellFormAllowed bool     `json:"shell-form-allowed,omitempty"`
	Entrypoint       []string `json:"entrypoint,omitempty"`
	Cmd              []string `json:"cmd,omitempty"`
}

// AllResult is the aggregated result for the "all" command.
type AllResult struct {
	Image   string        `json:"image"`
	Passed  bool          `json:"passed"`
	Checks  []CheckResult `json:"checks"`
	Summary Summary       `json:"summary"`
}

// Summary holds counts for the "all" command.
type Summary struct {
	Total   int      `json:"total"`
	Passed  int      `json:"passed"`
	Failed  int      `json:"failed"`
	Errored int      `json:"errored"`
	Skipped []string `json:"skipped,omitempty"`
}

// VersionResult holds the short version output for JSON mode (--short flag).
type VersionResult struct {
	Version string `json:"version"`
}

// BuildInfoResult holds the full build information for JSON mode.
type BuildInfoResult struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuiltAt   string `json:"built-at"`
	GoVersion string `json:"go-version"`
	Platform  string `json:"platform"`
}
