FROM golang:1.20-alpine3.18 AS builder

WORKDIR /app

COPY go.sum .
COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o /usr/bin/course-sense-go cmd/coursesense/main.go

FROM alpine:3.18.2

COPY --from=builder /usr/bin/course-sense-go /usr/bin/course-sense-go

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
ENV app_env=prod

ENTRYPOINT [ "/usr/bin/course-sense-go" ]
