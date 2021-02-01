#!/bin/bash

echo -e "\033[31m Building new Docker image"
docker build -t vitsensei/infogrid --rm -f build/package/app/Dockerfile .

echo -e "\033[31m Pushing built Docker image to Docker Hub"
docker push vitsensei/infogrid

echo -e "\033[31m Deploy code to remote server"

ssh root@www.infogrid.app "cd infogrid; \
    docker-compose down; \
    docker-compose pull; \
    docker-compose up"

