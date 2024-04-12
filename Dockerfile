FROM golang:1.21

WORKDIR /app
COPY ./ ./
RUN go mod download
RUN make build
RUN make buildbridge

RUN openssl genrsa -des3 -passout pass:x -out server.pass.key 2048
RUN openssl rsa -passin pass:x -in server.pass.key -out /app/server.key
RUN rm server.pass.key
RUN openssl req -new -key /app/server.key -out /app/server.csr \
    -subj "/C=ES/ST=Madrid/L=Madrid/O=Develatio/OU=ID Department/CN=localhost"
RUN openssl x509 -req -days 365 -in /app/server.csr -signkey /app/server.key -out /app/server.crt