#!/bin/bash

###
# Usage: VERSION=1.14.0 FORK_REPOSITORY=git@github.com:alexandrevilain/community-operators.git ./operatorhub.sh

if [ -z "$VERSION" ]; then
    >&2 echo "Please set VERSION variable"
    exit 1
fi

if [ -z "$FORK_REPOSITORY" ]; then
    >&2 echo "Please set VERSION variable"
    exit 1
fi

# Create a temporary directory
TEMPD=$(mktemp -d)

# Exit if the temp directory wasn't created successfully.
if [ ! -e "$TEMPD" ]; then
    >&2 echo "Failed to create temp directory"
    exit 1
else 
    echo "Working directory: $TEMPD"
fi

# trap "exit 1"           HUP INT PIPE QUIT TERM
# trap 'rm -rf "$TEMPD"'  EXIT

cd $TEMPD

echo "Getting PR template ..."
curl https://raw.githubusercontent.com/k8s-operatorhub/community-operators/main/docs/pull_request_template.md -o operatorhub-pr-template.md
vim operatorhub-pr-template.md

echo "Cloning repositories ..."
git clone --depth 1 --branch "v$VERSION" https://github.com/alexandrevilain/temporal-operator.git
git clone --depth 1 https://github.com/k8s-operatorhub/community-operators.git

echo "Adding files ..."
mkdir -p community-operators/operators/temporal-operator/$VERSION
cp -R temporal-operator/bundle/ community-operators/operators/temporal-operator/$VERSION

cd community-operators
git remote add fork $FORK_REPOSITORY
git checkout -b update-temporal-operator-to-$VERSION

echo "Please check diff in $TEMPD ..."
read -p "Press enter to continue"

git add .
git commit -sm "Update temporal-operator to $VERSION"
git push fork update-temporal-operator-to-$VERSION --force-with-lease

gh pr create --title "operator temporal-operator (${VERSION})" --body-file "$TEMPD/operatorhub-pr-template.md" --head update-temporal-operator-to-$VERSION