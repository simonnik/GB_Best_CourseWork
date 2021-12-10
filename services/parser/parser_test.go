package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewParser(t *testing.T) {
	sql := "SELECT * FROM employees.csv"
	p := NewParser(sql)

	assert.NotNil(t, p)
	assert.Equal(t, sql, p.sql)
}

func TestParseWithInvalidQueryType(t *testing.T) {
	sql := "SELECsT * FROM employeesk.csv"
	p := NewParser(sql)
	res, err := p.Parse()

	assert.NotNil(t, err)
	assert.Equal(t, "invalid query type", err.Error())
	assert.Nil(t, res)
}

func TestParseSuccess(t *testing.T) {
	sql := "SELECT age,name FROM employeesk.csv WHERE age > 10 AND id = 5 OR id < 2 AND age != 30 OR id <= 4 AND id" +
		" >= 2"
	p := NewParser(sql)
	res, err := p.Parse()

	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestParseWithInvalidSelectField(t *testing.T) {
	sql := "SELECT 10 FROM employeesk.csv"
	p := NewParser(sql)
	res, err := p.Parse()

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestParseWithInvalidSelect(t *testing.T) {
	sql := "SELECT age res FROM employeesk.csv"
	p := NewParser(sql)
	res, err := p.Parse()

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestParseWithInvalidTableName(t *testing.T) {
	sql := "SELECT * FROM ''"
	p := NewParser(sql)
	res, err := p.Parse()

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestParseWithInvalidWHERE(t *testing.T) {
	sql := "SELECT age FROM employeesk.csv age > 0"
	p := NewParser(sql)
	res, err := p.Parse()

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestParseWithInvalidWHEREField(t *testing.T) {
	sql := "SELECT age FROM employeesk.csv WHERE 10 > 0"
	p := NewParser(sql)
	res, err := p.Parse()

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestParseWithWHEREUnknownOperator(t *testing.T) {
	sql := "SELECT age FROM employeesk.csv WHERE age ? 0"
	p := NewParser(sql)
	res, err := p.Parse()

	assert.NotNil(t, err)
	assert.Equal(t, "at WHERE: unknown operator", err.Error())
	assert.Nil(t, res)
}

func TestParseWithWHEREUnknownCondition(t *testing.T) {
	sql := "SELECT age FROM employeesk.csv WHERE age = 0 Foo age < 0"
	p := NewParser(sql)
	res, err := p.Parse()

	assert.NotNil(t, err)
	assert.Equal(t, "at Condition: unknown condition", err.Error())
	assert.Nil(t, res)
}

func TestValidateWithEmptyWhereClause(t *testing.T) {
	p := &Parser{}
	p.step = stepWhereField
	err := p.Validate()

	assert.NotNil(t, err)
	assert.Equal(t, "empty WHERE clause", err.Error())
}

func TestValidateWithEmptyQueryType(t *testing.T) {
	p := &Parser{}
	err := p.Validate()

	assert.NotNil(t, err)
	assert.Equal(t, "query type cannot be empty", err.Error())
}

func TestValidateWithoutTableName(t *testing.T) {
	p := &Parser{}
	p.query.Type = Select
	err := p.Validate()

	assert.NotNil(t, err)
	assert.Equal(t, "table name cannot be empty", err.Error())
}

func TestValidateWithoutConditionOperator(t *testing.T) {
	p := &Parser{}
	p.query.Type = Select
	p.query.TableName = "qwe"
	c := Condition{
		Operator: UnknownOperator,
	}
	p.query.Conditions = append(p.query.Conditions, c)
	err := p.Validate()

	assert.NotNil(t, err)
	assert.Equal(t, "condition without operator", err.Error())
}

func TestValidateConditionWithEmptyLeftSideOperand(t *testing.T) {
	p := &Parser{}
	p.query.Type = Select
	p.query.TableName = "qwe"
	c := Condition{
		Operator: Eq,
	}
	p.query.Conditions = append(p.query.Conditions, c)
	err := p.Validate()

	assert.NotNil(t, err)
	assert.Equal(t, "condition with empty left side operand", err.Error())
}

func TestValidateConditionWithEmptyRightSideOperand(t *testing.T) {
	p := &Parser{}
	p.query.Type = Select
	p.query.TableName = "qwe"
	c := Condition{
		Operator:    Eq,
		OperandLeft: "age",
	}
	p.query.Conditions = append(p.query.Conditions, c)
	err := p.Validate()

	assert.NotNil(t, err)
	assert.Equal(t, "condition with empty right side operand", err.Error())
}

func TestValidateSuccess(t *testing.T) {
	p := &Parser{}
	p.query.Type = Select
	p.query.TableName = "qwe"
	c := Condition{
		Operator:     Eq,
		OperandLeft:  "age",
		OperandRight: "10",
	}
	p.query.Conditions = append(p.query.Conditions, c)
	err := p.Validate()

	assert.Nil(t, err)
}
