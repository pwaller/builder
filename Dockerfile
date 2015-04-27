FROM golang:1.4

RUN go get -v github.com/awslabs/aws-sdk-go/...

COPY . /go/src/github.com/pwaller/builder

RUN go install -v github.com/pwaller/builder