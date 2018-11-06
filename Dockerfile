# Build the manager binary
FROM golang:1 as builder

# Copy in the go src
COPY . /go/src/github.com/Ridecell/ridecell-operator
WORKDIR /go/src/github.com/Ridecell/ridecell-operator

# Build
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
  dep ensure && \
  make generate && \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager -tags release github.com/Ridecell/ridecell-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:latest
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /go/src/github.com/Ridecell/ridecell-operator/manager /ridecell-operator
CMD ["/ridecell-operator"]
