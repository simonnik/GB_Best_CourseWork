package scanner

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/simonnik/GB_Best_CourseWork_GO/services/parser"
	"github.com/stretchr/testify/assert"
)

func TestNewScannerWithEmptyFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantRes  bool
		wantErr  bool
	}{
		{name: "Empty file", filename: "", wantRes: false, wantErr: true},
		{name: "File not exist", filename: "../employees.csv", wantRes: false, wantErr: true},
		{name: "File exists", filename: "../../employees.csv", wantRes: true, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewScanner(tt.filename)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			if tt.wantRes {
				defer res.File.Close()
				assert.NotNil(t, res)
			} else {
				assert.Nil(t, res)
			}
		})
	}
}

func TestScann_ChanResult(t *testing.T) {
	s, _ := NewScanner("../../employees.csv")
	defer s.File.Close()

	sRes := ScanResult{
		Err:      nil,
		Results:  map[string]string{"id": "1"},
		Finished: false,
	}
	go func() {
		s.Results <- sRes
	}()

	res := <-s.ChanResult()
	assert.Equal(t, sRes, res)
}

func TestScann_GetHeaders(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantRes  bool
		wantErr  bool
	}{
		{name: "File true", filename: "../../employees.csv", wantRes: true, wantErr: false},
		{name: "File false", filename: "../../fail.csv", wantRes: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := NewScanner(tt.filename)
			defer s.File.Close()

			res, err := s.GetHeaders()
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			if tt.wantRes {
				assert.NotNil(t, res)
			} else {
				assert.Nil(t, res)
			}
		})
	}
}

func TestScann_IsApply(t *testing.T) {
	type args struct {
		row   map[string]string
		query parser.Query
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Empty conditions",
			args: args{
				row:   map[string]string{},
				query: parser.Query{},
			},
			want: true,
		},
		{
			name: "operand is undefined",
			args: args{
				row: map[string]string{"id": "10"},
				query: parser.Query{
					Conditions: []parser.Condition{
						{OperandLeft: "id", Operator: parser.Eq, OperandRight: "10"},
						{OperandLeft: "id", Operator: parser.Gt, OperandRight: "1"},
						{OperandLeft: "id", Operator: parser.Gte, OperandRight: "1"},
						{OperandLeft: "id", Operator: parser.Lt, OperandRight: "11"},
						{OperandLeft: "id", Operator: parser.Lte, OperandRight: "10", Condition: parser.And},
					},
				},
			},
			want: true,
		},
	}
	s := &Scann{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.IsApply(tt.args.row, tt.args.query); got != tt.want {
				t.Errorf("IsApply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScann_MapFieldsToRow(t *testing.T) {
	type args struct {
		row    map[string]string
		fields []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "*",
			args: args{
				row:    map[string]string{"id": "1"},
				fields: []string{"*"},
			},
			want: map[string]string{"id": "1"},
		},
		{
			name: "id, name",
			args: args{
				row:    map[string]string{"id": "10", "name": "Ford"},
				fields: []string{"id"},
			},
			want: map[string]string{"id": "10"},
		},
	}
	s := &Scann{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.MapFieldsToRow(tt.args.row, tt.args.fields); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapFieldsToRow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScann_PrepareRow(t *testing.T) {
	type args struct {
		columns []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "*",
			args: args{
				columns: []string{"1", "Ford"},
			},
			want: map[string]string{"id": "1", "name": "Ford"},
		},
	}
	s := &Scann{
		Fields: []string{"id", "name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.PrepareRow(tt.args.columns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrepareRow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScann_Scan(t *testing.T) {
	ctx := context.Background()
	p := parser.NewParser("SELECT * FROM '../../employees.csv'")
	q, _ := p.Parse()
	s, _ := NewScanner(q.TableName)
	defer s.File.Close()
	s.GetHeaders()
	go s.Scan(ctx, *q)

	doFor := true
	for doFor {
		select {
		case msg := <-s.ChanResult():
			switch {
			case msg.Err != nil:
				doFor = false
			case len(msg.Results) > 0:
				doFor = false
				fmt.Printf("%s\n", msg.Results)
			case msg.Finished:
				doFor = false
			}
		}
	}
}
