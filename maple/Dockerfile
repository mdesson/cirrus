FROM golang:latest as builder

RUN mkdir /build
RUN mkdir -p src/github.com
ADD ./maple/*.go /build/
ADD ./go.mod /build/
ADD ./go.sum /build/
# TODO: Remove configs from docker image, convert to env
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o maple *.go

FROM alpine:latest
COPY ./maple/wait-for-it.sh .
COPY --from=builder /build/maple .
ADD ./maple/*.json .
RUN apk update && apk add bash
RUN chmod +x wait-for-it.sh
ENTRYPOINT ["./wait-for-it.sh" , "postgres:5432" , "--timeout=10" , "--" , "./maple"]