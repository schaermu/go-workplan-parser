package interpreter

import (
	"image"
	"image/color"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

const SCHEDULE_OFFSET_X = 980
const SCHEDULE_OFFSET_Y = -30
const SCHEDULE_WIDTH = 2780
const SCHEDULE_HEIGHT = 95
const SCHEDULE_PADDING = 20

var MONTH_LIST = []string{
	"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli",
	"August", "September", "Oktober", "November", "Dezember",
}

type Interpreter struct {
	plan string
}

func New(planImage string) Interpreter {
	return Interpreter{plan: planImage}
}

func (i *Interpreter) GetSearchVector(needle string) (x int, y int, month int, year int) {
	// create an ocr readable file from the original page (convert to b/w, gauss, binarize using threshold)
	ocrImageFile := i.getNewFilename("ocr")
	mat := gocv.IMRead(i.plan, gocv.IMReadAnyColor)
	grayMat := gocv.NewMat()
	gocv.CvtColor(mat, &grayMat, gocv.ColorBGRAToGray)
	gaussMat := gocv.NewMat()
	gocv.GaussianBlur(grayMat, &gaussMat, image.Point{X: 5, Y: 5}, 0, 0, gocv.BorderDefault)
	threshMat := gocv.NewMat()
	gocv.Threshold(gaussMat, &threshMat, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)
	gocv.IMWrite(ocrImageFile, threshMat)

	// remove distracting lines
	removeLine(&threshMat, image.Point{X: 80, Y: 1}, color.RGBA{R: 0, G: 0, B: 0})
	removeLine(&threshMat, image.Point{X: 1, Y: 80}, color.RGBA{R: 0, G: 0, B: 0})

	// configure text detection
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 25, Y: 9})
	dilatedMat := gocv.NewMat()
	gocv.Dilate(threshMat, &dilatedMat, kernel)

	// mask irrelevant content
	maskedImageFile := i.getNewFilename("mask")
	maskMat := gocv.IMRead("assets/plan_mask.png", gocv.IMReadAnyColor)
	maskBwMat := gocv.NewMat()
	gocv.CvtColor(maskMat, &maskBwMat, gocv.ColorBGRAToGray)
	postProcessed := gocv.NewMat()
	gocv.Subtract(dilatedMat, maskBwMat, &postProcessed)
	gocv.IMWrite(maskedImageFile, postProcessed)
	log.Print("  Stored masked ocr image, starting contour detection...")

	contours := gocv.FindContours(postProcessed, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	ocrTargetMat := grayMat.Clone()
	namePattern := regexp.MustCompile(`[^\p{L}\d_ ]+`)
	yearOnlyPattern := regexp.MustCompile(`^2[0-9]{3}$`)
	yearPattern := regexp.MustCompile(`2[0-9]{3}`)

	tesseract := gosseract.NewClient()
	tesseract.Languages = []string{"deu"}
	defer tesseract.Close()

	log.Printf("  Starting OCR on %d contours looking for '%s'...", contours.Size(), needle)
	for i := 0; i < contours.Size(); i++ {
		ocrRect := gocv.BoundingRect(contours.At(i))
		croppedMat := ocrTargetMat.Region(ocrRect)

		imageBytes, _ := gocv.IMEncode(gocv.PNGFileExt, croppedMat)
		tesseract.SetImageFromBytes(imageBytes)
		text, err := tesseract.Text()
		if err != nil {
			log.Print(err)
		}
		if text != "" {
			// cleanup text by trimming and removing anything invalid
			text = namePattern.ReplaceAllString(strings.TrimSpace(text), "")

			// check if we are processing the year only (caution: sometimes, ocr will split the month and year)
			if year == 0 && yearOnlyPattern.MatchString(text) {
				// exact year found, parse
				year, err = strconv.Atoi(text)
				if err != nil {
					log.Print(err)
				}
				log.Printf("    Exact year match %s => %d", text, year)
			}

			if month == 0 {
				// check if we got both month an year in one line
				if yearPattern.MatchString(text) {
					year, err = strconv.Atoi(yearPattern.FindString(text))
					if err != nil {
						log.Print(err)
					}
					log.Printf("    Fuzzy year match %s => %d", text, year)
				}

				// strip numbers from text before checking for month
				nonIntText := strings.TrimSpace(yearPattern.ReplaceAllString(text, ""))
				for idx, m := range MONTH_LIST {
					if nonIntText == m {
						month = idx + 1
						log.Printf("    Month match %s => %d", text, month)
						break
					}
				}
			}

			//  get levenshtein distance against needle
			if x == 0 && y == 0 {
				distance := levenshtein.ComputeDistance(text, needle)

				if distance < 3 {
					x = ocrRect.Min.X + SCHEDULE_OFFSET_X
					y = ocrRect.Min.Y + SCHEDULE_OFFSET_Y
					log.Printf("    Found search string %s, saving vector %d,%d.", text, x, y)
				}
			}
		}
	}

	log.Printf("  Setting month to %s %d.", MONTH_LIST[month-1], year)

	return x, y, month, year
}

func (i *Interpreter) ExtractScheduleRow(x int, y int) string {
	scheduleRowFilename := i.getNewFilename("schedule_row")
	mat := gocv.IMRead(i.plan, gocv.IMReadAnyColor)
	area := image.Rect(x, y, x+SCHEDULE_WIDTH, y+SCHEDULE_HEIGHT)
	scheduleMat := mat.Region(area)

	// to make detection easier, we pad the extracted row with white color
	paddedScheduleMat := gocv.NewMat()
	gocv.CopyMakeBorder(
		scheduleMat,
		&paddedScheduleMat,
		SCHEDULE_PADDING,
		SCHEDULE_PADDING,
		SCHEDULE_PADDING,
		SCHEDULE_PADDING,
		gocv.BorderConstant,
		color.RGBA{R: 255, G: 255, B: 255},
	)
	gocv.IMWrite(scheduleRowFilename, paddedScheduleMat)
	log.Print("  Stored extracted schedule row.")
	return scheduleRowFilename
}

func (i *Interpreter) IdentifyWorkSchedule(scheduleRowFile string, startTime time.Time) ScheduleEntries {
	detectedScheduleRow := i.getNewFilename("schedule_row_detected")
	mat := gocv.IMRead(scheduleRowFile, gocv.IMReadAnyColor)

	maskMat := gocv.NewMat()
	scheduleResults := NewScheduleEntries()
	log.Print("    Starting detection loop for template icons...")
	for _, schedule := range GetScheduleTypes() {
		iconMat := gocv.IMRead(schedule.TemplateImage, gocv.IMReadAnyColor)

		resultMat := gocv.NewMatWithSize(mat.Rows()-iconMat.Rows()+1, mat.Cols()-iconMat.Cols()+1, gocv.MatTypeCV32FC1)
		gocv.MatchTemplate(mat, iconMat, &resultMat, gocv.TmCcoeffNormed, maskMat)
		gocv.Threshold(resultMat, &resultMat, 0.8, 1.0, gocv.ThresholdToZero)

		for {
			var threshold float32 = 0.80
			_, maxVal, _, maxLoc := gocv.MinMaxLoc(resultMat)

			if maxVal >= threshold {
				matchRect := image.Rect(maxLoc.X, maxLoc.Y, maxLoc.X+iconMat.Size()[1], maxLoc.Y+iconMat.Size()[0])

				// make sure we dont add false positives
				if added := scheduleResults.AddEntry(schedule, startTime, maxLoc.X, maxLoc.Y); added {
					log.Printf("      Found match for %s at %d,%d...", schedule.Code, maxLoc.X, maxLoc.Y)
					gocv.Rectangle(&mat, matchRect, color.RGBA{R: 0, G: 255, B: 0}, 2)
				}

				// fill the resultMat area to prevent finding the template again
				gocv.Rectangle(&resultMat, matchRect, color.RGBA{R: 0, G: 0, B: 0}, -1)
			} else {
				break
			}
		}
	}

	gocv.IMWrite(detectedScheduleRow, mat)
	log.Print("  Stored detection results.")
	return scheduleResults
}

func (i *Interpreter) getNewFilename(part string) string {
	return strings.Replace(i.plan, ".png", "_"+part+".png", -1)
}

func removeLine(threshMat *gocv.Mat, sizeVector image.Point, maskColor color.RGBA) {
	kernel := gocv.GetStructuringElement(gocv.MorphRect, sizeVector)
	morphMat := gocv.NewMat()
	gocv.MorphologyEx(*threshMat, &morphMat, gocv.MorphOpen, kernel)
	contours := gocv.FindContours(morphMat, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	for i := 0; i < contours.Size(); i++ {
		gocv.DrawContours(threshMat, contours, i, maskColor, 5)
	}
}
