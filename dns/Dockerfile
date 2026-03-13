FROM golang:1.25


WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . . 


RUN CGO_ENABLED=0 GOOS=linux go build -o ./bdns ./cmd/main/main.go


CMD ["./bdns", "-ingest"]
EXPOSE 853
CMD ["./bdns"]
