FROM golang:1.22 as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /app/cmd ./...

FROM scratch
COPY --from=build app/cmd .
EXPOSE 8080
ENTRYPOINT ["app"]