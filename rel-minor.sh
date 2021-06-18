#!/bin/bash
latesttag=$(git tag --sort=-taggerdate)
tag=${latesttag:1}

IFS='.'
read -a v <<< "$tag"

major="${v[0]:=0}"
minor="${v[1]:=0}"
patch="0"

((minor++))

newtag="v${major}.${minor}.${patch}"

echo "New minor-bumped tag: ${newtag}"
git tag -a "${newtag}" -m "${newtag}"

read -p "Press enter to push --tags or ^C to abort"
git push --tags
