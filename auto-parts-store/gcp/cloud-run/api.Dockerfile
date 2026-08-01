# Builds either cmd/api or cmd/worker depending on BUILD_TARGET - both are
# thin Go binaries with no runtime dependencies beyond the Postgres/Pub/Sub
# connections they open at startup.
FROM golang:1.26-alpine AS build
ARG BUILD_TARGET=./cmd/api
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ${BUILD_TARGET}

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
