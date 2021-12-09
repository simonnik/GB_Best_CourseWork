package domain

import (
	"github.com/simonnik/GB_Best_CourseWork_GO/services/parser"
)

type Parser interface {
	Parse() (parser.Query, error)
	DoParse() (parser.Query, error)
	Peek() string
	Pop() string
	PopWhitespace()
	PeekWithLength() (string, int)
	PeekQuotedStringWithLength() (string, int)
	PeekIdentifierWithLength() (string, int)
	Validate() error
	LogError()
}
