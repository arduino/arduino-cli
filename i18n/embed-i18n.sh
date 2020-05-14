#!/bin/sh

git add -N ./data
git diff --exit-code ./data &> /dev/null || rice embed-go