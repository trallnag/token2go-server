# Pre-Commit

Used for maintaining Git hooks. Must be installed on development system. As it
is written in Python, for example [`pipx`](https://github.com/pypa/pipx) can be
used to install it.

- <https://pre-commit.com/>
- <https://github.com/pre-commit/pre-commit>

Whenever this repository is initially cloned, the following should be executed:

```
pre-commit install --install-hooks
pre-commit install --install-hooks --hook-type commit-msg
```

Pre-commit should now run on every commit. It is also used in GitHub Actions.

Pre-commit is configured via
[`.pre-commit-config.yaml`](../../.pre-commit-config.yaml).

Note that Go must be installed for some hooks to work properly.

## Cheat Sheet

Run pre-commit against all files.

```
pre-commit run -a
```

Run specific hook against all files.

```
pre-commit run <hook> -a
```

## Housekeeping

Update hooks in general.

```
pre-commit autoupdate
```

Update local hook `shfmt` additional dependency by adjusting the version. Check
for new versions [here](https://github.com/mvdan/sh).
