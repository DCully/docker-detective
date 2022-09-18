#!/usr/bin/env bash

set -eu

rm -f docker-detective
cd frontend
npm install
npm run build
mv build ../backend/.
cd ../backend
go build -o ../
rm -rf build
