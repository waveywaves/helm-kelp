#!/bin/sh

# Copied w/ love from the excellent hypnoglow/helm-s3

kelpPath="$(dirname "$(realpath "$0")")"
cp $kelpPath/bin/kelp $HELM_HOME/plugins/
