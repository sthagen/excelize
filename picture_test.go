package excelize

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "golang.org/x/image/tiff"

	"github.com/stretchr/testify/assert"
)

func BenchmarkAddPictureFromBytes(b *testing.B) {
	f := NewFile()
	imgFile, err := os.ReadFile(filepath.Join("test", "images", "excel.png"))
	if err != nil {
		b.Error("unable to load image for benchmark")
	}
	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		if err := f.AddPictureFromBytes("Sheet1", fmt.Sprint("A", i), "", "excel", ".png", imgFile); err != nil {
			b.Error(err)
		}
	}
}

func TestAddPicture(t *testing.T) {
	f, err := OpenFile(filepath.Join("test", "Book1.xlsx"))
	assert.NoError(t, err)

	// Test add picture to worksheet with offset and location hyperlink
	assert.NoError(t, f.AddPicture("Sheet2", "I9", filepath.Join("test", "images", "excel.jpg"),
		`{"x_offset": 140, "y_offset": 120, "hyperlink": "#Sheet2!D8", "hyperlink_type": "Location"}`))
	// Test add picture to worksheet with offset, external hyperlink and positioning
	assert.NoError(t, f.AddPicture("Sheet1", "F21", filepath.Join("test", "images", "excel.jpg"),
		`{"x_offset": 10, "y_offset": 10, "hyperlink": "https://github.com/xuri/excelize", "hyperlink_type": "External", "positioning": "oneCell"}`))

	file, err := os.ReadFile(filepath.Join("test", "images", "excel.png"))
	assert.NoError(t, err)

	// Test add picture to worksheet with autofit
	assert.NoError(t, f.AddPicture("Sheet1", "A30", filepath.Join("test", "images", "excel.jpg"), `{"autofit": true}`))
	assert.NoError(t, f.AddPicture("Sheet1", "B30", filepath.Join("test", "images", "excel.jpg"), `{"x_offset": 10, "y_offset": 10, "autofit": true}`))
	f.NewSheet("AddPicture")
	assert.NoError(t, f.SetRowHeight("AddPicture", 10, 30))
	assert.NoError(t, f.MergeCell("AddPicture", "B3", "D9"))
	assert.NoError(t, f.MergeCell("AddPicture", "B1", "D1"))
	assert.NoError(t, f.AddPicture("AddPicture", "C6", filepath.Join("test", "images", "excel.jpg"), `{"autofit": true}`))
	assert.NoError(t, f.AddPicture("AddPicture", "A1", filepath.Join("test", "images", "excel.jpg"), `{"autofit": true}`))

	// Test add picture to worksheet from bytes
	assert.NoError(t, f.AddPictureFromBytes("Sheet1", "Q1", "", "Excel Logo", ".png", file))
	// Test add picture to worksheet from bytes with illegal cell reference
	assert.EqualError(t, f.AddPictureFromBytes("Sheet1", "A", "", "Excel Logo", ".png", file), newCellNameToCoordinatesError("A", newInvalidCellNameError("A")).Error())

	assert.NoError(t, f.AddPicture("Sheet1", "Q8", filepath.Join("test", "images", "excel.gif"), ""))
	assert.NoError(t, f.AddPicture("Sheet1", "Q15", filepath.Join("test", "images", "excel.jpg"), ""))
	assert.NoError(t, f.AddPicture("Sheet1", "Q22", filepath.Join("test", "images", "excel.tif"), ""))

	// Test write file to given path
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestAddPicture1.xlsx")))
	assert.NoError(t, f.Close())

	// Test add picture with unsupported charset content types
	f = NewFile()
	f.ContentTypes = nil
	f.Pkg.Store(defaultXMLPathContentTypes, MacintoshCyrillicCharset)
	assert.EqualError(t, f.AddPictureFromBytes("Sheet1", "Q1", "", "Excel Logo", ".png", file), "XML syntax error on line 1: invalid UTF-8")

	// Test add picture with invalid sheet name
	assert.EqualError(t, f.AddPicture("Sheet:1", "A1", filepath.Join("test", "images", "excel.jpg"), ""), ErrSheetNameInvalid.Error())
}

func TestAddPictureErrors(t *testing.T) {
	f, err := OpenFile(filepath.Join("test", "Book1.xlsx"))
	assert.NoError(t, err)

	// Test add picture to worksheet with invalid file path
	assert.Error(t, f.AddPicture("Sheet1", "G21", filepath.Join("test", "not_exists_dir", "not_exists.icon"), ""))

	// Test add picture to worksheet with unsupported file type
	assert.EqualError(t, f.AddPicture("Sheet1", "G21", filepath.Join("test", "Book1.xlsx"), ""), ErrImgExt.Error())
	assert.EqualError(t, f.AddPictureFromBytes("Sheet1", "G21", "", "Excel Logo", "jpg", make([]byte, 1)), ErrImgExt.Error())

	// Test add picture to worksheet with invalid file data
	assert.EqualError(t, f.AddPictureFromBytes("Sheet1", "G21", "", "Excel Logo", ".jpg", make([]byte, 1)), image.ErrFormat.Error())

	// Test add picture with custom image decoder and encoder
	decode := func(r io.Reader) (image.Image, error) { return nil, nil }
	decodeConfig := func(r io.Reader) (image.Config, error) { return image.Config{Height: 100, Width: 90}, nil }
	image.RegisterFormat("emf", "", decode, decodeConfig)
	image.RegisterFormat("wmf", "", decode, decodeConfig)
	image.RegisterFormat("emz", "", decode, decodeConfig)
	image.RegisterFormat("wmz", "", decode, decodeConfig)
	image.RegisterFormat("svg", "", decode, decodeConfig)
	assert.NoError(t, f.AddPicture("Sheet1", "Q1", filepath.Join("test", "images", "excel.emf"), ""))
	assert.NoError(t, f.AddPicture("Sheet1", "Q7", filepath.Join("test", "images", "excel.wmf"), ""))
	assert.NoError(t, f.AddPicture("Sheet1", "Q13", filepath.Join("test", "images", "excel.emz"), ""))
	assert.NoError(t, f.AddPicture("Sheet1", "Q19", filepath.Join("test", "images", "excel.wmz"), ""))
	assert.NoError(t, f.AddPicture("Sheet1", "Q25", "excelize.svg", `{"x_scale": 2.1}`))
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestAddPicture2.xlsx")))
	assert.NoError(t, f.Close())
}

func TestGetPicture(t *testing.T) {
	f := NewFile()
	assert.NoError(t, f.AddPicture("Sheet1", "A1", filepath.Join("test", "images", "excel.png"), ""))
	name, content, err := f.GetPicture("Sheet1", "A1")
	assert.NoError(t, err)
	assert.Equal(t, 13233, len(content))
	assert.Equal(t, "image1.png", name)

	f, err = prepareTestBook1()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	file, raw, err := f.GetPicture("Sheet1", "F21")
	assert.NoError(t, err)
	if !assert.NotEmpty(t, filepath.Join("test", file)) || !assert.NotEmpty(t, raw) ||
		!assert.NoError(t, os.WriteFile(filepath.Join("test", file), raw, 0o644)) {
		t.FailNow()
	}

	// Try to get picture from a worksheet with illegal cell reference
	_, _, err = f.GetPicture("Sheet1", "A")
	assert.EqualError(t, err, newCellNameToCoordinatesError("A", newInvalidCellNameError("A")).Error())

	// Try to get picture from a worksheet that doesn't contain any images
	file, raw, err = f.GetPicture("Sheet3", "I9")
	assert.EqualError(t, err, "sheet Sheet3 does not exist")
	assert.Empty(t, file)
	assert.Empty(t, raw)

	// Try to get picture from a cell that doesn't contain an image
	file, raw, err = f.GetPicture("Sheet2", "A2")
	assert.NoError(t, err)
	assert.Empty(t, file)
	assert.Empty(t, raw)

	// Test get picture with invalid sheet name
	_, _, err = f.GetPicture("Sheet:1", "A2")
	assert.EqualError(t, err, ErrSheetNameInvalid.Error())

	f.getDrawingRelationships("xl/worksheets/_rels/sheet1.xml.rels", "rId8")
	f.getDrawingRelationships("", "")
	f.getSheetRelationshipsTargetByID("", "")
	f.deleteSheetRelationships("", "")

	// Try to get picture from a local storage file.
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestGetPicture.xlsx")))

	f, err = OpenFile(filepath.Join("test", "TestGetPicture.xlsx"))
	assert.NoError(t, err)

	file, raw, err = f.GetPicture("Sheet1", "F21")
	assert.NoError(t, err)
	if !assert.NotEmpty(t, filepath.Join("test", file)) || !assert.NotEmpty(t, raw) ||
		!assert.NoError(t, os.WriteFile(filepath.Join("test", file), raw, 0o644)) {
		t.FailNow()
	}

	// Try to get picture from a local storage file that doesn't contain an image
	file, raw, err = f.GetPicture("Sheet1", "F22")
	assert.NoError(t, err)
	assert.Empty(t, file)
	assert.Empty(t, raw)
	assert.NoError(t, f.Close())

	// Test get picture from none drawing worksheet
	f = NewFile()
	file, raw, err = f.GetPicture("Sheet1", "F22")
	assert.NoError(t, err)
	assert.Empty(t, file)
	assert.Empty(t, raw)
	f, err = prepareTestBook1()
	assert.NoError(t, err)

	// Test get pictures with unsupported charset
	path := "xl/drawings/drawing1.xml"
	f.Pkg.Store(path, MacintoshCyrillicCharset)
	_, _, err = f.getPicture(20, 5, path, "xl/drawings/_rels/drawing2.xml.rels")
	assert.EqualError(t, err, "XML syntax error on line 1: invalid UTF-8")
	f.Drawings.Delete(path)
	_, _, err = f.getPicture(20, 5, path, "xl/drawings/_rels/drawing2.xml.rels")
	assert.EqualError(t, err, "XML syntax error on line 1: invalid UTF-8")
}

func TestAddDrawingPicture(t *testing.T) {
	// Test addDrawingPicture with illegal cell reference
	f := NewFile()
	assert.EqualError(t, f.addDrawingPicture("sheet1", "", "A", "", "", 0, 0, image.Config{}, nil), newCellNameToCoordinatesError("A", newInvalidCellNameError("A")).Error())

	path := "xl/drawings/drawing1.xml"
	f.Pkg.Store(path, MacintoshCyrillicCharset)
	assert.EqualError(t, f.addDrawingPicture("sheet1", path, "A1", "", "", 0, 0, image.Config{}, &pictureOptions{}), "XML syntax error on line 1: invalid UTF-8")
}

func TestAddPictureFromBytes(t *testing.T) {
	f := NewFile()
	imgFile, err := os.ReadFile("logo.png")
	assert.NoError(t, err, "Unable to load logo for test")
	assert.NoError(t, f.AddPictureFromBytes("Sheet1", fmt.Sprint("A", 1), "", "logo", ".png", imgFile))
	assert.NoError(t, f.AddPictureFromBytes("Sheet1", fmt.Sprint("A", 50), "", "logo", ".png", imgFile))
	imageCount := 0
	f.Pkg.Range(func(fileName, v interface{}) bool {
		if strings.Contains(fileName.(string), "media/image") {
			imageCount++
		}
		return true
	})
	assert.Equal(t, 1, imageCount, "Duplicate image should only be stored once.")
	assert.EqualError(t, f.AddPictureFromBytes("SheetN", fmt.Sprint("A", 1), "", "logo", ".png", imgFile), "sheet SheetN does not exist")
	// Test add picture from bytes with invalid sheet name
	assert.EqualError(t, f.AddPictureFromBytes("Sheet:1", fmt.Sprint("A", 1), "", "logo", ".png", imgFile), ErrSheetNameInvalid.Error())
}

func TestDeletePicture(t *testing.T) {
	f, err := OpenFile(filepath.Join("test", "Book1.xlsx"))
	assert.NoError(t, err)
	assert.NoError(t, f.DeletePicture("Sheet1", "A1"))
	assert.NoError(t, f.AddPicture("Sheet1", "P1", filepath.Join("test", "images", "excel.jpg"), ""))
	assert.NoError(t, f.DeletePicture("Sheet1", "P1"))
	assert.NoError(t, f.SaveAs(filepath.Join("test", "TestDeletePicture.xlsx")))
	// Test delete picture on not exists worksheet
	assert.EqualError(t, f.DeletePicture("SheetN", "A1"), "sheet SheetN does not exist")
	// Test delete picture with invalid sheet name
	assert.EqualError(t, f.DeletePicture("Sheet:1", "A1"), ErrSheetNameInvalid.Error())
	// Test delete picture with invalid coordinates
	assert.EqualError(t, f.DeletePicture("Sheet1", ""), newCellNameToCoordinatesError("", newInvalidCellNameError("")).Error())
	assert.NoError(t, f.Close())
	// Test delete picture on no chart worksheet
	assert.NoError(t, NewFile().DeletePicture("Sheet1", "A1"))
}

func TestDrawingResize(t *testing.T) {
	f := NewFile()
	// Test calculate drawing resize on not exists worksheet
	_, _, _, _, err := f.drawingResize("SheetN", "A1", 1, 1, nil)
	assert.EqualError(t, err, "sheet SheetN does not exist")
	// Test calculate drawing resize with invalid coordinates
	_, _, _, _, err = f.drawingResize("Sheet1", "", 1, 1, nil)
	assert.EqualError(t, err, newCellNameToCoordinatesError("", newInvalidCellNameError("")).Error())
	ws, ok := f.Sheet.Load("xl/worksheets/sheet1.xml")
	assert.True(t, ok)
	ws.(*xlsxWorksheet).MergeCells = &xlsxMergeCells{Cells: []*xlsxMergeCell{{Ref: "A:A"}}}
	assert.EqualError(t, f.AddPicture("Sheet1", "A1", filepath.Join("test", "images", "excel.jpg"), `{"autofit": true}`), newCellNameToCoordinatesError("A", newInvalidCellNameError("A")).Error())
}

func TestSetContentTypePartImageExtensions(t *testing.T) {
	f := NewFile()
	// Test set content type part image extensions with unsupported charset content types
	f.ContentTypes = nil
	f.Pkg.Store(defaultXMLPathContentTypes, MacintoshCyrillicCharset)
	assert.EqualError(t, f.setContentTypePartImageExtensions(), "XML syntax error on line 1: invalid UTF-8")
}

func TestSetContentTypePartVMLExtensions(t *testing.T) {
	f := NewFile()
	// Test set content type part VML extensions with unsupported charset content types
	f.ContentTypes = nil
	f.Pkg.Store(defaultXMLPathContentTypes, MacintoshCyrillicCharset)
	assert.EqualError(t, f.setContentTypePartVMLExtensions(), "XML syntax error on line 1: invalid UTF-8")
}

func TestAddContentTypePart(t *testing.T) {
	f := NewFile()
	// Test add content type part with unsupported charset content types
	f.ContentTypes = nil
	f.Pkg.Store(defaultXMLPathContentTypes, MacintoshCyrillicCharset)
	assert.EqualError(t, f.addContentTypePart(0, "unknown"), "XML syntax error on line 1: invalid UTF-8")
}
