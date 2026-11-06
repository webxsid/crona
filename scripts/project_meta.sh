#!/usr/bin/env sh

PROJECT_NAME="crona"
# PROJECT_OWNER owns the GitHub repository, the source of truth for releases; keep it in
# step with RepoOwner in shared/version/version.go (a test checks). PACKAGE_OWNER owns the
# Homebrew tap and Scoop bucket, which stay on the original account so existing taps keep
# their names; keep it in step with PackageOwner.
PROJECT_OWNER="crona-app"
PACKAGE_OWNER="webxsid"
PROJECT_REPO="${PROJECT_OWNER}/crona"
PROJECT_BIN_PREFIX="crona"
