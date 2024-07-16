FROM golang:latest AS build

WORKDIR /app

COPY . /app/

RUN go mod tidy && go build -o mp .

FROM debian:stable-slim

WORKDIR /app

COPY --from=build /app/mp /app/

CMD [ "/app/mp" ]
