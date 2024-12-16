# 第一阶段：Node.js 构建阶段
FROM node:16 as builder

WORKDIR /build
COPY web/package.json .
RUN npm install
COPY ./web .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) npm run build

# 第二阶段：Go 构建阶段
FROM golang AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux

WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=builder /build/dist ./web/dist
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)'" -o one-api

# 第三阶段：最终镜像阶段
FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=builder2 /build/one-api /
EXPOSE 3000
ENTRYPOINT ["/one-api"]
