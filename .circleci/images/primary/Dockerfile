FROM ubuntu:focal

ENV DEBIAN_FRONTEND=noninteractive

# install base, tesseract and magickwand
RUN apt-get update && apt-get install -y \
        sudo \
        curl \
        git \
        gnupg \
        gzip \
        make \
		build-essential \
        software-properties-common \
		ca-certificates \
        tesseract-ocr \
        libtesseract-dev \
        libmagickwand-dev \
    && apt-get clean \
	&& rm -rf /var/lib/apt/lists/* \
    && sed -i '/disable ghostscript format types/,+6d' /etc/ImageMagick-6/policy.xml
    
ENV PATH="/usr/local/go/bin:${PATH}"

# install golang
RUN curl -L https://golang.org/dl/go1.16.4.linux-amd64.tar.gz -o go1.16.4.linux-amd64.tar.gz \
    && rm -rf /usr/local/go && tar -C /usr/local -xzf go1.16.4.linux-amd64.tar.gz

ENV GOPATH="/root/go"

# install opencv using gocv package
RUN go get -u -d gocv.io/x/gocv \
    && cd /root/go/pkg/mod/gocv.io/x/gocv@v0.27.0/ \
    && make install