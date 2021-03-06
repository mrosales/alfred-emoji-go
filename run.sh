#!/usr/bin/env bash

set -e

SCRIPT_DIRECTORY=$(dirname "$0")

export alfred_workflow_bundleid="com.micahrosales.alfred-emoji-go"
export alfred_workflow_cache="${SCRIPT_DIRECTORY}/alfreddata/cache"
export alfred_workflow_data="${SCRIPT_DIRECTORY}/alfreddata/data"
export alfred_no_update=1

go run ./main.go "$1"
