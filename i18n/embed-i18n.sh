#!/bin/sh

git add -N ./data
git diff --exit-code ./data || rice embed-go