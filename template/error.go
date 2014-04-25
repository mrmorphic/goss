package template

import (
	"fmt"
)

type errorLocationReporter interface {
	// This function returns the filename and line number for reporting
	errorLocation() (string, int)
}

// templateError represents an error in scanning/parsing/executing.
type templateError struct {
	file    string
	line    int
	message string
}

func newTemplateError(message string, reporter errorLocationReporter) *templateError {
	file, line := reporter.errorLocation()
	return &templateError{file: file, line: line, message: message}
}

func (p *templateError) Error() string {
	if p.file == "" {
		return fmt.Sprintf("line %d: %s", p.line, p.message)
	}
	return fmt.Sprintf("%s (line %d): %s", p.file, p.line, p.message)
}
