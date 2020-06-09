#!/bin/sh

set -x

curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"uid":"p1-74cbdcda-bdc3-4fe3-8602-fbaac01689cc","secret":"4FAAA42E3DEB4C4F0AD20CC9A2A441F400B0A3DD0E57C7FB33EA73D7BFA966BB"}' \
  http://localhost:5000/rooms/aaa3ff11-9ff3-44b8-ab95-b2f339fb9765/peers

curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"uid":"p2-af868c84-ab5a-4835-8503-93f295068f98","secret":"88BDA59097E5840A25C2E7B442E88C7790C508F4C759E82047F9637DA6ACB2C5"}' \
  http://localhost:5000/rooms/aaa3ff11-9ff3-44b8-ab95-b2f339fb9765/peers

curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"uid":"p3-7977c6f9-16d6-4b35-91ec-72f81003914c","secret":"5A3EDF2142FFDE0B2D9803D845C795C24BFDD610D2B9D68408F5207D47E11B4A"}' \
  http://localhost:5000/rooms/aaa3ff11-9ff3-44b8-ab95-b2f339fb9765/peers
