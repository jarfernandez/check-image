package commands

import (
	"fmt"
	"os"
	"sort"

	"github.com/jarfernandez/check-image/internal/output"
)

// renderResult renders a CheckResult according to the current OutputFmt.
// In text mode, it calls the appropriate text renderer.
// In JSON mode, it writes JSON to stdout.
func renderResult(r *output.CheckResult) error {
	if OutputFmt == output.FormatJSON {
		return output.RenderJSON(os.Stdout, r)
	}

	switch r.Check {
	case "age":
		renderAgeText(r)
	case "size":
		renderSizeText(r)
	case "ports":
		renderPortsText(r)
	case "registry":
		renderRegistryText(r)
	case "root-user":
		renderRootUserText(r)
	case "secrets":
		renderSecretsText(r)
	case "healthcheck":
		renderHealthcheckText(r)
	case "labels":
		renderLabelsText(r)
	case "entrypoint":
		renderEntrypointText(r)
	}

	return nil
}

func renderAgeText(r *output.CheckResult) {
	d := r.Details.(output.AgeDetails)
	fmt.Printf("Checking age of image %s\n", r.Image)
	fmt.Printf("Image creation date: %s\n", d.CreatedAt)
	fmt.Printf("Image age: %.0f days\n", d.AgeDays)
	fmt.Println(r.Message)
}

func renderSizeText(r *output.CheckResult) {
	d := r.Details.(output.SizeDetails)
	fmt.Printf("Checking size and layers of image %s\n", r.Image)
	fmt.Printf("Number of layers: %d\n", d.LayerCount)
	// #nosec G115 -- LayerCount is always non-negative (derived from layer enumeration)
	if uint(d.LayerCount) > d.MaxLayers {
		fmt.Printf("Image has more than %d layers\n", d.MaxLayers)
	}
	for _, l := range d.Layers {
		fmt.Printf("  Layer %d: %d bytes\n", l.Index, l.Bytes)
	}
	fmt.Printf("Total size: %d bytes (%.2f MB)\n", d.TotalBytes, d.TotalMB)
	fmt.Println(r.Message)
}

func renderPortsText(r *output.CheckResult) {
	d := r.Details.(output.PortsDetails)
	fmt.Printf("Checking ports of image %s\n", r.Image)

	if len(d.ExposedPorts) == 0 {
		fmt.Println("No ports are exposed in this image")
		return
	}

	fmt.Println("Exposed ports:")
	for _, port := range d.ExposedPorts {
		fmt.Printf("  - %d\n", port)
	}

	if len(d.AllowedPorts) == 0 {
		fmt.Println("No allowed ports were provided")
		return
	}

	if len(d.UnauthorizedPorts) > 0 {
		fmt.Println("The following ports are not in the allowed list:")
		for _, port := range d.UnauthorizedPorts {
			fmt.Printf("  - %d\n", port)
		}
	}

	if r.Message != "" {
		fmt.Println(r.Message)
	}
}

func renderRegistryText(r *output.CheckResult) {
	d := r.Details.(output.RegistryDetails)
	fmt.Printf("Checking registry of image %s\n", r.Image)

	if d.Skipped {
		fmt.Println("Registry validation skipped (not applicable for this transport)")
		return
	}

	fmt.Printf("Image registry: %s\n", d.Registry)
	fmt.Println(r.Message)
}

func renderRootUserText(r *output.CheckResult) {
	fmt.Printf("Checking if image %s is configured to run as a non-root user\n", r.Image)
	fmt.Println(r.Message)
}

func renderSecretsText(r *output.CheckResult) {
	d := r.Details.(output.SecretsDetails)
	fmt.Printf("Checking secrets in image %s\n", r.Image)

	if len(d.EnvVarFindings) > 0 {
		fmt.Println("\nEnvironment Variables:")
		for _, finding := range d.EnvVarFindings {
			fmt.Printf("  - %s (%s)\n", finding.Name, finding.Description)
		}
	}

	if len(d.FileFindings) > 0 {
		fmt.Println("\nFiles with Sensitive Patterns:")

		// Group findings by layer for better readability
		layerMap := make(map[int][]output.FileFinding)
		for _, finding := range d.FileFindings {
			layerMap[finding.LayerIndex] = append(layerMap[finding.LayerIndex], finding)
		}

		for layerIdx := 0; layerIdx < len(layerMap)+10; layerIdx++ {
			if findings, ok := layerMap[layerIdx]; ok {
				fmt.Printf("  Layer %d:\n", layerIdx+1)
				for _, finding := range findings {
					fmt.Printf("    - %s (%s)\n", finding.Path, finding.Description)
				}
			}
		}
	}

	fmt.Printf("\nTotal findings: %d", d.TotalFindings)
	if d.EnvVarCount >= 0 && d.FileCount >= 0 && (d.EnvVarCount > 0 || d.FileCount > 0 || d.TotalFindings > 0) {
		fmt.Printf(" (%d environment variables, %d files)\n", d.EnvVarCount, d.FileCount)
	} else {
		fmt.Println()
	}

	fmt.Println(r.Message)
}

func renderHealthcheckText(r *output.CheckResult) {
	fmt.Printf("Checking if image %s has a healthcheck defined\n", r.Image)
	fmt.Println(r.Message)
}

func renderEntrypointText(r *output.CheckResult) {
	d := r.Details.(output.EntrypointDetails)
	fmt.Printf("Checking entrypoint of image %s\n", r.Image)
	if len(d.Entrypoint) > 0 {
		fmt.Printf("Entrypoint: %v\n", d.Entrypoint)
	}
	if len(d.Cmd) > 0 {
		fmt.Printf("Cmd: %v\n", d.Cmd)
	}
	fmt.Println(r.Message)
}

func renderLabelsText(r *output.CheckResult) {
	d := r.Details.(output.LabelsDetails)
	fmt.Printf("Checking labels of image %s\n", r.Image)

	// Show required labels
	if len(d.RequiredLabels) > 0 {
		fmt.Println("\nRequired labels:")
		for _, req := range d.RequiredLabels {
			switch {
			case req.Pattern != "":
				fmt.Printf("  - %s (pattern: %q)\n", req.Name, req.Pattern)
			case req.Value != "":
				fmt.Printf("  - %s (exact: %q)\n", req.Name, req.Value)
			default:
				fmt.Printf("  - %s (existence check)\n", req.Name)
			}
		}
	}

	// Show actual labels from image
	if len(d.ActualLabels) > 0 {
		fmt.Printf("\nActual labels (%d):\n", len(d.ActualLabels))
		// Sort keys for deterministic output
		keys := make([]string, 0, len(d.ActualLabels))
		for k := range d.ActualLabels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %s: %s\n", k, d.ActualLabels[k])
		}
	} else {
		fmt.Println("\nNo labels found in image")
	}

	// Show missing labels
	if len(d.MissingLabels) > 0 {
		fmt.Printf("\nMissing labels (%d):\n", len(d.MissingLabels))
		for _, name := range d.MissingLabels {
			fmt.Printf("  - %s\n", name)
		}
	}

	// Show invalid labels
	if len(d.InvalidLabels) > 0 {
		fmt.Printf("\nInvalid labels (%d):\n", len(d.InvalidLabels))
		for _, inv := range d.InvalidLabels {
			fmt.Printf("  - %s: %s\n", inv.Name, inv.Reason)
		}
	}

	fmt.Printf("\n%s\n", r.Message)
}
