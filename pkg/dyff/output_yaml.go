package dyff

import (
	"bufio"
	"github.com/gonvenience/neat"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type YAMLReport struct {
	Report
}

type YAMLReportDiff struct {
	Details []Detail
	Path    string
}
type YAMLReportOutput struct {
	APIVersion string
	Kind       string
	Metadata   metav1.ObjectMeta
	Diffs      []YAMLReportDiff
}

// TODO: Support non-Kubernetes yaml documents
func (report *YAMLReport) WriteReport(out io.Writer) error {
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for file, diffs := range report.consolidateDiff() {
		meta, err := K8sMetaFromName(file)
		if err != nil {
			return err
		}
		var d []YAMLReportDiff
		for _, diff := range diffs {
			d = append(d, YAMLReportDiff{
				Path:    diff.Path.String(),
				Details: diff.Details,
			})
		}
		data := &YAMLReportOutput{
			APIVersion: meta.APIVersion,
			Kind:       meta.Kind,
			Metadata:   meta.Metadata,
			Diffs:      d,
		}

		// Use neat to format the YAML output
		neatProcessor := neat.NewOutputProcessor(false, true, nil)
		yamlData, err := neatProcessor.ToYAML(data)
		if err != nil {
			return err
		}

		if _, err := writer.WriteString(yamlData); err != nil {
			return err
		}
	}

	_, _ = writer.WriteString("\n") // Ensure a newline at the end of the report
	return nil
}

func (report *YAMLReport) consolidateDiff() map[string][]Diff {
	fileDiffs := make(map[string][]Diff)

	for _, diff := range report.Diffs {
		var deet Detail
		switch len(diff.Details) {
		case 1:
			deet = diff.Details[0]
		case 2:
			for _, detail := range diff.Details {
				switch detail.Kind {
				case ADDITION:
					deet.To = detail.To
				case REMOVAL:
					deet.From = detail.From
				}
			}
			deet.Kind = MODIFICATION
		}

		fileDiffs[diff.Path.RootDescription()] = append(fileDiffs[diff.Path.RootDescription()], Diff{
			Path:    diff.Path,
			Details: []Detail{deet},
		})
	}

	return fileDiffs
}
