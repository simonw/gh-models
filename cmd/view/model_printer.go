package view

import (
	"strings"

	"github.com/cli/cli/v2/pkg/markdown"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/pkg/command"
	"github.com/mgutz/ansi"
)

var (
	lightGrayUnderline = ansi.ColorFunc("white+du")
)

type modelPrinter struct {
	modelSummary  *azuremodels.ModelSummary
	modelDetails  *azuremodels.ModelDetails
	printer       tableprinter.TablePrinter
	terminalWidth int
}

func newModelPrinter(summary *azuremodels.ModelSummary, details *azuremodels.ModelDetails, cfg *command.Config) modelPrinter {
	return modelPrinter{
		modelSummary:  summary,
		modelDetails:  details,
		printer:       cfg.NewTablePrinter(),
		terminalWidth: cfg.TerminalWidth,
	}
}

func (p *modelPrinter) render() error {
	modelSummary := p.modelSummary
	if modelSummary != nil {
		p.printLabelledLine("Display name:", modelSummary.FriendlyName)
		p.printLabelledLine("Model name:", modelSummary.Name)
		p.printLabelledLine("Publisher:", modelSummary.Publisher)
		p.printLabelledLine("Summary:", modelSummary.Summary)
	}

	modelDetails := p.modelDetails
	if modelDetails != nil {
		p.printLabelledLine("Context:", modelDetails.ContextLimits())
		p.printLabelledLine("Rate limit tier:", modelDetails.RateLimitTier)
		p.printLabelledList("Tags:", modelDetails.Tags)
		p.printLabelledList("Supported input types:", modelDetails.SupportedInputModalities)
		p.printLabelledList("Supported output types:", modelDetails.SupportedOutputModalities)
		p.printLabelledMultiLineList("Supported languages:", modelDetails.SupportedLanguages)
		p.printLabelledLine("License:", modelDetails.License)
		p.printMultipleLinesWithLabel("License description:", modelDetails.LicenseDescription)
		p.printMultipleLinesWithLabel("Description:", modelDetails.Description)
		p.printMultipleLinesWithLabel("Notes:", modelDetails.Notes)
		p.printMultipleLinesWithLabel("Evaluation:", modelDetails.Evaluation)
	}

	err := p.printer.Render()
	if err != nil {
		return err
	}

	return nil
}

func (p *modelPrinter) printLabelledLine(label, value string) {
	if value == "" {
		return
	}
	p.addLabel(label)
	p.printer.AddField(strings.TrimSpace(value))
	p.printer.EndRow()
}

func (p *modelPrinter) printLabelledList(label string, values []string) {
	p.printLabelledLine(label, strings.Join(values, ", "))
}

func (p *modelPrinter) printLabelledMultiLineList(label string, values []string) {
	p.printMultipleLinesWithLabel(label, strings.Join(values, ", "))
}

func (p *modelPrinter) printMultipleLinesWithLabel(label, value string) {
	if value == "" {
		return
	}
	p.addLabel(label)
	renderedValue, err := markdown.Render(strings.TrimSpace(value), markdown.WithWrap(p.terminalWidth))
	displayValue := value
	if err == nil {
		displayValue = renderedValue
	}
	p.printer.AddField(displayValue, tableprinter.WithTruncate(nil))
	p.printer.EndRow()
}

func (p *modelPrinter) addLabel(label string) {
	p.printer.AddField(label, tableprinter.WithTruncate(nil), tableprinter.WithColor(lightGrayUnderline))
}
