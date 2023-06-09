FROM golang:1.20-alpine3.18 AS builder

WORKDIR /app

COPY go.mod .
COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o /usr/bin/course-sense-go cmd/coursesense/main.go

FROM alpine:3.16.2

COPY --from=builder /usr/bin/course-sense-go /usr/bin/course-sense-go

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

ENTRYPOINT [ "/usr/bin/course-sense-go" ]
