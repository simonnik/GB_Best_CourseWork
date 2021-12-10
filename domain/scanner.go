//go:generate mockery --name=Scanner
package domain

import (
	"context"

	"github.com/simonnik/GB_Best_CourseWork_GO/services/parser"
	"github.com/simonnik/GB_Best_CourseWork_GO/services/scanner"
)

type Scanner interface {
	GetHeaders() ([]string, error)
	Scan(ctx context.Context, query parser.Query)
	PrepareRow([]string) map[string]string
	MapFieldsToRow(row map[string]string, fields []string) map[string]string
	IsApply(row map[string]string, query parser.Query) bool
	ChanResult() <-chan scanner.ScanResult
}
