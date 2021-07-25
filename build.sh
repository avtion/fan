#!/bin/bash
tag=$(git rev-parse --short HEAD)
docker build -t fan:"${tag}" .
