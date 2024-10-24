FROM golang:alpine

COPY . /app

RUN cd /app && go build -o jekill cmd/main.go

FROM alpine

COPY --from=0 /app/jekill /app/jekill

ENTRYPOINT [ "/app/jekill" ]