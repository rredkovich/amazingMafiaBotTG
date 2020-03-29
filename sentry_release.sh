#!/bin/bash

export SENTRY_AUTH_TOKEN=084dc30ad82044a98a1667d25f6caf0a6da012487c8a41218ad997af12f16eb1
export SENTRY_DSN=https://9857777662ac40b1afa9148bb0f3ebe2@sentry.dg9.eu/6
export SENTRY_ORG=dg9
export SENTRY_PROJECT=amazing-mafia-bot-tg
export SENTRY_URL=https://sentry.dg9.eu
TAG_VERSION=$(echo v$(date -u +"%Y-%m-%d")_$(git rev-parse --short=7 HEAD))

sentry-cli releases new -p amazing-mafia-bot-tg "$TAG_VERSION"
sentry-cli releases set-commits "$TAG_VERSION" --auto
sentry-cli releases finalize "$TAG_VERSION"
sentry-cli releases deploys "$TAG_VERSION" new -e production