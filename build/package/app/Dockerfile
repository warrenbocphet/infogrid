############################
# STEP 1: Add source code to generate executable
############################
FROM vitsensei/infogrid_dep AS builder

ADD . .

# Build the binary.
RUN go build -o infogrid ./cmd/main.go

############################
# STEP 2: Start a small image that only contains executable
############################
FROM alpine:latest

# Copy our static executable.
COPY --from=builder ./go/src/infogrid/infogrid .
COPY --from=builder ./go/src/infogrid/pkg/views/articles/*.gohtml ./views/articles/
COPY --from=builder ./go/src/infogrid/pkg/views/layouts/*.gohtml ./views/layouts/

# Add nano and vim for convenient
RUN apk add --update nano vim

# Run the infogrid binary.
ENTRYPOINT ["./infogrid"]
