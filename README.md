# workplan-parser
WIP

## Prerequisites (wip)
* tesseract-ocr
* tesseract-data-deu
* libmagickwand
* OpenCV 4.5.1

### Arch linux
* tesseract-ocr
* tesseract-data-deu-git
* opencv
* vtk
* glew
* hdf5
* imagemagick

## Setting up prereqs
1. https://github.com/tesseract-ocr/tessdoc/blob/master/Downloads.md
2. sudo apt-get install libmagickwand-dev
3. remove code policy line from /etc/Imagemagick-7/policy.xml
4. go get -u -d gocv.io/x/gocv
5. cd $GOPATH/src/gocv.io/x/gocv
6. make install

## Usage (wip)
```
workplan-parser import -i [PATH_TO_PDF] -e "[EMPLOYEE_NAME]"
```