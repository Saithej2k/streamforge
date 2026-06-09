FROM golang:1.23-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/streamforge ./cmd/streamforge

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/streamforge /usr/local/bin/streamforge

USER nonroot:nonroot
ENTRYPOINT ["streamforge"]
