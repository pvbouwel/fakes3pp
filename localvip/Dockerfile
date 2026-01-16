# Doing a multi-stage build to make sure to have passing of unittests enforced
FROM docker.io/golang:1.25 AS base

LABEL org.opencontainers.image.source=https://github.com/VITObelgium/fakes3pp/localvip
LABEL org.opencontainers.image.description="FakeS3++ init-container for creating a local virtual IP"
LABEL org.opencontainers.image.licenses=AGPL-3.0

COPY go.mod /usr/src/localvip/go.mod
COPY go.sum /usr/src/localvip/go.sum
WORKDIR /usr/src/localvip
# To not have to fetch dependencies each build
RUN go mod download
RUN go mod tidy
ADD . /usr/src/localvip/
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# If tests pass then build the final stage
FROM scratch

# We need our binary
COPY --from=base /usr/src/localvip/main /localvip

ENTRYPOINT [ "/localvip" ]
