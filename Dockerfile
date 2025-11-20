FROM golang:1.25-alpine as builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/main ./cmd/app/main.go

RUN CGO_ENABLED=0 go build -o /app/migrate ./cmd/migrate/main.go


FROM alpine:latest

COPY --from=builder /app/main /main
COPY --from=builder /app/migrate /migrate

CMD ["/main"]

