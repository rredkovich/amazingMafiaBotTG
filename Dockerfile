FROM golang:alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o mafiabot .


FROM alpine
ARG RELEASE_VERSION
ENV RELEASE_VERSION=$RELEASE_VERSION
RUN adduser -S -D -H -h /app bot
USER bot
COPY --from=builder /build/mafiabot /app/
WORKDIR /app
COPY *.go ./
# source code for sentry
COPY /authorization ./
COPY /game ./
COPY /types ./

CMD ["./mafiabot"]