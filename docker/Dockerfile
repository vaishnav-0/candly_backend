# syntax=docker/dockerfile:1

## Build
FROM golang:1.19.3-bullseye as build
LABEL stage=builder

WORKDIR /app

ADD ./server/go.mod ./server/go.sum ./
RUN go mod download

COPY server ./
RUN go build  -o /candly-server ./cmd/server/server.go



## Deploy
FROM gcr.io/distroless/base-debian11:debug

WORKDIR /

COPY --from=build /candly-server /

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/candly-server"]
