#!/usr/bin/env sh

PROJECT_NAME="crona"
# PROJECT_OWNER owns the GitHub repository, the source of truth for releases; keep it in
# step with RepoOwner in shared/version/version.go (a test checks). PACKAGE_OWNER owns the
# Homebrew tap and Scoop bucket and follows it until those move.
PROJECT_OWNER="webxsid"
PACKAGE_OWNER="${PROJECT_OWNER}"
PROJECT_REPO="${PROJECT_OWNER}/crona"
PROJECT_BIN_PREFIX="crona"
