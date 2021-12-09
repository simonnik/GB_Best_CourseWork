package scanner

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/simonnik/GB_Best_CourseWork_GO/services/parser"
)

type ScanResult struct {
	Err      error
	Results  map[string]string
	Finished bool
}

type Scann struct {
	File    *os.File
	Reader  *csv.Reader
	Fields  []string
	Results chan ScanResult
}

func NewScanner(filename string) (*Scann, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &Scann{
		File:    f,
		Reader:  csv.NewReader(f),
		Fields:  make([]string, 0),
		Results: make(chan ScanResult),
	}, nil
}

func (s *Scann) GetHeaders() ([]string, error) {
	r, err := s.Reader.Read()
	if err != nil {
		return nil, err
	}
	s.Fields = r

	return r, nil
}

func (s *Scann) PrepareRow(columns []string) map[string]string {
	row := make(map[string]string)
	for i, item := range columns {
		row[s.Fields[i]] = item
	}

	return row
}

func (s *Scann) MapFieldsToRow(row map[string]string, fields []string) map[string]string {
	if fields[0] == "*" {
		return row
	}
	res := make(map[string]string)
	fmt.Print(fields)
	for _, item := range fields {
		res[item] = row[item]
	}

	return res
}

func (s *Scann) ChanResult() <-chan ScanResult {
	return s.Results
}

func (s *Scann) Scan(ctx context.Context, query parser.Query) {
	select {
	case <-ctx.Done(): // Если контекст завершен - прекращаем выполнение, (lint: gocritic)
		return
	default:
		for {
			r, err := s.Reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				s.Results <- ScanResult{Err: err} // Записываем ошибку в канал, (lint: gocritic)
				return
			}
			row := s.PrepareRow(r)
			if s.IsApply(row, query) {
				s.Results <- ScanResult{
					Results: s.MapFieldsToRow(row, query.Fields),
				}
			}
		}

		s.Results <- ScanResult{
			Finished: true,
		}
	}
}

func (s *Scann) IsApply(row map[string]string, query parser.Query) bool {
	var isValid bool
	if len(query.Conditions) == 0 {
		return true
	}
	for i, cond := range query.Conditions {
		if val, ok := row[cond.Operand1]; !ok {
			s.Results <- ScanResult{Err: fmt.Errorf(
				"operand \"%s\" is undefined", cond.Operand1)} // Записываем ошибку в канал
			break
		} else {
			var c bool
			switch cond.Operator {
			case parser.Eq:
				c = val == cond.Operand2
			case parser.Gt:
				c = val > cond.Operand2
			case parser.Gte:
				c = val >= cond.Operand2
			case parser.Lt:
				c = val < cond.Operand2
			case parser.Lte:
				c = val <= cond.Operand2
			}

			if i > 0 {
				if cond.Condition == parser.And {
					isValid = isValid && c
				} else {
					isValid = isValid || c
				}
			} else {
				isValid = c
			}
		}
	}

	return isValid
}
