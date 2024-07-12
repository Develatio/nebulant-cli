FROM golang:1.21.0

WORKDIR /app
COPY ./ /app
RUN go mod download

ENTRYPOINT ["/app/docker/entrypoint.sh"]
CMD ["all"]