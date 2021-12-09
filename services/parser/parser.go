package parser

import (
	"fmt"
	"regexp"
	"strings"
)

type step int

const (
	stepType step = iota
	stepSelectField
	stepSelectFrom
	stepSelectComma
	stepSelectFromTable
	stepWhere
	stepWhereField
	stepWhereOperator
	stepWhereValue
	stepWhereCondition
)

type Parser struct {
	i     int
	sql   string
	step  step
	query Query
	err   error
}

func NewParser(sql string) *Parser {
	return &Parser{0, strings.TrimSpace(sql), stepType, Query{}, nil}
}

// Parse takes a string representing a SQL query and parses it into a Query struct. It may fail.
func (p *Parser) Parse() (Query, error) {
	q, err := p.DoParse()
	p.err = err
	if p.err == nil {
		p.err = p.Validate()
	}
	p.LogError()
	return q, p.err
}

func (p *Parser) DoParse() (Query, error) {
	for {
		if p.i >= len(p.sql) {
			return p.query, p.err
		}
		switch p.step {
		case stepType:
			switch strings.ToUpper(p.Peek()) {
			case "SELECT":
				p.query.Type = Select
				p.Pop()
				p.step = stepSelectField
			default:
				return p.query, fmt.Errorf("invalid query type")
			}
		case stepSelectField:
			identifier := p.Peek()
			if !isIdentifierOrAsterisk(identifier) {
				return p.query, fmt.Errorf("at SELECT: expected field to SELECT")
			}
			p.query.Fields = append(p.query.Fields, identifier)
			p.Pop()
			maybeFrom := p.Peek()
			if strings.ToUpper(maybeFrom) == "FROM" {
				p.step = stepSelectFrom
				continue
			}
			p.step = stepSelectComma
		case stepSelectComma:
			commaRWord := p.Peek()
			if commaRWord != "," {
				return p.query, fmt.Errorf("at SELECT: expected comma or FROM")
			}
			p.Pop()
			p.step = stepSelectField
		case stepSelectFrom:
			fromRWord := p.Peek()
			if strings.ToUpper(fromRWord) != "FROM" {
				return p.query, fmt.Errorf("at SELECT: expected FROM")
			}
			p.Pop()
			p.step = stepSelectFromTable
		case stepSelectFromTable:
			tableName := p.Peek()
			if len(tableName) == 0 {
				return p.query, fmt.Errorf("at SELECT: expected quoted table name")
			}
			p.query.TableName = tableName
			p.Pop()
			p.step = stepWhere
		case stepWhere:
			whereRWord := p.Peek()
			if strings.ToUpper(whereRWord) != "WHERE" {
				return p.query, fmt.Errorf("expected WHERE")
			}
			p.Pop()
			p.step = stepWhereField
		case stepWhereField:
			identifier := p.Peek()
			if !isIdentifier(identifier) {
				return p.query, fmt.Errorf("at WHERE: expected field")
			}
			cond := Condition{
				Operand1:        identifier,
				Operand1IsField: true,
			}
			if p.query.LastCondWhere != UnknownCondWhere {
				cond.Condition = p.query.LastCondWhere
			}
			p.query.Conditions = append(p.query.Conditions, cond)
			p.Pop()
			p.step = stepWhereOperator
		case stepWhereOperator:
			operator := p.Peek()
			currentCondition := p.query.Conditions[len(p.query.Conditions)-1]
			switch operator {
			case "=":
				currentCondition.Operator = Eq
			case ">":
				currentCondition.Operator = Gt
			case ">=":
				currentCondition.Operator = Gte
			case "<":
				currentCondition.Operator = Lt
			case "<=":
				currentCondition.Operator = Lte
			case "!=":
				currentCondition.Operator = Ne
			default:
				return p.query, fmt.Errorf("at WHERE: unknown operator")
			}
			p.query.Conditions[len(p.query.Conditions)-1] = currentCondition
			p.Pop()
			p.step = stepWhereValue
		case stepWhereValue:
			currentCondition := p.query.Conditions[len(p.query.Conditions)-1]
			identifier := p.Peek()
			currentCondition.Operand2 = identifier
			currentCondition.Operand2IsField = isIdentifier(identifier)
			p.query.Conditions[len(p.query.Conditions)-1] = currentCondition
			p.Pop()
			p.step = stepWhereCondition
		case stepWhereCondition:
			andRWord := p.Peek()
			switch strings.ToUpper(andRWord) {
			case "AND":
				p.query.LastCondWhere = And
			case "OR":
				p.query.LastCondWhere = Or
			default:
				return p.query, fmt.Errorf("at Condition: unknown condition")
			}
			p.Pop()
			p.step = stepWhereField
		}
	}
}

func (p *Parser) Peek() string {
	peeked, _ := p.PeekWithLength()
	return peeked
}

func (p *Parser) Pop() string {
	peeked, length := p.PeekWithLength()
	p.i += length
	p.PopWhitespace()
	return peeked
}

func (p *Parser) PopWhitespace() {
	for ; p.i < len(p.sql) && p.sql[p.i] == ' '; p.i++ {
	}
}

func reservedWords() []string {
	return []string{
		"(", ")", ">=", "<=", "!=", ",", "=", ">", "<", "SELECT", "WHERE", "FROM",
	}
}

func (p *Parser) PeekWithLength() (string, int) {
	if p.i >= len(p.sql) {
		return "", 0
	}
	for _, rWord := range reservedWords() {
		token := strings.ToUpper(p.sql[p.i:min(len(p.sql), p.i+len(rWord))])
		if token == rWord {
			return token, len(token)
		}
	}
	if p.sql[p.i] == '\'' { // Quoted string
		return p.PeekQuotedStringWithLength()
	}
	return p.PeekIdentifierWithLength()
}

func (p *Parser) PeekQuotedStringWithLength() (string, int) {
	if len(p.sql) < p.i || p.sql[p.i] != '\'' {
		return "", 0
	}
	for i := p.i + 1; i < len(p.sql); i++ {
		if p.sql[i] == '\'' && p.sql[i-1] != '\\' {
			return p.sql[p.i+1 : i], len(p.sql[p.i+1:i]) + 2 // +2 for the two quotes
		}
	}
	return "", 0
}

func (p *Parser) PeekIdentifierWithLength() (string, int) {
	for i := p.i; i < len(p.sql); i++ {
		re := regexp.MustCompile(`[a-zA-Z0-9_*.]`)
		if !re.MatchString(string(p.sql[i])) {
			return p.sql[p.i:i], len(p.sql[p.i:i])
		}
	}
	return p.sql[p.i:], len(p.sql[p.i:])
}

func (p *Parser) Validate() error {
	if len(p.query.Conditions) == 0 && p.step == stepWhereField {
		return fmt.Errorf("at WHERE: empty WHERE clause")
	}
	if p.query.Type == UnknownType {
		return fmt.Errorf("query type cannot be empty")
	}
	if p.query.TableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}
	for _, c := range p.query.Conditions {
		if c.Operator == UnknownOperator {
			return fmt.Errorf("at WHERE: condition without operator")
		}
		if c.Operand1 == "" && c.Operand1IsField {
			return fmt.Errorf("at WHERE: condition with empty left side operand")
		}
		if c.Operand2 == "" && c.Operand2IsField {
			return fmt.Errorf("at WHERE: condition with empty right side operand")
		}
	}
	return nil
}

func (p *Parser) LogError() {
	if p.err == nil {
		return
	}
	fmt.Println(p.sql)
	fmt.Println(strings.Repeat(" ", p.i) + "^")
	fmt.Println(p.err)
}

func isIdentifier(s string) bool {
	for _, rw := range reservedWords() {
		if strings.ToUpper(s) == rw {
			return false
		}
	}
	matched, _ := regexp.MatchString("[a-zA-Z_][a-zA-Z_0-9]*", s)
	return matched
}

func isIdentifierOrAsterisk(s string) bool {
	return isIdentifier(s) || s == "*"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
