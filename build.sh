#!/bin/bash
GO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o out/main .
docker build -t interactive-lecture/lecture-service .

cd sql
./build.sh
cd ..
