package pdf

import (
	"github.com/jung-kurt/gofpdf"
	"math"
	"os"
)

const (
	fontPathBold    = "./font/PingFang-SC-Bold.ttf"
	fontPathRegular = "./font/PingFang-SC-Regular.ttf"
	fontPathLight   = "./font/PingFang-SC-Light.ttf"
)

func generatePdf(name, baseInfo, content string, images []string) (pdfPath string, err error) {
	//删除图片
	defer func() {
		for _, image := range images {
			os.Remove(image)
		}
	}()
	pdf := gofpdf.New("P", "mm", "A4", "")
	//生成封面
	pdf.AddPage()

	// 读取图像文件
	pdf.ImageOptions(
		"./static/base.png",
		0, 0,
		210, 0,
		false,
		gofpdf.ImageOptions{ImageType: "png", ReadDpi: true},
		0,
		"",
	)

	pdf.AddUTF8Font("NotoSansSC", "", fontPathLight)
	pdf.SetFont("NotoSansSC", "", 10)

	pdf.SetXY(22, 207)
	pdf.SetTextColor(255, 255, 255)
	pdf.MultiCell(0, 6, baseInfo, "", "", false)

	addPdfContent(pdf, "内容1", content)

	//添加图片
	for _, image := range images {
		pdf.AddPage()
		setPdfLogo(pdf)
		pdf.ImageOptions(
			image,
			12, 23,
			180, 0,
			false,
			gofpdf.ImageOptions{ImageType: "png", ReadDpi: true},
			0,
			"",
		)
	}
	pdfPath = "./temp-files/" + name + ".pdf"
	err = pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return
	}
	return
}

func addPdfContent(pdf *gofpdf.Fpdf, title, content string) {
	pdf.AddUTF8Font("NotoSansSC", "", fontPathBold)
	pdf.SetFont("NotoSansSC", "", 10)

	pdf.AddPage()
	setPdfLogo(pdf)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFontSize(20)
	pageWidth, _ := pdf.GetPageSize()
	leftWidth, _, rightWidth, _ := pdf.GetMargins()
	//textWidth := pdf.GetStringWidth(title)
	//x := (pageWidth - leftWidth - rightWidth - textWidth) / 2

	pdf.SetXY(10, 10)
	pdf.MultiCell(0, 5, title, "", "", false)
	pdf.AddUTF8Font("NotoSansSC", "", fontPathLight)
	pdf.SetFont("NotoSansSC", "", 12)

	pdf.SetFontSize(12)
	_, fontUnitSize := pdf.GetFontSize()
	pageWidth, _ = pdf.GetPageSize()
	leftWidth, _, rightWidth, _ = pdf.GetMargins()
	num := math.Floor((pageWidth - leftWidth - rightWidth) / fontUnitSize)
	content = CovertMultilineStr(content, int(num))
	pdf.SetXY(10, 23)
	pdf.MultiCell(0, 5, content, "", "", false)
}

func setPdfLogo(pdf *gofpdf.Fpdf) {
	// 读取图像文件
	pdf.ImageOptions(
		"./static/logo.png",
		180, 9,
		22, 0,
		false,
		gofpdf.ImageOptions{ImageType: "png", ReadDpi: true},
		0,
		"",
	)
}

func SplitByN(s string, n int) []string {
	var result []string
	runes := []rune(s) // 将字符串转换为rune切片
	j := 0
	count := 0
	for i := 0; i < len(runes); i++ {
		count++
		if string(runes[i]) == "\n" || count == n || i == len(runes)-1 {
			result = append(result, string(runes[j:i+1]))
			j = i + 1
			count = 0
		}
	}
	return result
}

func CovertMultilineStr(str string, num int) string {
	textLines := SplitByN(str, num)
	newTxtStr := ""
	for _, textLine := range textLines {
		if string(textLine[len(textLine)-1]) == "\n" {
			newTxtStr += textLine
		} else {
			newTxtStr += textLine + "\n"
		}
	}
	return newTxtStr
}
