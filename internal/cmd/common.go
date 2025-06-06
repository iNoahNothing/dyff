// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gonvenience/bunt"
	"github.com/gonvenience/neat"
	"github.com/gonvenience/ytbx"
	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"

	"github.com/homeport/dyff/pkg/dyff"
)

type reportConfig struct {
	Style                     string   `yaml:"style"`
	IgnoreOrderChanges        bool     `yaml:"ignore-order-changes"`
	IgnoreWhitespaceChanges   bool     `yaml:"ignore-whitespace-changes"`
	KubernetesEntityDetection bool     `yaml:"kubernetes-entity-detection"`
	NoTableStyle              bool     `yaml:"no-table-style"`
	DoNotInspectCerts         bool     `yaml:"do-not-inspect-certs"`
	ExitWithCode              bool     `yaml:"exit-with-code"`
	OmitHeader                bool     `yaml:"omit-header"`
	UseGoPatchPaths           bool     `yaml:"use-go-patch-paths"`
	IgnoreValueChanges        bool     `yaml:"ignore-value-changes"`
	IgnoreNewDocuments        bool     `yaml:"ignore-new-documents"`
	DetectRenames             bool     `yaml:"detect-renames"`
	MarshalJsonStrings        bool     `yaml:"marshal-json-strings"`
	ChompBlockScalars         bool     `yaml:"chomp-block-scalars"`
	MinorChangeThreshold      float64  `yaml:"minor-change-threshold"`
	MultilineContextLines     int      `yaml:"multiline-context-lines"`
	AdditionalIdentifiers     []string `yaml:"additional-identifier"`
	Filters                   []string `yaml:"filter"`
	Excludes                  []string `yaml:"exclude"`
	FilterRegexps             []string `yaml:"filter-regexp"`
	ExcludeRegexps            []string `yaml:"exclude-regexp"`
	FilterDocuments           []string `yaml:"filter-document"`
	ExcludeDocuments          []string `yaml:"exclude-document"`
	FilterDocumentRegexps     []string `yaml:"filter-document-regexp"`
	ExcludeDocumentRegexps    []string `yaml:"exclude-document-regexp"`
}

var defaults = reportConfig{
	Style:                     "human",
	IgnoreOrderChanges:        false,
	IgnoreWhitespaceChanges:   false,
	KubernetesEntityDetection: true,
	NoTableStyle:              false,
	DoNotInspectCerts:         false,
	ExitWithCode:              false,
	OmitHeader:                false,
	UseGoPatchPaths:           false,
	IgnoreValueChanges:        false,
	IgnoreNewDocuments:        false,
	DetectRenames:             true,
	MarshalJsonStrings:        false,
	ChompBlockScalars:         false,
	MinorChangeThreshold:      0.1,
	MultilineContextLines:     4,
	AdditionalIdentifiers:     nil,
	Filters:                   nil,
	Excludes:                  nil,
	FilterRegexps:             nil,
	ExcludeRegexps:            nil,
	FilterDocuments:           nil,
	ExcludeDocuments:          nil,
	FilterDocumentRegexps:     nil,
	ExcludeDocumentRegexps:    nil,
}

var reportOptions reportConfig

func applyReportOptionsFlags(cmd *cobra.Command) {
	// Compare options
	cmd.Flags().BoolVarP(&reportOptions.IgnoreOrderChanges, "ignore-order-changes", "i", defaults.IgnoreOrderChanges, "ignore order changes in lists")
	cmd.Flags().BoolVar(&reportOptions.IgnoreWhitespaceChanges, "ignore-whitespace-changes", defaults.IgnoreWhitespaceChanges, "ignore leading or trailing whitespace changes")
	cmd.Flags().BoolVarP(&reportOptions.KubernetesEntityDetection, "detect-kubernetes", "", defaults.KubernetesEntityDetection, "detect kubernetes entities")
	cmd.Flags().StringArrayVar(&reportOptions.AdditionalIdentifiers, "additional-identifier", defaults.AdditionalIdentifiers, "use additional identifier candidates in named entry lists")
	cmd.Flags().StringSliceVar(&reportOptions.Filters, "filter", defaults.Filters, "filter reports to a subset of differences based on supplied arguments")
	cmd.Flags().StringSliceVar(&reportOptions.Excludes, "exclude", defaults.Excludes, "exclude reports from a set of differences based on supplied arguments")
	cmd.Flags().StringSliceVar(&reportOptions.FilterRegexps, "filter-regexp", defaults.FilterRegexps, "filter reports to a subset of differences based on supplied regular expressions")
	cmd.Flags().StringSliceVar(&reportOptions.ExcludeRegexps, "exclude-regexp", defaults.ExcludeRegexps, "exclude reports from a set of differences based on supplied regular expressions")
	cmd.Flags().StringSliceVar(&reportOptions.FilterDocuments, "filter-document", defaults.FilterDocuments, "filter report to a subset of documents")
	cmd.Flags().StringSliceVar(&reportOptions.ExcludeDocuments, "exclude-document", defaults.ExcludeDocuments, "exclude documents from report")
	cmd.Flags().StringSliceVar(&reportOptions.FilterDocumentRegexps, "filter-document-regexp", defaults.FilterDocumentRegexps, "filter report to a subset of documents based on supplied regular expressions")
	cmd.Flags().StringSliceVar(&reportOptions.ExcludeDocumentRegexps, "exclude-document-regexp", defaults.ExcludeDocumentRegexps, "exclude documents from report based on supplied regular expressions")
	cmd.Flags().BoolVarP(&reportOptions.IgnoreValueChanges, "ignore-value-changes", "v", defaults.IgnoreValueChanges, "exclude changes in values")
	cmd.Flags().BoolVar(&reportOptions.IgnoreNewDocuments, "ignore-new-documents", defaults.IgnoreNewDocuments, "exclude new documents")
	cmd.Flags().BoolVar(&reportOptions.DetectRenames, "detect-renames", defaults.DetectRenames, "enable detection for renames (document level for Kubernetes resources)")
	cmd.Flags().BoolVar(&reportOptions.MarshalJsonStrings, "marshal-json-strings", defaults.MarshalJsonStrings, "marshal Json strings for comparison, otherwise compare unformatted strings")
	cmd.Flags().BoolVar(&reportOptions.ChompBlockScalars, "chomp-block-scalars", defaults.ChompBlockScalars, "chomp block scalars for comparison, otherwise compare unformatted strings")

	// Main output preferences
	cmd.Flags().StringVarP(&reportOptions.Style, "output", "o", defaults.Style, "specify the output style, supported styles: human, brief, github, gitlab, gitea, yaml")
	cmd.Flags().BoolVarP(&reportOptions.OmitHeader, "omit-header", "b", defaults.OmitHeader, "omit the dyff summary header")
	cmd.Flags().BoolVarP(&reportOptions.ExitWithCode, "set-exit-code", "s", defaults.ExitWithCode, "set program exit code, with 0 meaning no difference, 1 for differences detected, and 255 for program error")

	// Human/BOSH output related flags
	cmd.Flags().BoolVarP(&reportOptions.NoTableStyle, "no-table-style", "l", defaults.NoTableStyle, "do not place blocks next to each other, always use one row per text block")
	cmd.Flags().BoolVarP(&reportOptions.DoNotInspectCerts, "no-cert-inspection", "x", defaults.DoNotInspectCerts, "disable x509 certificate inspection, compare as raw text")
	cmd.Flags().BoolVarP(&reportOptions.UseGoPatchPaths, "use-go-patch-style", "g", defaults.UseGoPatchPaths, "use Go-Patch style paths in outputs")
	cmd.Flags().Float64VarP(&reportOptions.MinorChangeThreshold, "minor-change-threshold", "", defaults.MinorChangeThreshold, "minor change threshold")
	cmd.Flags().IntVarP(&reportOptions.MultilineContextLines, "multi-line-context-lines", "", defaults.MultilineContextLines, "multi-line context lines")

	// Deprecated
	cmd.Flags().BoolVar(&reportOptions.ExitWithCode, "set-exit-status", defaults.ExitWithCode, "set program exit code, with 0 meaning no difference, 1 for differences detected, and 255 for program error")
	_ = cmd.Flags().MarkDeprecated("set-exit-status", "use --set-exit-code instead")

	// Set reportOptions from a config file
	var configFile string
	cmd.Flags().StringVar(&configFile, "config", ".dyffconfig.yml", "set dyff options from a yaml config file.")

	if err := cmd.ParseFlags(os.Args); err != nil {
		bunt.Errorf("failed to parse flags: %v", err)
		os.Exit(1)
	}

	if _, err := os.Stat(configFile); err == nil {
		file, err := os.Open(configFile)
		if err != nil {
			bunt.Errorf("failed to open config file %s: %v", configFile, err)
			os.Exit(1)
		}
		defer file.Close()

		decoder := yamlv3.NewDecoder(file)
		if err := decoder.Decode(&reportOptions); err != nil {
			bunt.Errorf("failed to decode config file %s: %v", configFile, err)
			os.Exit(1)
		}
	}
}

// OutputWriter encapsulates the required fields to define the look and feel of
// the output
type OutputWriter struct {
	PlainMode        bool
	Restructure      bool
	OmitIndentHelper bool
	OutputStyle      string
}

func humanReadableFilename(filename string) string {
	if ytbx.IsStdin(filename) {
		return bunt.Sprint("_*stdin*_")
	}

	return bunt.Sprintf("_*%s*_", filename)
}

// WriteToStdout is a convenience function to write the content of the documents
// stored in the provided input file to the standard output
func (w *OutputWriter) WriteToStdout(filename string) error {
	if err := w.write(os.Stdout, filename); err != nil {
		return bunt.Errorf("failed to write output to _*stdout*_: %w", err)
	}

	return nil
}

// WriteInplace writes the content of the documents stored in the provided input
// file to the file itself overwriting the content in place.
func (w *OutputWriter) WriteInplace(filename string) error {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	// Force plain mode to make sure there are no ANSI sequences
	w.PlainMode = true
	if err := w.write(bufWriter, filename); err != nil {
		return fmt.Errorf("failed to write output to %s: %w", humanReadableFilename(filename), err)
	}

	// Write the buffered output to the provided input file (override in place)
	bufWriter.Flush()
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to overwrite %s in place: %w", humanReadableFilename(filename), err)
	}

	return nil
}

func (w *OutputWriter) write(writer io.Writer, filename string) error {
	inputFile, err := ytbx.LoadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load input from %s: %w", humanReadableFilename(filename), err)
	}

	for _, document := range inputFile.Documents {
		if w.Restructure {
			ytbx.RestructureObject(document)
		}

		switch {
		case w.PlainMode && w.OutputStyle == "json":
			output, err := neat.NewOutputProcessor(false, false, &neat.DefaultColorSchema).ToCompactJSON(document)
			if err != nil {
				return err
			}
			fmt.Fprintf(writer, "%s\n", output)

		case w.PlainMode && w.OutputStyle == "yaml":
			fmt.Fprintln(writer, "---")
			encoder := yamlv3.NewEncoder(writer)
			encoder.SetIndent(2)

			if err := encoder.Encode(document); err != nil {
				return err
			}

			if err := encoder.Close(); err != nil {
				return err
			}

		case w.OutputStyle == "json":
			output, err := neat.NewOutputProcessor(!w.OmitIndentHelper, true, &neat.DefaultColorSchema).ToJSON(document)
			if err != nil {
				return err
			}
			fmt.Fprintf(writer, "%s\n", output)

		case w.OutputStyle == "yaml":
			output, err := neat.NewOutputProcessor(!w.OmitIndentHelper, true, &neat.DefaultColorSchema).ToYAML(document)
			if err != nil {
				return err
			}
			fmt.Fprintf(writer, "%s\n", output)
		}
	}

	return nil
}

func writeReport(cmd *cobra.Command, report dyff.Report) error {
	var reportWriter dyff.ReportWriter
	switch strings.ToLower(reportOptions.Style) {
	case "human", "bosh":
		reportWriter = &dyff.HumanReport{
			Report:                report,
			Indent:                2,
			DoNotInspectCerts:     reportOptions.DoNotInspectCerts,
			NoTableStyle:          reportOptions.NoTableStyle,
			OmitHeader:            reportOptions.OmitHeader,
			UseGoPatchPaths:       reportOptions.UseGoPatchPaths,
			MinorChangeThreshold:  reportOptions.MinorChangeThreshold,
			MultilineContextLines: reportOptions.MultilineContextLines,
			PrefixMultiline:       false,
		}

	case "github", "linguist":
		reportWriter = &dyff.DiffSyntaxReport{
			PathPrefix:            "@@",
			RootDescriptionPrefix: "#",
			ChangeTypePrefix:      "!",
			HumanReport: dyff.HumanReport{
				Report:                report,
				Indent:                0,
				DoNotInspectCerts:     reportOptions.DoNotInspectCerts,
				NoTableStyle:          true,
				OmitHeader:            true,
				UseGoPatchPaths:       reportOptions.UseGoPatchPaths,
				MinorChangeThreshold:  reportOptions.MinorChangeThreshold,
				MultilineContextLines: reportOptions.MultilineContextLines,
				PrefixMultiline:       true,
			},
		}

	case "gitlab", "rogue":
		reportWriter = &dyff.DiffSyntaxReport{
			PathPrefix:            "=",
			RootDescriptionPrefix: "=",
			ChangeTypePrefix:      "#",
			HumanReport: dyff.HumanReport{
				Report:                report,
				Indent:                0,
				DoNotInspectCerts:     reportOptions.DoNotInspectCerts,
				NoTableStyle:          true,
				OmitHeader:            true,
				UseGoPatchPaths:       reportOptions.UseGoPatchPaths,
				MinorChangeThreshold:  reportOptions.MinorChangeThreshold,
				MultilineContextLines: reportOptions.MultilineContextLines,
				PrefixMultiline:       true,
			},
		}

	case "gitea", "forgejo":
		reportWriter = &dyff.DiffSyntaxReport{
			PathPrefix:            "@@",
			RootDescriptionPrefix: "=",
			ChangeTypePrefix:      "!",
			HumanReport: dyff.HumanReport{
				Report:                report,
				Indent:                0,
				DoNotInspectCerts:     reportOptions.DoNotInspectCerts,
				NoTableStyle:          true,
				OmitHeader:            true,
				UseGoPatchPaths:       reportOptions.UseGoPatchPaths,
				MinorChangeThreshold:  reportOptions.MinorChangeThreshold,
				MultilineContextLines: reportOptions.MultilineContextLines,
				PrefixMultiline:       true,
			},
		}

	case "yaml", "yml":
		reportWriter = &dyff.YAMLReport{
			Report: report,
		}

	case "brief", "short", "summary":
		reportWriter = &dyff.BriefReport{
			Report: report,
		}

	default:
		return fmt.Errorf("unknown output style %s: %w", reportOptions.Style, fmt.Errorf("%s", cmd.UsageString()))
	}

	if err := reportWriter.WriteReport(os.Stdout); err != nil {
		return fmt.Errorf("failed to print report: %w", err)
	}

	// If configured, make sure `dyff` exists with an exit status
	if reportOptions.ExitWithCode {
		switch len(report.Diffs) {
		case 0:
			return errorWithExitCode{value: 0}

		default:
			return errorWithExitCode{value: 1}
		}
	}

	return nil
}
