#!/bin/bash

IFS=$'\n' read -r -d '' -a tags < <( git tag --sort=-taggerdate && printf '\0' )

git log --pretty=format:"%s" ${tags[1]}..${tags[0]}
