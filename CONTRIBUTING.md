# Contributing <!-- omit in toc -->

Thank you for your interest in improving this project. Your contributions are
appreciated.

In the following you can find a collection of frequently asked questions and
hopefully good answers.

- [How to setup local dev environment?](#how-to-setup-local-dev-environment)

Also consider taking a look at the development documentation at
[`docs/devel`](docs/devel).

## How to setup local dev environment?

### Pre-commit <!-- omit in toc -->

Tool written in Python used for maintaining Git hooks. Must be installed
beforehand.

- <https://pre-commit.com/>
- <https://github.com/pre-commit/pre-commit>

Whenever this repository is initially cloned, the following should be executed:

```
pre-commit install --install-hooks
pre-commit install --install-hooks --hook-type commit-msg
```

Pre-commit should now run on every commit.

Read [`docs/devel/pre-commit.md`](docs/devel/pre-commit.md) for more
information.
