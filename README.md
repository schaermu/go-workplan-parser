# workplan-parser
WIP

## Prerequisites (wip)
* tesseract-ocr
* libmagickwand
* OpenCV 4.5.1

## Setting up prereqs
1. https://github.com/tesseract-ocr/tessdoc/blob/master/Downloads.md
2. sudo apt-get install libmagickwand-dev
3. sed -i '/disable ghostscript format types/,+6d' /etc/ImageMagick-6/policy.xml
4. go get -u -d gocv.io/x/gocv
5. cd $GOPATH/src/gocv.io/x/gocv
6. make install

## Usage (wip)
```
workplan-parser parse -i [PATH_TO_PDF] -e "[EMPLOYEE_NAME]]"
```