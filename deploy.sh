#!/usr/bin/env sh
export IMG="localhost:32000/kun-operator:latest"
make docker-build docker-push IMG=${IMG}
make deploy IMG=${IMG}
