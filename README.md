# Hyper os k8s api service

This adapter is designed to act as a gaia api gateway to get all k8s resources and gaia-related resources.

Without the need to post Bearer token.

## how to make swagger files when apis are updated.
`` make swagger``

## How to build this adapter
``make build`` And remember the git version.

## how to build a new docker images?
``docker build -t 172.17.9.231:8880/nightwatcher/nightwatcher:v1.0.5 . `` v1.0.5 here is 
the git commit version.

## how to make chart
``helm package .``  and then upload to harbor.

## how to install `nightwatcher` in global
``helm install nightwatcher nightwatcher/nightwatcher -n gaia-system``

Then enjoy!