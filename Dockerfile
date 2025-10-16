# =========================================
# Stage 1: Build the Go Application
# =========================================
FROM golang:latest AS resetter
WORKDIR /resetter
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
# Build the Go application
RUN go build -buildvcs=false -o /blaze-reset github.com/geniot/blaze-tools/src
# =========================================
# Stage 2: Prepare executable
# =========================================
FROM golang:latest as run
WORKDIR /app
COPY --from=resetter /blaze-reset ./blaze-reset
EXPOSE 8333
CMD ["/app/blaze-reset"]