#!/usr/bin/env bash

set -o nounset

echo "Packaging:"
rm bin/tarpon.zip
zip -r bin/tarpon.zip bin Procfile
echo "Done"