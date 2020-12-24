FROM golang:1.15-alpine AS build

WORKDIR /src
COPY . .

RUN go get -d -v ./...
RUN ./build/build.sh

FROM alpine:latest AS bin
WORKDIR /tarpon
COPY --from=build /src/bin/tarpon .

ENTRYPOINT ["./tarpon"]
EXPOSE 5000