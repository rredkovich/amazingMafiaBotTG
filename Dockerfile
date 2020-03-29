FROM golang:alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o mafiabot .


FROM alpine
ARG release_version
ENV RELEASE_VERSION=$release_version
RUN adduser -S -D -H -h /app bot
USER bot
COPY --from=builder /build/mafiabot /app/
WORKDIR /app
COPY *.go ./
# source code for sentry
COPY /authorization ./authorization
COPY /game ./game
COPY /types ./types
COPY /templates ./templates

CMD ["./mafiabot"]