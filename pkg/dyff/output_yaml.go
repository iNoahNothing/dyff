package dyff

import (
	"bufio"
	"github.com/gonvenience/neat"
	"io"
)

type YAMLReport struct {
	Report
}

// I want the output to be
// apiVersion:
// kind:
// metadata:
//   name:
//   namespace:
// diffs:
//   - path:
//     kind:
//     from:
//     to:
//   - path:
//     kind:
//     from:
//     to:

type YAMLReportOutput struct {
	Path    string
	File    string
	Details []Detail
}

func (report *YAMLReport) WriteReport(out io.Writer) error {
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for _, diff := range report.Diffs {
		if err := report.generateYAMLDiffOutput(writer, diff); err != nil {
			return err
		}
	}
	_, _ = writer.WriteString("\n") // Ensure a newline at the end of the report
	return nil
}

func (report *YAMLReport) generateYAMLDiffOutput(writer *bufio.Writer, diff Diff) error {
	data := &YAMLReportOutput{
		Path:    diff.Path.String(),
		File:    diff.Path.RootDescription(),
		Details: diff.Details,
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

	return nil
}
