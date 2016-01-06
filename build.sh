#!/bin/bash
set -ev
go test
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o out/main .
docker build -t openservice/acl-service:latest .
if [ "${TRAVIS_PULL_REQUEST}" = "false" ] && [ "${TRAVIS_REPO_SLUG}" = "InteractiveLecture/acl-service" ] ; then
  docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD" -e="$DOCKER_EMAIL"
  docker push openservice/acl-service:latest
fi
