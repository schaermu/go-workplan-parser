package parser

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/schaermu/workplan-parser/internal/interpreter"
	"gopkg.in/gographics/imagick.v2/imagick"
)

type Parser struct {
	needle    string
	pdfFile   string
	fuzziness int
}

func New(needle string, filePath string, detectionFuzziness int) *Parser {
	return &Parser{
		needle:    needle,
		pdfFile:   filePath,
		fuzziness: detectionFuzziness,
	}
}

func (p *Parser) ProcessPages(pageNumber int) map[int]*interpreter.ScheduleEntries {
	file, err := os.Open(p.pdfFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	tempDir, err := ioutil.TempDir("temp", "workplan-parser-*")
	if err != nil {
		log.Fatal(err)
	}

	selectedPages := []string{}
	if pageNumber > 0 {
		// construct pagenumber restriction
		selectedPages = append(selectedPages, fmt.Sprintf("%d", pageNumber))
	}

	config := pdfcpu.NewDefaultConfiguration()
	err = api.ExtractPagesFile(p.pdfFile, tempDir, selectedPages, config)
	if err != nil {
		log.Fatal(err)
	}

	pageFiles, err := ioutil.ReadDir(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	res := map[int]*interpreter.ScheduleEntries{}
	for index, f := range pageFiles {
		imageFileName := path.Join(tempDir, strings.Replace(f.Name(), ".pdf", ".png", -1))
		log.Printf("Processing page %s...\n", f.Name())
		p.convertPageToImage(path.Join(tempDir, f.Name()), imageFileName)

		interpreter := interpreter.New(imageFileName)
		x, y, month, year := interpreter.GetSearchVector(p.needle)
		scheduleRowFile := interpreter.ExtractScheduleRow(x, y)
		startTime := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Now().Location())
		schedule := interpreter.IdentifyWorkSchedule(scheduleRowFile, startTime, p.fuzziness)
		schedule.SortEntriesByDate()
		schedule.RemoveDuplicates()

		res[index+1] = &schedule
	}
	return res
}

func (p *Parser) convertPageToImage(pdfPath string, target string) {
	ctx, err := api.ReadContextFile(pdfPath)
	if err != nil {
		log.Fatalf("Could not load context for file %s", pdfPath)
	}

	// auto-rotate pdf if it was scanned in portrait format
	if dimensions, err := ctx.PageDims(); err == nil && len(dimensions) == 1 {
		if dimensions[0].Width > dimensions[0].Height {
			// wrong orientation, rotate file
			api.RotateFile(pdfPath, pdfPath, 90, nil, nil)
		}
	}

	imagick.Initialize()
	defer imagick.Terminate()
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	mw.SetOption("density", "500")
	mw.SetOption("psd:fit-page", "5000x")

	if err := mw.ReadImage(pdfPath); err != nil {
		log.Println(err)
	}

	mw.SetIteratorIndex(0)
	mw.SetImageFormat("png")

	if err := mw.WriteImage(target); err != nil {
		log.Println(err)
	}

	log.Printf("  Stored rasterized page at %s\n", target)
}
