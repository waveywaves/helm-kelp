#!/bin/sh

# Copied w/ love from the excellent hypnoglow/helm-s3

kustPath="$(dirname "$(realpath "$0")")"
cp $kustPath/bin/kust $HELM_HOME/plugins/
