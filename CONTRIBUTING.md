# Contributing <!-- omit in toc -->

Thank you for your interest in improving this project. Your contributions are
appreciated.

In the following you can find a collection of frequently asked questions and
hopefully good answers.

- [How to setup local dev environment?](#how-to-setup-local-dev-environment)
- [How to release a new version?](#how-to-release-a-new-version)

Also consider taking a look at the development documentation at
[`docs/devel`](docs/devel).

## How to setup local dev environment?

### Pre-commit <!-- omit in toc -->

Tool written in Python used for maintaining Git hooks.

- <https://pre-commit.com/>

Whenever this repository is initially cloned, execute:

```
pre-commit install --install-hooks
pre-commit install --install-hooks --hook-type commit-msg
```

Pre-commit should now run on every commit.

Note that Go must be installed for some hooks to work properly.

Read [`docs/devel/pre-commit.md`](docs/devel/pre-commit.md) for more
information.

## How to release a new version?

Decide for a new release tag. Make sure to follow semantic versioning.

Make sure that the "Unreleased" section in the changelog is up-to-date.

Now move the content of the "Unreleased" section to a new section with an
appropiate tile for the release. Replace the content of the "Unreleased" section
with:

    Nothing.

Commit the changes with a message that follows this pattern:

    chore: Release <tag>

Tag the latest commit.

Push to remote. This will trigger the release workflow.

Check the workflow run. Give attention to the created artifacts like archived
binaries and multiplatform Docker images pushed to Docker Hub. Publish the
release draft in GitHub if everything looks fine.
