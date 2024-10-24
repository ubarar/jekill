FROM golang:alpine

COPY . /app

RUN cd /app && go build -o jekill cmd/main.go

ENTRYPOINT [ "/app/jekill" ]