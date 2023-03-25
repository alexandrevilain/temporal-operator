# Publishing a new release

This document explains how to publish a new release.

## Steps

First of all, update the `VERSION` file with the desired release version.
For instance:
```
0.7.0
```

Then, update the `CHANGELOG.md` by writing down all merged PRs since the last release.

Then run: `make release`

Once done create a new release branch and submit a Pull Request:

```
git checkout -b release/v0.7.0
git commit -am "Release v0.7.0"
git push origin/release/v0.7.0
gh pr create --title "Release v0.7.0" --base main --head release/v0.7.0
```

Once the PR has been approved and merged, publish a new tag from the `main` branch

```
git tag v0.7.0
git push origin --tags
```