#!/bin/bash

set -e

case $1 in
    all)
        /app/docker/entrypoint.sh "linux-arm"
        /app/docker/entrypoint.sh "linux-arm64"
        /app/docker/entrypoint.sh "linux-386"
        /app/docker/entrypoint.sh "linux-amd64"
        /app/docker/entrypoint.sh "freebsd-arm"
        /app/docker/entrypoint.sh "freebsd-arm64"
        /app/docker/entrypoint.sh "freebsd-386"
        /app/docker/entrypoint.sh "freebsd-amd64"
        /app/docker/entrypoint.sh "openbsd-arm"
        /app/docker/entrypoint.sh "openbsd-arm64"
        /app/docker/entrypoint.sh "openbsd-386"
        /app/docker/entrypoint.sh "openbsd-amd64"
        /app/docker/entrypoint.sh "windows-arm"
        /app/docker/entrypoint.sh "windows-arm64"
        /app/docker/entrypoint.sh "windows-386"
        /app/docker/entrypoint.sh "windows-amd64"
        /app/docker/entrypoint.sh "darwin-arm64"
        /app/docker/entrypoint.sh "darwin-amd64"
        /app/docker/entrypoint.sh wasm
        ;;
    "linux-arm")
        echo "--> Building linux arm executable"
        GOOS=linux GOARCH=arm GOEXE= DIST_SUFFIX=-linux-arm make build_platform
        ;;
    "linux-arm64")
        echo "--> Building linux arm64 executable"
        GOOS=linux GOARCH=arm64 GOEXE= DIST_SUFFIX=-linux-arm64 make build_platform
        ;;
    "linux-386")
        echo "--> Building linux 386 executable"
        GOOS=linux GOARCH=386 GOEXE= DIST_SUFFIX=-linux-386 make build_platform
        ;;
    "linux-amd64")
        echo "--> Building linux amd64 executable"
        GOOS=linux GOARCH=amd64 GOEXE= DIST_SUFFIX=-linux-amd64 make build_platform
        ;;
    "freebsd-arm")
        echo "--> Building freebsd arm executable"
        GOOS=freebsd GOARCH=arm GOEXE= DIST_SUFFIX=-freebsd-arm make build_platform
        ;;
    "freebsd-arm64")
        echo "--> Building freebsd arm64 executable"
        GOOS=freebsd GOARCH=arm64 GOEXE= DIST_SUFFIX=-freebsd-arm64 make build_platform
        ;;
    "freebsd-386")
        echo "--> Building freebsd 386 executable"
        GOOS=freebsd GOARCH=386 GOEXE= DIST_SUFFIX=-freebsd-386 make build_platform
        ;;
    "freebsd-amd64")
        echo "--> Building freebsd amd64 executable"
        GOOS=freebsd GOARCH=amd64 GOEXE= DIST_SUFFIX=-freebsd-amd64 make build_platform
        ;;
    "openbsd-arm")
        echo "--> Building openbsd arm executable"
        GOOS=openbsd GOARCH=arm GOEXE= DIST_SUFFIX=-openbsd-arm make build_platform
        ;;
    "openbsd-arm64")
        echo "--> Building openbsd arm64 executable"
        GOOS=openbsd GOARCH=arm64 GOEXE= DIST_SUFFIX=-openbsd-arm64 make build_platform
        ;;
    "openbsd-386")
        echo "--> Building openbsd 386 executable"
        GOOS=openbsd GOARCH=386 GOEXE= DIST_SUFFIX=-openbsd-386 make build_platform
        ;;
    "openbsd-amd64")
        echo "--> Building openbsd amd64 executable"
        GOOS=openbsd GOARCH=amd64 GOEXE= DIST_SUFFIX=-openbsd-amd64 make build_platform
        ;;
    "windows-arm")
        echo "--> Building windows arm executable"
        GOOS=windows GOARCH=arm GOEXE=.exe DIST_SUFFIX=-windows-arm.exe make build_platform
        ;;
    "windows-arm64")
        echo "--> Building windows arm64 executable"
        GOOS=windows GOARCH=arm64 GOEXE=.exe DIST_SUFFIX=-windows-arm64.exe make build_platform
        ;;
    "windows-386")
        echo "--> Building windows 386 executable"
        GOOS=windows GOARCH=386 GOEXE=.exe DIST_SUFFIX=-windows-386.exe make build_platform
        ;;
    "windows-amd64")
        echo "--> Building windows amd64 executable"
        GOOS=windows GOARCH=amd64 GOEXE=.exe DIST_SUFFIX=-windows-amd64.exe make build_platform
        ;;
    "darwin-arm64")
        echo "--> Building darwin arm64 executable"
        GOOS=darwin GOARCH=arm64 GOEXE= DIST_SUFFIX=-darwin-arm64 make build_platform
        ;;
    "darwin-amd64")
        echo "--> Building darwin amd64 executable"
        GOOS=darwin GOARCH=amd64 GOEXE= DIST_SUFFIX=-darwin-amd64 make build_platform
        ;;
    wasm)
        echo "--> Building wasm executable"
        GOOS=js GOARCH=wasm GOEXE= DIST_SUFFIX=-js-wasm make build_platform
        ;;
    *)
        exec "$@"
        ;;
esac
