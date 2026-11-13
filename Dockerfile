# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/rw-pair .

FROM scratch
COPY --from=build /out/rw-pair /rw-pair
ENV GOMEMLIMIT=48MiB
EXPOSE 8080
ENTRYPOINT ["/rw-pair"]
