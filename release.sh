#!/bin/bash
set -e

RELEASE_BRANCH=master
VERSION=`cat VERSION`

# Build the release binaries.
make release

# Add necessary files.
cd .builds
for dir in `ls`
do
	cp ../README.md ../LICENSE $dir/
	tar -zcvf $dir.tar.gz $dir
	rm -r $dir
done

# Compute the sums.
shasum *.tar.gz > SHASUMS.txt

cat SHASUMS.txt

if [[ $DRONE_BRANCH != $RELEASE_BRANCH ]]; then
	echo "Only publish a release for the $RELEASE_BRANCH branch."
	exit 0
fi

for file in `ls`
do
	aws s3 cp $file s3://newsdev-pub/stocker/$VERSION/$file
done
