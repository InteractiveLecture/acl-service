#!/bin/bash
GO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o out/main .
docker build -t openservice/acl-service .
