#!/bin/sh
# Check for CHANGELOG updates and fails when not found
changelog_filename="CHANGELOG.md"

git fetch --depth 1 origin master
if [ -f ${changelog_filename} ]
then
    if [ "$(git diff --name-only origin/master -- ${changelog_filename} | grep -cw ${changelog_filename})" -ne 1 ]
    then
        echo "Changelog should be updated!"
        exit 1
    fi
fi