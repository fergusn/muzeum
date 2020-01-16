FROM golang:alpine as build

WORKDIR /build
COPY . .
RUN go build -o muzeum cmd/* 

FROM alpine
COPY --from=build /build/muzeum /usr/local/bin/
ENTRYPOINT [ "muzeum" ]








