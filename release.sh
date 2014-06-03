#!/bin/bash
set -e

VERSION=v0.4.0
RELEASE_BRANCH=master

# Add the source to the build directory.
mkdir -p build/stocker-$VERSION
rsync -av --exclude .git --exclude-from=.gitignore ./ build/stocker-$VERSION/

# Build the binaries.
gox -output="build/stocker-$VERSION-{{.OS}}-{{.Arch}}/bin/stocker" -os="linux darwin"

# Add the README and LICENSE to the binary directories.
cd build
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
