FROM golang:1.22

WORKDIR /usr/src/user

COPY user/ ./

WORKDIR /usr/src/frontend

COPY frontend/go.mod frontend/go.sum ./
RUN go mod download && go mod verify

COPY frontend/ .
RUN go build -v -o /usr/local/bin/frontend ./...

ENTRYPOINT ["frontend"]
