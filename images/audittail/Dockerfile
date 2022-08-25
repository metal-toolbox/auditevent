FROM docker.io/library/golang:1.19 as build-env

WORKDIR /go/src/audittail

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/audittail cmds/audittail/main.go 

FROM gcr.io/distroless/static

COPY --from=build-env /go/bin/audittail /
ENTRYPOINT ["/audittail"]