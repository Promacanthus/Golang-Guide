---
title: 04-Dockerfile多阶段构建
date: 2020-04-14T10:09:14.098627+08:00
draft: false
---

只需要在 Dockerfile 中多次使用 FORM 声明，每次 FROM 指令可以使用不同的基础镜像，并且每次 FROM 指令都会开始新的构建，可以选择将一个阶段的构建结果复制到另一个阶段，在最终的镜像中只会留下最后一次构建的结果，并且只需要编写一个 Dockerfile 文件。

> 这里值得注意的是：需要确保 Docker 的版本在 17.05 及以上。

## 具体操作

1. 在 Dockerfile 里可以使用 as 来为某一阶段取一个别名”build-env”：`FROM golang:1.11.2-alpine3.8 AS build-env`
2. 然后从上一阶段的镜像中复制文件，也可以复制任意镜像中的文件：`COPY –from=build-env /go/bin/hello /usr/bin/hello`

```dockerfile
FROM golang:1.11.4-alpine3.8 AS build-env
ENV GO111MODULE=off
ENV GO15VENDOREXPERIMENT=1
ENV GITPATH=github.com/lattecake/hello
RUN mkdir -p /go/src/${GITPATH}
COPY ./ /go/src/${GITPATH}
RUN cd /go/src/${GITPATH} && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -v

FROM alpine:latest
ENV apk –no-cache add ca-certificates
COPY --from=build-env /go/bin/hello /root/hello
WORKDIR /root
CMD ["/root/hello"]
```
