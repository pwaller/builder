FROM golang:1.4

RUN go get -v github.com/awslabs/aws-sdk-go/...
RUN go get -v github.com/gorilla/websocket

COPY . /go/src/github.com/pwaller/builder

RUN go install -v github.com/pwaller/builder