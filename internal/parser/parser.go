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
	needle  string
	pdfFile string
}

func New(needle string, filePath string) *Parser {
	return &Parser{
		needle:  needle,
		pdfFile: filePath,
	}
}

func (p *Parser) ProcessPages() {
	file, err := os.Open(p.pdfFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	tempDir, err := ioutil.TempDir("temp", "workplan-parser-*")
	if err != nil {
		log.Fatal(err)
	}

	config := pdfcpu.NewDefaultConfiguration()
	err = api.ExtractPagesFile(p.pdfFile, tempDir, nil, config)
	if err != nil {
		log.Fatal(err)
	}

	images, err := ioutil.ReadDir(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range images {
		imageFileName := path.Join(tempDir, strings.Replace(f.Name(), ".pdf", ".png", -1))
		log.Printf("Processing page %s...\n", f.Name())
		p.convertPageToImage(path.Join(tempDir, f.Name()), imageFileName)

		interpreter := interpreter.New(imageFileName)
		x, y, month, year := interpreter.GetSearchVector(p.needle)
		scheduleRowFile := interpreter.ExtractScheduleRow(x, y)
		startTime := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Now().Location())
		schedule := interpreter.IdentifyWorkSchedule(scheduleRowFile, startTime)
		schedule.SortEntriesByDate()

		for _, entry := range schedule.Entries {
			println(fmt.Sprintf("%s - %s", entry.GetWorktime(), entry.Code))
		}
	}
}

func (p *Parser) convertPageToImage(pdfPath string, target string) {
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
