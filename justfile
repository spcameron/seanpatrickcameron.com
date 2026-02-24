# print this help message
[default]
help:
    @just --list

# runs audit & tests
[group('*workflow')]
check:
    @scripts/audit 
    @scripts/test unit

# runs audit & tests -race
[group('*workflow')]
ci:
    @scripts/audit
    @scripts/test race
    @scripts/build-site

# runs tidy, gofmt, and go-mod-upgrade
[group('*workflow')]
maintain:
    @scripts/tidy
    @scripts/mod-upgrade

# runs a chain of QC commands
[group('quality')]
audit:
    @scripts/audit

# runs tidy & gofmt
[group('quality')]
tidy:
    @scripts/tidy

# runs go-mod-upgrade
[group('quality')]
mod-upgrade:
    @scripts/mod-upgrade

# runs tests, accepts commands [unit|race|cover]
[group('test')]
test mode="":
    @scripts/test {{mode}}

# builds command binary with native target
[group('build')]
build:
    @scripts/build

# builds command binary with linux_amd64 target
[group('build')]
build-linux:
    @scripts/build linux_amd64

# builds the site pages
[group('build')]
build-site:
    @scripts/build-site

# builds command binary, builds the site pages
[group('build')]
build-all:
    @scripts/build
    @scripts/build-site

# build site output and serve locally (single run)
[group('run')]
run *args="":
    @scripts/serve --build -- {{args}}

# start live development server (auto-rebuild on change)
[group('run')]
run-live:
    @air -c .air.toml

# sync main and delete local branch (for branches with no PR)
[group('git')]
branch-delete:
    @scripts/git/branch-delete

# sync main and create a new branch
[group('git')]
branch-create:
    @scripts/git/branch-create

# create PR for current branch
[group('git')]
pr-create:
    @scripts/git/pr-create

# merge PR for local branch
[group('git')]
pr-merge:
    @scripts/git/pr-merge

# merge PR for local branch, sync main, and delete local branch
[group('git')]
pr-finish:
    @scripts/git/pr-merge --cleanup

# view open PR in browser
[group('git')]
pr-view:
    @scripts/git/pr-view

# push local branch and set upstream to 'origin'
[group('git')]
push-upstream:
    @scripts/git/push-upstream

# rebase branch onto origin/main
[group('git')]
rebase-main:
    @scripts/git/rebase-main

# rebase branch onto upstream
[group('git')]
rebase-upstream:
    @scripts/git/rebase-upstream

# backup current main and reset main to origin/main
[group('git')]
repair-main:
    @scripts/git/repair-main

# rebase branch onto upstream then origin/main, audit, and publish (force-with-lease)
[group('git')]
sync:
    @scripts/git/sync-branch

# fast-forward main from origin/main
[group('git')]
sync-main:
    @scripts/git/sync-main

