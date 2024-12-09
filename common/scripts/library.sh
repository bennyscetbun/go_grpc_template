#!/bin/bash

function lib() {
    function build_docker_images() {
        docker image inspect xxxyourappyyy/node >/dev/null || docker build --build-arg USERID=$(id -u) --build-arg USERGRP=$(id -g) -f "$LIBRARY_SH_DIR/dockerfile/node.dockerfile" -t xxxyourappyyy/node .
        docker image inspect xxxyourappyyy/protoc >/dev/null || docker build --build-arg USERID=$(id -u) --build-arg USERGRP=$(id -g) -f "$LIBRARY_SH_DIR/dockerfile/protoc.dockerfile" -t xxxyourappyyy/protoc .
    }

    function get_script_dir() {
        SOURCE=${BASH_SOURCE[0]}
        while [ -L "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
            DIR=$(cd -P "$(dirname "$SOURCE")" >/dev/null 2>&1 && pwd)
            SOURCE=$(readlink "$SOURCE")
            [[ $SOURCE != /* ]] && SOURCE=$DIR/$SOURCE # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
        done
        cd -P "$(dirname "$SOURCE")" >/dev/null 2>&1 && pwd
    }

    function run_if_change() {
        cachepath="$LIBRARY_SH_DIR/.cache"
        mkdir -p "$cachepath"
        scriptAbsolutePathName=$(get_abs_filename "$0")
        filePathHash=($(echo "$1" "$scriptAbsolutePathName" | md5sum -b))
        md5filePath="$cachepath/$filePathHash"
        tmpfile=$(mktemp "$cachepath/tmpsum.XXXXXXX")
        newmd5=$(find $1 -type f -exec md5sum -b {} + | LC_ALL=C sort | md5sum -b)
        (echo "$newmd5" >"$tmpfile")
        shift 1
        if [[ -z ${FORCE_NO_CACHE+x} ]]; then
            diff "$tmpfile" "$md5filePath" >/dev/null 2>&1 || "$@" && cp "$tmpfile" "$md5filePath"
        else
            "$@" && cp "$tmpfile" "$md5filePath"
        fi
        ret=$?
        rm -f "$tmpfile"
        return $ret
    }

    function get_abs_filename() {
        filename=$1
        parentdir=$(dirname "${filename}")

        if [ -d "${filename}" ]; then
            echo "$(cd "${filename}" && pwd)"
        elif [ -d "${parentdir}" ]; then
            echo "$(cd "${parentdir}" && pwd)/$(basename "${filename}")"
        fi
    }

    LIBRARY_SH_DIR=$(get_script_dir)
}
lib
