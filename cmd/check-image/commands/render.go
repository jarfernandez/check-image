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
	case "platform":
		renderPlatformText(r)
	}

	return nil
}

func renderAgeText(r *output.CheckResult) {
	d := r.Details.(output.AgeDetails)
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking age of image %s", r.Image)))
	fmt.Printf("Image creation date: %s\n", valueStyle.Render(d.CreatedAt))
	fmt.Printf("Image age: %s\n", valueStyle.Render(fmt.Sprintf("%.0f days", d.AgeDays)))
	fmt.Println(statusPrefix(r.Passed) + r.Message)
}

func renderSizeText(r *output.CheckResult) {
	d := r.Details.(output.SizeDetails)
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking size and layers of image %s", r.Image)))
	fmt.Printf("Number of layers: %s\n", valueStyle.Render(fmt.Sprintf("%d", d.LayerCount)))
	// #nosec G115 -- LayerCount is always non-negative (derived from layer enumeration)
	if uint(d.LayerCount) > d.MaxLayers {
		fmt.Printf("Image has more than %s layers\n", valueStyle.Render(fmt.Sprintf("%d", d.MaxLayers)))
	}
	for _, l := range d.Layers {
		fmt.Printf("  Layer %d: %s\n", l.Index, dimStyle.Render(fmt.Sprintf("%d bytes", l.Bytes)))
	}
	fmt.Printf("Total size: %s\n", valueStyle.Render(fmt.Sprintf("%d bytes (%.2f MB)", d.TotalBytes, d.TotalMB)))
	fmt.Println(statusPrefix(r.Passed) + r.Message)
}

func renderPortsText(r *output.CheckResult) {
	d := r.Details.(output.PortsDetails)
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking ports of image %s", r.Image)))

	if len(d.ExposedPorts) == 0 {
		fmt.Println(statusPrefix(r.Passed) + "No ports are exposed in this image")
		return
	}

	fmt.Println(keyStyle.Render("Exposed ports:"))
	for _, port := range d.ExposedPorts {
		fmt.Printf("  - %s\n", valueStyle.Render(fmt.Sprintf("%d", port)))
	}

	if len(d.AllowedPorts) == 0 {
		fmt.Println("No allowed ports were provided")
		return
	}

	if len(d.UnauthorizedPorts) > 0 {
		fmt.Println(keyStyle.Render("The following ports are not in the allowed list:"))
		for _, port := range d.UnauthorizedPorts {
			fmt.Printf("  - %s\n", FailStyle.Render(fmt.Sprintf("%d", port)))
		}
	}

	if r.Message != "" {
		fmt.Println(statusPrefix(r.Passed) + r.Message)
	}
}

func renderRegistryText(r *output.CheckResult) {
	d := r.Details.(output.RegistryDetails)
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking registry of image %s", r.Image)))

	if d.Skipped {
		fmt.Println(dimStyle.Render("Registry validation skipped (not applicable for this transport)"))
		return
	}

	fmt.Printf("Image registry: %s\n", valueStyle.Render(d.Registry))
	fmt.Println(statusPrefix(r.Passed) + r.Message)
}

func renderRootUserText(r *output.CheckResult) {
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking if image %s is configured to run as a non-root user", r.Image)))
	fmt.Println(statusPrefix(r.Passed) + r.Message)
}

func renderSecretsText(r *output.CheckResult) {
	d := r.Details.(output.SecretsDetails)
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking secrets in image %s", r.Image)))

	if len(d.EnvVarFindings) > 0 {
		fmt.Printf("\n%s\n", keyStyle.Render("Environment Variables:"))
		for _, finding := range d.EnvVarFindings {
			fmt.Printf("  - %s (%s)\n", FailStyle.Render(finding.Name), finding.Description)
		}
	}

	if len(d.FileFindings) > 0 {
		fmt.Printf("\n%s\n", keyStyle.Render("Files with Sensitive Patterns:"))

		// Group findings by layer for better readability
		layerMap := make(map[int][]output.FileFinding)
		for _, finding := range d.FileFindings {
			layerMap[finding.LayerIndex] = append(layerMap[finding.LayerIndex], finding)
		}

		for layerIdx := 0; layerIdx < len(layerMap)+10; layerIdx++ {
			if findings, ok := layerMap[layerIdx]; ok {
				fmt.Printf("  Layer %d:\n", layerIdx+1)
				for _, finding := range findings {
					fmt.Printf("    - %s (%s)\n", FailStyle.Render(finding.Path), finding.Description)
				}
			}
		}
	}

	fmt.Printf("\nTotal findings: %s", valueStyle.Render(fmt.Sprintf("%d", d.TotalFindings)))
	if d.EnvVarCount >= 0 && d.FileCount >= 0 && (d.EnvVarCount > 0 || d.FileCount > 0 || d.TotalFindings > 0) {
		fmt.Printf(" (%d environment variables, %d files)\n", d.EnvVarCount, d.FileCount)
	} else {
		fmt.Println()
	}

	fmt.Println(statusPrefix(r.Passed) + r.Message)
}

func renderHealthcheckText(r *output.CheckResult) {
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking if image %s has a healthcheck defined", r.Image)))
	fmt.Println(statusPrefix(r.Passed) + r.Message)
}

func renderEntrypointText(r *output.CheckResult) {
	d := r.Details.(output.EntrypointDetails)
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking entrypoint of image %s", r.Image)))
	if len(d.Entrypoint) > 0 {
		fmt.Printf("Entrypoint: %s\n", valueStyle.Render(fmt.Sprintf("%v", d.Entrypoint)))
	}
	if len(d.Cmd) > 0 {
		fmt.Printf("Cmd: %s\n", valueStyle.Render(fmt.Sprintf("%v", d.Cmd)))
	}
	fmt.Println(statusPrefix(r.Passed) + r.Message)
}

func renderPlatformText(r *output.CheckResult) {
	d := r.Details.(output.PlatformDetails)
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking platform of image %s", r.Image)))
	fmt.Printf("Image platform: %s\n", valueStyle.Render(d.Platform))
	fmt.Println(statusPrefix(r.Passed) + r.Message)
}

func renderLabelsText(r *output.CheckResult) {
	d := r.Details.(output.LabelsDetails)
	fmt.Println(headerStyle.Render(fmt.Sprintf("Checking labels of image %s", r.Image)))

	// Show required labels
	if len(d.RequiredLabels) > 0 {
		fmt.Printf("\n%s\n", keyStyle.Render("Required labels:"))
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
		fmt.Printf("\n%s\n", keyStyle.Render(fmt.Sprintf("Actual labels (%d):", len(d.ActualLabels))))
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
		fmt.Printf("\n%s\n", keyStyle.Render(fmt.Sprintf("Missing labels (%d):", len(d.MissingLabels))))
		for _, name := range d.MissingLabels {
			fmt.Printf("  - %s\n", FailStyle.Render(name))
		}
	}

	// Show invalid labels
	if len(d.InvalidLabels) > 0 {
		fmt.Printf("\n%s\n", keyStyle.Render(fmt.Sprintf("Invalid labels (%d):", len(d.InvalidLabels))))
		for _, inv := range d.InvalidLabels {
			fmt.Printf("  - %s: %s\n", FailStyle.Render(inv.Name), inv.Reason)
		}
	}

	fmt.Printf("\n%s\n", statusPrefix(r.Passed)+r.Message)
}
