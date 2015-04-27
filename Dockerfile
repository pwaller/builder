FROM golang:1.4

COPY . /go/src/github.com/pwaller/builder

RUN go install -v github.com/pwaller/builder