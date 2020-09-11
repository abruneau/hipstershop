#!/bin/bash -e

# protos are needed in adservice folder for compiling during Docker build.

mkdir -p proto && \
cp ../../pb/demo.proto src/main/proto
