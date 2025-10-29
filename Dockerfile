FROM golang:1.25.2-alpine AS builder 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download 

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o taskflow_server ./main.go

FROM alpine:3.20 AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates
COPY --from=builder /app/taskflow_server ./

EXPOSE 8080

CMD [ "./taskflow_server" ]