#!/bin/sh

changelog_filename="CHANGELOG.md"

git fetch --depth 1 origin master

if ! git diff --name-only origin/master | grep -E '.*\.go|go.mod|go.sum'
then
  echo "Go code change not found."
  exit 0
fi

if [ -f ${changelog_filename} ]
then
    if [ "$(git diff --name-only origin/master -- ${changelog_filename} | grep -cw ${changelog_filename})" -ne 1 ]
    then
        echo "Changelog should be updated!"
        exit 1
    fi
fi