# Step 1: Modules caching
FROM golang:1.25.14-alpine3.23 as modules

COPY go.mod go.sum /modules/

WORKDIR /modules

RUN go mod download

# Step 2: Tests
FROM golang:1.25.14-alpine3.23

COPY --from=modules /go/pkg /go/pkg
COPY . /app

WORKDIR /app

RUN go env -w CGO_ENABLED=0 \
    && go env -w GOOS=linux  \
    && go env -w GOARCH=amd64

# -count=1 disables Go test cache so CI always re-runs (F04).
CMD ["go", "test", "-v", "-count=1", "./integration-test/..."]
