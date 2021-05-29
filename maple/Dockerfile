FROM golang:latest as builder

RUN mkdir /build
ADD ./maple/*.go /build/
ADD ./go.mod /build/
ADD ./go.sum /build/
# TODO: Remove configs from docker image, convert to env
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o maple *.go

FROM alpine:latest
COPY --from=builder /build/maple .
ADD ./maple/*.json .

ENTRYPOINT [ "./maple" ]