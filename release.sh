#!/bin/bash
set -e

VERSION=v0.2.0
RELEASE_BRANCH=master

if [[ $DRONE_BRANCH != $RELEASE_BRANCH ]]; then
	echo "Only publish a release for the $RELEASE_BRANCH branch."
	exit 0
fi

gox -output="build/stocker-$VERSION-{{.OS}}-{{.Arch}}/bin/stocker" -os="linux darwin"

cd build

for dir in `ls`
do
	cp ../README.md ../LICENSE $dir/
	tar -zcvf $dir.tar.gz $dir
	rm -r $dir
done

shasum *.tar.gz > SHASUMS.txt

for file in `ls`
do
	aws s3 cp $file s3://newsdev-pub/stocker/$VERSION/$file
done
