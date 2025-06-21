FROM golang:alpine AS golang-alpine-build
RUN apk add build-base git

FROM golang-alpine-build AS builder
WORKDIR /go/src/woller
COPY . .
RUN go build

FROM alpine
ENV PATH $PATH:/go/bin/
COPY --from=builder /go/src/woller /go/bin/woller
ENTRYPOINT ["woller"]