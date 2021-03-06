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
const SCHEDULE_OFFSET_Y = -25
const SCHEDULE_WIDTH = 2980
const SCHEDULE_HEIGHT = 95
const SCHEDULE_PADDING_Y = 20
const SCHEDULE_PADDING_X = 40

var MONTH_LIST = [][]string{
	{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli",
		"August", "September", "Oktober", "November", "Dezember"},
	{"janvier", "février", "mars", "avril", "mai", "juin", "juillet",
		"août", "septembre", "octobre", "novembre", "décembre"},
}

type Interpreter struct {
	plan         string
	deskewedPlan string
}

func New(planImage string) Interpreter {
	return Interpreter{plan: planImage}
}

func (i *Interpreter) GetSearchVector(needle string) (x int, y int, month int, year int) {
	mat := gocv.IMRead(i.plan, gocv.IMReadAnyColor)

	// pre-process original image and save all steps
	grayMat := gocv.NewMat()
	gocv.CvtColor(mat, &grayMat, gocv.ColorBGRAToGray)
	gocv.IMWrite(i.getNewFilename("1_gray"), grayMat)
	gaussMat := gocv.NewMat()
	gocv.GaussianBlur(grayMat, &gaussMat, image.Point{X: 5, Y: 5}, 0, 0, gocv.BorderDefault)
	gocv.IMWrite(i.getNewFilename("2_gaussed"), gaussMat)
	threshMat := gocv.NewMat()
	gocv.Threshold(gaussMat, &threshMat, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)
	gocv.IMWrite(i.getNewFilename("3_thresh"), threshMat)

	matWidth := mat.Size()[1]
	matHeight := mat.Size()[0]

	// deskew image based on threshold image
	skewAngle := i.getSkewAngle(&threshMat)
	log.Printf("  De-Skewing image by %.4f degrees", skewAngle)
	center := image.Point{X: matWidth / 2, Y: matHeight / 2}
	rotationMatrix := gocv.GetRotationMatrix2D(center, skewAngle, 1.0)

	// de-skew both grayscale/thresh and original material
	gocv.WarpAffineWithParams(mat, &mat, rotationMatrix, image.Point{X: matWidth, Y: matHeight}, gocv.InterpolationCubic, gocv.BorderConstant, color.RGBA{R: 255, G: 255, B: 255})
	gocv.WarpAffineWithParams(grayMat, &grayMat, rotationMatrix, image.Point{X: matWidth, Y: matHeight}, gocv.InterpolationCubic, gocv.BorderConstant, color.RGBA{R: 255, G: 255, B: 255})
	gocv.WarpAffineWithParams(threshMat, &threshMat, rotationMatrix, image.Point{X: matWidth, Y: matHeight}, gocv.InterpolationCubic, gocv.BorderConstant, color.RGBA{R: 255, G: 255, B: 255})
	i.deskewedPlan = i.getNewFilename("4_deskewed")
	gocv.IMWrite(i.deskewedPlan, mat)

	// remove distracting lines
	removeLine(&threshMat, image.Point{X: 80, Y: 1}, color.RGBA{R: 0, G: 0, B: 0})
	removeLine(&threshMat, image.Point{X: 1, Y: 80}, color.RGBA{R: 0, G: 0, B: 0})
	ocrImageFile := i.getNewFilename("5_ocr_base")
	gocv.IMWrite(ocrImageFile, threshMat)
	threshMat.Close()

	// reload fresh ocr mat from disk
	ocrMat := gocv.IMRead(ocrImageFile, gocv.IMReadGrayScale)

	// configure text detection
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 25, Y: 9})
	gocv.Dilate(ocrMat, &ocrMat, kernel)

	// mask irrelevant content
	maskMat := gocv.IMRead("assets/plan_mask.png", gocv.IMReadAnyColor)
	// resize mask to fit input
	gocv.Resize(maskMat, &maskMat, image.Point{X: mat.Size()[1], Y: mat.Size()[0]}, 0, 0, gocv.InterpolationLinear)
	maskBwMat := gocv.NewMat()
	gocv.CvtColor(maskMat, &maskBwMat, gocv.ColorBGRAToGray)

	// mask irrelevant content
	gocv.Subtract(ocrMat, maskBwMat, &ocrMat)
	maskedImageFile := i.getNewFilename("6_ocr_mask")
	gocv.IMWrite(maskedImageFile, ocrMat)
	log.Print("  Stored masked ocr image, starting contour detection...")

	contours := gocv.FindContours(ocrMat, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	ocrTargetMat := grayMat.Clone()
	namePattern := regexp.MustCompile(`[^\p{L}\d_ ]+`)
	yearOnlyPattern := regexp.MustCompile(`^2[0-9]{3}$`)
	yearPattern := regexp.MustCompile(`2[0-9]{3}`)
	monthYearPattern := regexp.MustCompile(`(?P<Month>[a-zA-Z]{3,10}) (?P<Year>2[0-9]{3})`)

	tesseract := gosseract.NewClient()
	tesseract.Languages = []string{"deu"}
	defer tesseract.Close()

	var monthName string

	log.Printf("  Starting OCR on %d contours looking for '%s'...", contours.Size(), needle)
	for i := 0; i < contours.Size(); i++ {
		ocrRect := gocv.BoundingRect(contours.At(i))
		croppedMat := ocrTargetMat.Region(ocrRect)

		imageBytes, _ := gocv.IMEncode(gocv.PNGFileExt, croppedMat)
		tesseract.SetImageFromBytes(imageBytes.GetBytes())
		text, err := tesseract.Text()
		if err != nil {
			log.Print(err)
		}
		if text != "" {
			// cleanup text by trimming and removing anything invalid
			text = namePattern.ReplaceAllString(strings.TrimSpace(text), "")

			log.Printf("    Processing text %s", text)

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
				if monthYearPattern.MatchString(text) {
					matches := monthYearPattern.FindStringSubmatch(text)
					if len(matches) > 0 {
						year, err = strconv.Atoi(matches[monthYearPattern.SubexpIndex("Year")])
						if err != nil {
							log.Print(err)
						}
						monthName := matches[monthYearPattern.SubexpIndex("Month")]
						month = GetMonthIndex(monthName)
						log.Printf("    Fuzzy year/month match: %s => y = %d / m = %d", text, year, month)
					}
				}

				if month == 0 {
					// strip numbers from text before checking for month directly
					month = GetMonthIndex(strings.TrimSpace(yearPattern.ReplaceAllString(text, "")))
					if month > 0 {
						monthName = text
						log.Printf("    Month match %s => %d", text, month)
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

	log.Printf("  Setting month to %s %d.", monthName, year)

	return x, y, month, year
}

func GetMonthIndex(month string) int {
	for _, lang := range MONTH_LIST {
		for idx, m := range lang {
			if month == m {
				return idx + 1
			}
		}
	}
	return 0
}

func (i *Interpreter) ExtractScheduleRow(x int, y int) string {
	scheduleRowFilename := i.getNewFilename("schedule_row")
	mat := gocv.IMRead(i.deskewedPlan, gocv.IMReadAnyColor)
	area := image.Rect(x, y, x+SCHEDULE_WIDTH, y+SCHEDULE_HEIGHT)
	scheduleMat := mat.Region(area)

	// to make detection easier, we pad the extracted row with white color
	paddedScheduleMat := gocv.NewMat()
	gocv.CopyMakeBorder(
		scheduleMat,
		&paddedScheduleMat,
		SCHEDULE_PADDING_Y,
		SCHEDULE_PADDING_Y,
		SCHEDULE_PADDING_X,
		SCHEDULE_PADDING_X,
		gocv.BorderConstant,
		color.RGBA{R: 255, G: 255, B: 255, A: 1},
	)
	gocv.IMWrite(scheduleRowFilename, paddedScheduleMat)
	log.Print("  Stored extracted schedule row.")
	return scheduleRowFilename
}

func (i *Interpreter) IdentifyWorkSchedule(scheduleRowFile string, startTime time.Time, fuzziness int) ScheduleEntries {
	scheduleRowMat := gocv.IMRead(scheduleRowFile, gocv.IMReadAnyColor)

	// apply gaussion blur to counter scan-fuzziness
	gocv.GaussianBlur(scheduleRowMat, &scheduleRowMat, image.Point{X: fuzziness, Y: fuzziness}, 0, 0, gocv.BorderDefault)
	gocv.IMWrite(i.getNewFilename("schedule_row_1_gaussed"), scheduleRowMat)

	maskMat := gocv.NewMat()
	scheduleResults := NewScheduleEntries(startTime)
	var threshold float32 = 0.8
	log.Print("    Starting detection loop for template icons...")
	for _, schedule := range GetScheduleTypes() {
		log.Printf("    Starting detection loop for %q...", schedule.Code)
		iconMat := gocv.IMRead(schedule.TemplateImage, gocv.IMReadAnyColor)

		resultMat := gocv.NewMatWithSize(scheduleRowMat.Rows()-iconMat.Rows()+1, scheduleRowMat.Cols()-iconMat.Cols()+1, gocv.MatTypeCV32FC1)
		gocv.MatchTemplate(scheduleRowMat, iconMat, &resultMat, gocv.TmCcoeffNormed, maskMat)
		gocv.Threshold(resultMat, &resultMat, threshold, 1.0, gocv.ThresholdToZero)

		for {
			_, maxVal, _, maxLoc := gocv.MinMaxLoc(resultMat)

			if maxVal >= threshold {
				matchRect := image.Rect(maxLoc.X, maxLoc.Y, maxLoc.X+iconMat.Size()[1], maxLoc.Y+iconMat.Size()[0])

				// make sure we dont add false positives
				if added := scheduleResults.AddEntry(schedule, startTime, maxLoc.X, maxLoc.Y); added {
					log.Printf("      Found match for %q at %d,%d...", schedule.Code, maxLoc.X, maxLoc.Y)
					gocv.Rectangle(&scheduleRowMat, matchRect, color.RGBA{R: 0, G: 255, B: 0}, 2)
				}

				// fill the resultMat area to prevent finding the template again
				gocv.Rectangle(&resultMat, matchRect, color.RGBA{R: 0, G: 0, B: 0}, -1)
			} else {
				break
			}
		}
	}

	gocv.IMWrite(i.getNewFilename("schedule_row_2_detected"), scheduleRowMat)
	log.Print("  Stored detection results.")
	return scheduleResults
}

func (i *Interpreter) getNewFilename(part string) string {
	return strings.Replace(i.plan, ".png", "_"+part+".png", -1)
}

func (i *Interpreter) getSkewAngle(threshMat *gocv.Mat) float64 {
	skewMat := gocv.NewMat()
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 3, Y: 80})
	gocv.MorphologyEx(*threshMat, &skewMat, gocv.MorphOpen, kernel)
	contours := gocv.FindContours(skewMat, gocv.RetrievalList, gocv.ChainApproxSimple)

	highestContour := gocv.PointVector{}
	currentMaxHeight := 0
	for i := 0; i < contours.Size(); i++ {
		gocv.DrawContours(&skewMat, contours, i, color.RGBA{R: 0, G: 255, B: 0}, 1)
		currentContour := contours.At(i)
		rect := gocv.BoundingRect(currentContour)
		if currentMaxHeight < rect.Dy() {
			highestContour = currentContour
			currentMaxHeight = rect.Dy()
		}
		gocv.Rectangle(&skewMat, rect, color.RGBA{0, 0, 255, 0}, 2)
	}

	minAreaRect := gocv.MinAreaRect(highestContour)
	return minAreaRect.Angle - 90
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
