package excelize

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStyleFill(t *testing.T) {
	cases := []struct {
		label      string
		format     string
		expectFill bool
	}{{
		label:      "no_fill",
		format:     `{"alignment":{"wrap_text":true}}`,
		expectFill: false,
	}, {
		label:      "fill",
		format:     `{"fill":{"type":"pattern","pattern":1,"color":["#000000"]}}`,
		expectFill: true,
	}}

	for _, testCase := range cases {
		xl := NewFile()
		styleID, err := xl.NewStyle(testCase.format)
		assert.NoError(t, err)

		styles, err := xl.stylesReader()
		assert.NoError(t, err)
		style := styles.CellXfs.Xf[styleID]
		if testCase.expectFill {
			assert.NotEqual(t, *style.FillID, 0, testCase.label)
		} else {
			assert.Equal(t, *style.FillID, 0, testCase.label)
		}
	}
	f := NewFile()
	styleID1, err := f.NewStyle(`{"fill":{"type":"pattern","pattern":1,"color":["#000000"]}}`)
	assert.NoError(t, err)
	styleID2, err := f.NewStyle(`{"fill":{"type":"pattern","pattern":1,"color":["#000000"]}}`)
	assert.NoError(t, err)
	assert.Equal(t, styleID1, styleID2)
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestStyleFill.xlsx")))
}

func TestSetConditionalFormat(t *testing.T) {
	cases := []struct {
		label  string
		format string
		rules  []*xlsxCfRule
	}{{
		label: "3_color_scale",
		format: `[{
			"type":"3_color_scale",
			"criteria":"=",
			"min_type":"num",
			"mid_type":"num",
			"max_type":"num",
			"min_value": "-10",
			"mid_value": "0",
			"max_value": "10",
			"min_color":"ff0000",
			"mid_color":"00ff00",
			"max_color":"0000ff"
		}]`,
		rules: []*xlsxCfRule{{
			Priority: 1,
			Type:     "colorScale",
			ColorScale: &xlsxColorScale{
				Cfvo: []*xlsxCfvo{{
					Type: "num",
					Val:  "-10",
				}, {
					Type: "num",
					Val:  "0",
				}, {
					Type: "num",
					Val:  "10",
				}},
				Color: []*xlsxColor{{
					RGB: "FFFF0000",
				}, {
					RGB: "FF00FF00",
				}, {
					RGB: "FF0000FF",
				}},
			},
		}},
	}, {
		label: "3_color_scale default min/mid/max",
		format: `[{
			"type":"3_color_scale",
			"criteria":"=",
			"min_type":"num",
			"mid_type":"num",
			"max_type":"num",
			"min_color":"ff0000",
			"mid_color":"00ff00",
			"max_color":"0000ff"
		}]`,
		rules: []*xlsxCfRule{{
			Priority: 1,
			Type:     "colorScale",
			ColorScale: &xlsxColorScale{
				Cfvo: []*xlsxCfvo{{
					Type: "num",
					Val:  "0",
				}, {
					Type: "num",
					Val:  "50",
				}, {
					Type: "num",
					Val:  "0",
				}},
				Color: []*xlsxColor{{
					RGB: "FFFF0000",
				}, {
					RGB: "FF00FF00",
				}, {
					RGB: "FF0000FF",
				}},
			},
		}},
	}, {
		label: "2_color_scale default min/max",
		format: `[{
			"type":"2_color_scale",
			"criteria":"=",
			"min_type":"num",
			"max_type":"num",
			"min_color":"ff0000",
			"max_color":"0000ff"
		}]`,
		rules: []*xlsxCfRule{{
			Priority: 1,
			Type:     "colorScale",
			ColorScale: &xlsxColorScale{
				Cfvo: []*xlsxCfvo{{
					Type: "num",
					Val:  "0",
				}, {
					Type: "num",
					Val:  "0",
				}},
				Color: []*xlsxColor{{
					RGB: "FFFF0000",
				}, {
					RGB: "FF0000FF",
				}},
			},
		}},
	}}

	for _, testCase := range cases {
		f := NewFile()
		const sheet = "Sheet1"
		const cellRange = "A1:A1"

		err := f.SetConditionalFormat(sheet, cellRange, testCase.format)
		if err != nil {
			t.Fatalf("%s", err)
		}

		ws, err := f.workSheetReader(sheet)
		assert.NoError(t, err)
		cf := ws.ConditionalFormatting
		assert.Len(t, cf, 1, testCase.label)
		assert.Len(t, cf[0].CfRule, 1, testCase.label)
		assert.Equal(t, cellRange, cf[0].SQRef, testCase.label)
		assert.EqualValues(t, testCase.rules, cf[0].CfRule, testCase.label)
	}
}

func TestGetConditionalFormats(t *testing.T) {
	for _, format := range []string{
		`[{"type":"cell","format":1,"criteria":"greater than","value":"6"}]`,
		`[{"type":"cell","format":1,"criteria":"between","minimum":"6","maximum":"8"}]`,
		`[{"type":"top","format":1,"criteria":"=","value":"6"}]`,
		`[{"type":"bottom","format":1,"criteria":"=","value":"6"}]`,
		`[{"type":"average","above_average":true,"format":1,"criteria":"="}]`,
		`[{"type":"duplicate","format":1,"criteria":"="}]`,
		`[{"type":"unique","format":1,"criteria":"="}]`,
		`[{"type":"3_color_scale","criteria":"=","min_type":"num","mid_type":"num","max_type":"num","min_value":"-10","mid_value":"50","max_value":"10","min_color":"#FF0000","mid_color":"#00FF00","max_color":"#0000FF"}]`,
		`[{"type":"2_color_scale","criteria":"=","min_type":"num","max_type":"num","min_color":"#FF0000","max_color":"#0000FF"}]`,
		`[{"type":"data_bar","criteria":"=","min_type":"min","max_type":"max","bar_color":"#638EC6"}]`,
		`[{"type":"formula","format":1,"criteria":"="}]`,
	} {
		f := NewFile()
		err := f.SetConditionalFormat("Sheet1", "A1:A2", format)
		assert.NoError(t, err)
		opts, err := f.GetConditionalFormats("Sheet1")
		assert.NoError(t, err)
		assert.Equal(t, format, opts["A1:A2"])
	}
	// Test get conditional formats on no exists worksheet
	f := NewFile()
	_, err := f.GetConditionalFormats("SheetN")
	assert.EqualError(t, err, "sheet SheetN does not exist")
	// Test get conditional formats with invalid sheet name
	_, err = f.GetConditionalFormats("Sheet:1")
	assert.EqualError(t, err, ErrSheetNameInvalid.Error())
}

func TestUnsetConditionalFormat(t *testing.T) {
	f := NewFile()
	assert.NoError(t, f.SetCellValue("Sheet1", "A1", 7))
	assert.NoError(t, f.UnsetConditionalFormat("Sheet1", "A1:A10"))
	format, err := f.NewConditionalStyle(`{"font":{"color":"#9A0511"},"fill":{"type":"pattern","color":["#FEC7CE"],"pattern":1}}`)
	assert.NoError(t, err)
	assert.NoError(t, f.SetConditionalFormat("Sheet1", "A1:A10", fmt.Sprintf(`[{"type":"cell","criteria":">","format":%d,"value":"6"}]`, format)))
	assert.NoError(t, f.UnsetConditionalFormat("Sheet1", "A1:A10"))
	// Test unset conditional format on not exists worksheet
	assert.EqualError(t, f.UnsetConditionalFormat("SheetN", "A1:A10"), "sheet SheetN does not exist")
	// Test unset conditional format with invalid sheet name
	assert.EqualError(t, f.UnsetConditionalFormat("Sheet:1", "A1:A10"), ErrSheetNameInvalid.Error())
	// Save spreadsheet by the given path
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestUnsetConditionalFormat.xlsx")))
}

func TestNewStyle(t *testing.T) {
	f := NewFile()
	styleID, err := f.NewStyle(`{"font":{"bold":true,"italic":true,"family":"Times New Roman","size":36,"color":"#777777"}}`)
	assert.NoError(t, err)
	styles, err := f.stylesReader()
	assert.NoError(t, err)
	fontID := styles.CellXfs.Xf[styleID].FontID
	font := styles.Fonts.Font[*fontID]
	assert.Contains(t, *font.Name.Val, "Times New Roman", "Stored font should contain font name")
	assert.Equal(t, 2, styles.CellXfs.Count, "Should have 2 styles")
	_, err = f.NewStyle(&Style{})
	assert.NoError(t, err)
	_, err = f.NewStyle(Style{})
	assert.EqualError(t, err, ErrParameterInvalid.Error())

	var exp string
	_, err = f.NewStyle(&Style{CustomNumFmt: &exp})
	assert.EqualError(t, err, ErrCustomNumFmt.Error())
	_, err = f.NewStyle(&Style{Font: &Font{Family: strings.Repeat("s", MaxFontFamilyLength+1)}})
	assert.EqualError(t, err, ErrFontLength.Error())
	_, err = f.NewStyle(&Style{Font: &Font{Size: MaxFontSize + 1}})
	assert.EqualError(t, err, ErrFontSize.Error())

	// Test create numeric custom style
	numFmt := "####;####"
	f.Styles.NumFmts = nil
	styleID, err = f.NewStyle(&Style{
		CustomNumFmt: &numFmt,
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, styleID)

	assert.NotNil(t, f.Styles)
	assert.NotNil(t, f.Styles.CellXfs)
	assert.NotNil(t, f.Styles.CellXfs.Xf)

	nf := f.Styles.CellXfs.Xf[styleID]
	assert.Equal(t, 164, *nf.NumFmtID)

	// Test create currency custom style
	f.Styles.NumFmts = nil
	styleID, err = f.NewStyle(&Style{
		Lang:   "ko-kr",
		NumFmt: 32, // must not be in currencyNumFmt

	})
	assert.NoError(t, err)
	assert.Equal(t, 3, styleID)

	assert.NotNil(t, f.Styles)
	assert.NotNil(t, f.Styles.CellXfs)
	assert.NotNil(t, f.Styles.CellXfs.Xf)

	nf = f.Styles.CellXfs.Xf[styleID]
	assert.Equal(t, 32, *nf.NumFmtID)

	// Test set build-in scientific number format
	styleID, err = f.NewStyle(&Style{NumFmt: 11})
	assert.NoError(t, err)
	assert.NoError(t, f.SetCellStyle("Sheet1", "A1", "B1", styleID))
	assert.NoError(t, f.SetSheetRow("Sheet1", "A1", &[]float64{1.23, 1.234}))
	rows, err := f.GetRows("Sheet1")
	assert.NoError(t, err)
	assert.Equal(t, [][]string{{"1.23E+00", "1.23E+00"}}, rows)

	f = NewFile()
	// Test currency number format
	customNumFmt := "[$$-409]#,##0.00"
	style1, err := f.NewStyle(&Style{CustomNumFmt: &customNumFmt})
	assert.NoError(t, err)
	style2, err := f.NewStyle(&Style{NumFmt: 165})
	assert.NoError(t, err)
	assert.Equal(t, style1, style2)

	style3, err := f.NewStyle(&Style{NumFmt: 166})
	assert.NoError(t, err)
	assert.Equal(t, 2, style3)

	f = NewFile()
	f.Styles.NumFmts = nil
	f.Styles.CellXfs.Xf = nil
	style4, err := f.NewStyle(&Style{NumFmt: 160, Lang: "unknown"})
	assert.NoError(t, err)
	assert.Equal(t, 0, style4)

	f = NewFile()
	f.Styles.NumFmts = nil
	f.Styles.CellXfs.Xf = nil
	style5, err := f.NewStyle(&Style{NumFmt: 160, Lang: "zh-cn"})
	assert.NoError(t, err)
	assert.Equal(t, 0, style5)

	// Test create style with unsupported charset style sheet
	f.Styles = nil
	f.Pkg.Store(defaultXMLPathStyles, MacintoshCyrillicCharset)
	_, err = f.NewStyle(&Style{NumFmt: 165})
	assert.EqualError(t, err, "XML syntax error on line 1: invalid UTF-8")
}

func TestNewConditionalStyle(t *testing.T) {
	f := NewFile()
	// Test create conditional style with unsupported charset style sheet
	f.Styles = nil
	f.Pkg.Store(defaultXMLPathStyles, MacintoshCyrillicCharset)
	_, err := f.NewConditionalStyle(`{"font":{"color":"#9A0511"},"fill":{"type":"pattern","color":["#FEC7CE"],"pattern":1}}`)
	assert.EqualError(t, err, "XML syntax error on line 1: invalid UTF-8")
}

func TestGetDefaultFont(t *testing.T) {
	f := NewFile()
	s, err := f.GetDefaultFont()
	assert.NoError(t, err)
	assert.Equal(t, s, "Calibri", "Default font should be Calibri")
	// Test get default font with unsupported charset style sheet
	f.Styles = nil
	f.Pkg.Store(defaultXMLPathStyles, MacintoshCyrillicCharset)
	_, err = f.GetDefaultFont()
	assert.EqualError(t, err, "XML syntax error on line 1: invalid UTF-8")
}

func TestSetDefaultFont(t *testing.T) {
	f := NewFile()
	assert.NoError(t, f.SetDefaultFont("Arial"))
	styles, err := f.stylesReader()
	assert.NoError(t, err)
	s, err := f.GetDefaultFont()
	assert.NoError(t, err)
	assert.Equal(t, s, "Arial", "Default font should change to Arial")
	assert.Equal(t, *styles.CellStyles.CellStyle[0].CustomBuiltIn, true)
	// Test set default font with unsupported charset style sheet
	f.Styles = nil
	f.Pkg.Store(defaultXMLPathStyles, MacintoshCyrillicCharset)
	assert.EqualError(t, f.SetDefaultFont("Arial"), "XML syntax error on line 1: invalid UTF-8")
}

func TestStylesReader(t *testing.T) {
	f := NewFile()
	// Test read styles with unsupported charset
	f.Styles = nil
	f.Pkg.Store(defaultXMLPathStyles, MacintoshCyrillicCharset)
	styles, err := f.stylesReader()
	assert.EqualError(t, err, "XML syntax error on line 1: invalid UTF-8")
	assert.EqualValues(t, new(xlsxStyleSheet), styles)
}

func TestThemeReader(t *testing.T) {
	f := NewFile()
	// Test read theme with unsupported charset
	f.Pkg.Store(defaultXMLPathTheme, MacintoshCyrillicCharset)
	theme, err := f.themeReader()
	assert.EqualError(t, err, "XML syntax error on line 1: invalid UTF-8")
	assert.EqualValues(t, &xlsxTheme{XMLNSa: NameSpaceDrawingML.Value, XMLNSr: SourceRelationship.Value}, theme)
}

func TestSetCellStyle(t *testing.T) {
	f := NewFile()
	// Test set cell style on not exists worksheet.
	assert.EqualError(t, f.SetCellStyle("SheetN", "A1", "A2", 1), "sheet SheetN does not exist")
	// Test set cell style with invalid style ID.
	assert.EqualError(t, f.SetCellStyle("Sheet1", "A1", "A2", -1), newInvalidStyleID(-1).Error())
	// Test set cell style with not exists style ID.
	assert.EqualError(t, f.SetCellStyle("Sheet1", "A1", "A2", 10), newInvalidStyleID(10).Error())
	// Test set cell style with unsupported charset style sheet.
	f.Styles = nil
	f.Pkg.Store(defaultXMLPathStyles, MacintoshCyrillicCharset)
	assert.EqualError(t, f.SetCellStyle("Sheet1", "A1", "A2", 1), "XML syntax error on line 1: invalid UTF-8")
}

func TestGetStyleID(t *testing.T) {
	f := NewFile()
	styleID, err := f.getStyleID(&xlsxStyleSheet{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, -1, styleID)
	// Test get style ID with unsupported charset style sheet.
	f.Styles = nil
	f.Pkg.Store(defaultXMLPathStyles, MacintoshCyrillicCharset)
	_, err = f.getStyleID(&xlsxStyleSheet{
		CellXfs: &xlsxCellXfs{},
		Fonts: &xlsxFonts{
			Font: []*xlsxFont{{}},
		},
	}, &Style{NumFmt: 0, Font: &Font{}})
	assert.EqualError(t, err, "XML syntax error on line 1: invalid UTF-8")
}

func TestGetFillID(t *testing.T) {
	styles, err := NewFile().stylesReader()
	assert.NoError(t, err)
	assert.Equal(t, -1, getFillID(styles, &Style{Fill: Fill{Type: "unknown"}}))
}

func TestThemeColor(t *testing.T) {
	for _, clr := range [][]string{
		{"FF000000", ThemeColor("000000", -0.1)},
		{"FF000000", ThemeColor("000000", 0)},
		{"FF33FF33", ThemeColor("00FF00", 0.2)},
		{"FFFFFFFF", ThemeColor("000000", 1)},
		{"FFFFFFFF", ThemeColor(strings.Repeat(string(rune(math.MaxUint8+1)), 6), 1)},
		{"FFFFFFFF", ThemeColor(strings.Repeat(string(rune(-1)), 6), 1)},
	} {
		assert.Equal(t, clr[0], clr[1])
	}
}

func TestGetNumFmtID(t *testing.T) {
	f := NewFile()

	fs1, err := parseFormatStyleSet(`{"protection":{"hidden":false,"locked":false},"number_format":10}`)
	assert.NoError(t, err)
	id1 := getNumFmtID(&xlsxStyleSheet{}, fs1)

	fs2, err := parseFormatStyleSet(`{"protection":{"hidden":false,"locked":false},"number_format":0}`)
	assert.NoError(t, err)
	id2 := getNumFmtID(&xlsxStyleSheet{}, fs2)

	assert.NotEqual(t, id1, id2)
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestStyleNumFmt.xlsx")))
}
