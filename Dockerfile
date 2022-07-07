FROM golang
COPY . .
RUN CGO_ENABLED=0 go build -o sender cmd/sender/*.go
CMD ["cat", "sender"]

