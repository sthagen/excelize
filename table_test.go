package excelize

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddTable(t *testing.T) {
	f, err := prepareTestBook1()
	assert.NoError(t, err)
	assert.NoError(t, f.AddTable("Sheet1", "B26:A21", nil))
	assert.NoError(t, f.AddTable("Sheet2", "A2:B5", &TableOptions{
		Name:              "table",
		StyleName:         "TableStyleMedium2",
		ShowFirstColumn:   true,
		ShowLastColumn:    true,
		ShowRowStripes:    boolPtr(true),
		ShowColumnStripes: true,
	},
	))
	assert.NoError(t, f.AddTable("Sheet2", "F1:F1", &TableOptions{StyleName: "TableStyleMedium8"}))

	// Test add table in not exist worksheet
	assert.EqualError(t, f.AddTable("SheetN", "B26:A21", nil), "sheet SheetN does not exist")
	// Test add table with illegal cell reference
	assert.EqualError(t, f.AddTable("Sheet1", "A:B1", nil), newCellNameToCoordinatesError("A", newInvalidCellNameError("A")).Error())
	assert.EqualError(t, f.AddTable("Sheet1", "A1:B", nil), newCellNameToCoordinatesError("B", newInvalidCellNameError("B")).Error())

	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestAddTable.xlsx")))

	// Test add table with invalid sheet name
	assert.EqualError(t, f.AddTable("Sheet:1", "B26:A21", nil), ErrSheetNameInvalid.Error())
	// Test addTable with illegal cell reference
	f = NewFile()
	assert.EqualError(t, f.addTable("sheet1", "", 0, 0, 0, 0, 0, nil), "invalid cell reference [0, 0]")
	assert.EqualError(t, f.addTable("sheet1", "", 1, 1, 0, 0, 0, nil), "invalid cell reference [0, 0]")
	// Test add table with invalid table name
	for _, cases := range []struct {
		name string
		err  error
	}{
		{name: "1Table", err: newInvalidTableNameError("1Table")},
		{name: "-Table", err: newInvalidTableNameError("-Table")},
		{name: "'Table", err: newInvalidTableNameError("'Table")},
		{name: "Table 1", err: newInvalidTableNameError("Table 1")},
		{name: "A&B", err: newInvalidTableNameError("A&B")},
		{name: "_1Table'", err: newInvalidTableNameError("_1Table'")},
		{name: "\u0f5f\u0fb3\u0f0b\u0f21", err: newInvalidTableNameError("\u0f5f\u0fb3\u0f0b\u0f21")},
		{name: strings.Repeat("c", MaxFieldLength+1), err: ErrTableNameLength},
	} {
		assert.EqualError(t, f.AddTable("Sheet1", "A1:B2", &TableOptions{
			Name: cases.name,
		}), cases.err.Error())
	}
}

func TestSetTableHeader(t *testing.T) {
	f := NewFile()
	_, err := f.setTableHeader("Sheet1", 1, 0, 1)
	assert.EqualError(t, err, "invalid cell reference [1, 0]")
}

func TestAutoFilter(t *testing.T) {
	outFile := filepath.Join("test", "TestAutoFilter%d.xlsx")
	f, err := prepareTestBook1()
	assert.NoError(t, err)
	for i, opts := range []*AutoFilterOptions{
		nil,
		{Column: "B", Expression: ""},
		{Column: "B", Expression: "x != blanks"},
		{Column: "B", Expression: "x == blanks"},
		{Column: "B", Expression: "x != nonblanks"},
		{Column: "B", Expression: "x == nonblanks"},
		{Column: "B", Expression: "x <= 1 and x >= 2"},
		{Column: "B", Expression: "x == 1 or x == 2"},
		{Column: "B", Expression: "x == 1 or x == 2*"},
	} {
		t.Run(fmt.Sprintf("Expression%d", i+1), func(t *testing.T) {
			assert.NoError(t, f.AutoFilter("Sheet1", "D4:B1", opts))
			assert.NoError(t, f.SaveAs(fmt.Sprintf(outFile, i+1)))
		})
	}

	// Test add auto filter with invalid sheet name
	assert.EqualError(t, f.AutoFilter("Sheet:1", "A1:B1", nil), ErrSheetNameInvalid.Error())
	// Test add auto filter with illegal cell reference
	assert.EqualError(t, f.AutoFilter("Sheet1", "A:B1", nil), newCellNameToCoordinatesError("A", newInvalidCellNameError("A")).Error())
	assert.EqualError(t, f.AutoFilter("Sheet1", "A1:B", nil), newCellNameToCoordinatesError("B", newInvalidCellNameError("B")).Error())
	// Test add auto filter with unsupported charset workbook
	f.WorkBook = nil
	f.Pkg.Store(defaultXMLPathWorkbook, MacintoshCyrillicCharset)
	assert.EqualError(t, f.AutoFilter("Sheet1", "D4:B1", nil), "XML syntax error on line 1: invalid UTF-8")
}

func TestAutoFilterError(t *testing.T) {
	outFile := filepath.Join("test", "TestAutoFilterError%d.xlsx")
	f, err := prepareTestBook1()
	assert.NoError(t, err)
	for i, opts := range []*AutoFilterOptions{
		{Column: "B", Expression: "x <= 1 and x >= blanks"},
		{Column: "B", Expression: "x -- y or x == *2*"},
		{Column: "B", Expression: "x != y or x ? *2"},
		{Column: "B", Expression: "x -- y o r x == *2"},
		{Column: "B", Expression: "x -- y"},
		{Column: "A", Expression: "x -- y"},
	} {
		t.Run(fmt.Sprintf("Expression%d", i+1), func(t *testing.T) {
			if assert.Error(t, f.AutoFilter("Sheet2", "D4:B1", opts)) {
				assert.NoError(t, f.SaveAs(fmt.Sprintf(outFile, i+1)))
			}
		})
	}

	assert.EqualError(t, f.autoFilter("SheetN", "A1", 1, 1, &AutoFilterOptions{
		Column:     "A",
		Expression: "",
	}), "sheet SheetN does not exist")
	assert.EqualError(t, f.autoFilter("Sheet1", "A1", 1, 1, &AutoFilterOptions{
		Column:     "-",
		Expression: "-",
	}), newInvalidColumnNameError("-").Error())
	assert.EqualError(t, f.autoFilter("Sheet1", "A1", 1, 100, &AutoFilterOptions{
		Column:     "A",
		Expression: "-",
	}), `incorrect index of column 'A'`)
	assert.EqualError(t, f.autoFilter("Sheet1", "A1", 1, 1, &AutoFilterOptions{
		Column:     "A",
		Expression: "-",
	}), `incorrect number of tokens in criteria '-'`)
}

func TestParseFilterTokens(t *testing.T) {
	f := NewFile()
	// Test with unknown operator
	_, _, err := f.parseFilterTokens("", []string{"", "!"})
	assert.EqualError(t, err, "unknown operator: !")
	// Test invalid operator in context
	_, _, err = f.parseFilterTokens("", []string{"", "<", "x != blanks"})
	assert.EqualError(t, err, "the operator '<' in expression '' is not valid in relation to Blanks/NonBlanks'")
}
