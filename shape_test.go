package excelize

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddShape(t *testing.T) {
	f, err := prepareTestBook1()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.NoError(t, f.AddShape("Sheet1", "A30", `{"type":"rect","paragraph":[{"text":"Rectangle","font":{"color":"CD5C5C"}},{"text":"Shape","font":{"bold":true,"color":"2980B9"}}]}`))
	assert.NoError(t, f.AddShape("Sheet1", "B30", `{"type":"rect","paragraph":[{"text":"Rectangle"},{}]}`))
	assert.NoError(t, f.AddShape("Sheet1", "C30", `{"type":"rect","paragraph":[]}`))
	assert.EqualError(t, f.AddShape("Sheet3", "H1", `{
		"type": "ellipseRibbon",
		"color":
		{
			"line": "#4286f4",
			"fill": "#8eb9ff"
		},
		"paragraph": [
		{
			"font":
			{
				"bold": true,
				"italic": true,
				"family": "Times New Roman",
				"size": 36,
				"color": "#777777",
				"underline": "single"
			}
		}],
		"height": 90
	}`), "sheet Sheet3 does not exist")
	assert.EqualError(t, f.AddShape("Sheet3", "H1", ""), "unexpected end of JSON input")
	assert.EqualError(t, f.AddShape("Sheet1", "A", `{
		"type": "rect",
		"paragraph": [
		{
			"text": "Rectangle",
			"font":
			{
				"color": "CD5C5C"
			}
		},
		{
			"text": "Shape",
			"font":
			{
				"bold": true,
				"color": "2980B9"
			}
		}]
	}`), newCellNameToCoordinatesError("A", newInvalidCellNameError("A")).Error())
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestAddShape1.xlsx")))

	// Test add first shape for given sheet
	f = NewFile()
	assert.NoError(t, f.AddShape("Sheet1", "A1", `{
		"type": "ellipseRibbon",
		"color":
		{
			"line": "#4286f4",
			"fill": "#8eb9ff"
		},
		"paragraph": [
		{
			"font":
			{
				"bold": true,
				"italic": true,
				"family": "Times New Roman",
				"size": 36,
				"color": "#777777",
				"underline": "single"
			}
		}],
		"height": 90,
		"line":
		{
			"width": 1.2
		}
	}`))
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestAddShape2.xlsx")))
	// Test add shape with invalid sheet name
	assert.EqualError(t, f.AddShape("Sheet:1", "A30", `{"type":"rect","paragraph":[{"text":"Rectangle","font":{"color":"CD5C5C"}},{"text":"Shape","font":{"bold":true,"color":"2980B9"}}]}`), ErrSheetNameInvalid.Error())
	// Test add shape with unsupported charset style sheet
	f.Styles = nil
	f.Pkg.Store(defaultXMLPathStyles, MacintoshCyrillicCharset)
	assert.EqualError(t, f.AddShape("Sheet1", "B30", `{"type":"rect","paragraph":[{"text":"Rectangle"},{}]}`), "XML syntax error on line 1: invalid UTF-8")
	// Test add shape with unsupported charset content types
	f = NewFile()
	f.ContentTypes = nil
	f.Pkg.Store(defaultXMLPathContentTypes, MacintoshCyrillicCharset)
	assert.EqualError(t, f.AddShape("Sheet1", "B30", `{"type":"rect","paragraph":[{"text":"Rectangle"},{}]}`), "XML syntax error on line 1: invalid UTF-8")
}

func TestAddDrawingShape(t *testing.T) {
	f := NewFile()
	path := "xl/drawings/drawing1.xml"
	f.Pkg.Store(path, MacintoshCyrillicCharset)
	assert.EqualError(t, f.addDrawingShape("sheet1", path, "A1", &shapeOptions{}), "XML syntax error on line 1: invalid UTF-8")
}
