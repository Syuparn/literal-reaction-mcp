# ---- Stage 1: Build SQLite word database from MeCab IPADIC ----
FROM debian:bookworm-slim AS db-builder

ENV IPADIC_URL=https://drive.google.com/uc?export=download&id=1aQCOnE4aGPz-3tXpt6Mok2fcPgbnnouh
ENV SQL_NAME=words.sqlite3
ENV TMP_SQL_DIR=/tmp/sqlite

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        wget \
        ca-certificates \
        sqlite3 \
        gawk \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY db/ ./db/

RUN wget -O ipadic.tar.gz "${IPADIC_URL}" \
    && tar zxf ipadic.tar.gz \
    && mkdir -p "${TMP_SQL_DIR}" \
    && COPY_TO="${TMP_SQL_DIR}" SQL_NAME="${SQL_NAME}" bash db/setup-db.bash

# ---- Stage 2: Build Go binary ----
FROM golang:1.26-bookworm AS go-builder

ENV CGO_ENABLED=1

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build \
    -a \
    -ldflags '-linkmode external -extldflags "-static"' \
    -o literal-reaction-mcp \
    .

# ---- Stage 3: Final minimal image ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=go-builder /app/literal-reaction-mcp .
COPY --from=db-builder /tmp/sqlite/words.sqlite3 .

ENV DB_PATH=/app/words.sqlite3
ENV ADDR=:8080

EXPOSE 8080

ENTRYPOINT ["./literal-reaction-mcp"]
