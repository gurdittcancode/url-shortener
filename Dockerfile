FROM golang:1.24-alpine

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o url-shortener .

CMD [ "./url-shortener" ]