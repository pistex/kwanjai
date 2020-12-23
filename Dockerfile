FROM golang:alpine
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/pistex/kwanjai
RUN apk update && apk add --no-cache git
COPY . .
#RUN go test ./... -v
RUN go build -o app .

FROM alpine:latest  
ENV GIN_MODE=release
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=0 /go/src/github.com/pistex/kwanjai/app .
CMD ["./app"]
