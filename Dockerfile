FROM golang:alpine as base

FROM base as builder
WORKDIR /opt/bot

ADD .gitignore go.mod go.sum main.go ./

RUN go build
RUN ls
RUN echo 000

FROM base
WORKDIR /opt

COPY --from=builder /opt/bot/bot bot

RUN chmod +x /opt/bot

ENTRYPOINT [ "/opt/bot" ]