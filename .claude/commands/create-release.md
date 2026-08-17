---
name: create-release
description: Creates a versioned git tag and GitHub release with proper structured release notes for the pedantigo library. Use this skill whenever the user asks to cut a release, create a new version, tag a release, publish a version, or bump the version. The skill auto-generates well-formatted changelog notes from commits since the last tag, matching the established "What's Changed" release format.
allowed-tools: Bash, Read, Glob, Grep
---

# create-release command

## Arguments

This command accepts an optional version number argument:

```
/create-release 1.2.0
/create-release v1.2.0
/create-release          # omit to auto-determine version
```

The argument is available as `$ARGUMENTS`. Strip any leading `v` prefix when comparing versions, but always use the `v` prefix when creating tags and releases (e.g. `v1.2.0`).

## Purpose

This command creates a properly tagged and documented GitHub release for pedantigo. It ensures release notes follow the project's established format (matching v1.1.3, v1.1.2, etc.) rather than minimal or auto-generated notes.

The full release chain after this command runs:
1. Tag + release created here → `notify-benchmarks.yml` triggers → benchmarks run → `pedantigo-docs` gets dispatched → versioned docs + `llms-full.txt` regenerated

So getting the release right matters: the release notes become part of the changelog in pedantigo-docs.

## Hard Invariant: All Three Modules Always Release at the Same Version, Together

This repository contains **three separate Go modules**, and every release MUST cover all three, at
the exact same version number, every time — there is no such thing as a partial release:

1. **Root module** — `github.com/SmrutAI/pedantigo/v2` (repo root, tag: `v<version>`)
2. **Echo plugin** — `github.com/SmrutAI/pedantigo/plugins/web/pedantigoecho/v2` (directory: `plugins/web/pedantigoecho/v2/`, tag: `plugins/web/pedantigoecho/v<version>`)
3. **Gin plugin** — `github.com/SmrutAI/pedantigo/plugins/web/pedantigogin/v2` (directory: `plugins/web/pedantigogin/v2/`, tag: `plugins/web/pedantigogin/v<version>`)

Two things must both be true for every release, or it is incomplete:

- **All three modules must be tagged**, using the module-specific tag prefix for the two nested ones.
  A tag at the repo root only ever versions the root module — Go's own module resolution rules require
  a separately-prefixed tag (`<subdirectory>/v<version>`) for every nested module, full stop, with no
  exception. Tagging only the root leaves both plugins permanently uninstallable at that version.
- **Both plugin `go.mod` files must require the same version** of the root module that is being
  released, committed *before* any tag is created. A tag is immutable — if a plugin's `go.mod` still
  points at an older root version at tag time, that mismatch is frozen into the release forever, and a
  fresh `go get` of that plugin can silently resolve an outdated (possibly buggy) root module while the
  consumer believes they installed the latest version.

If a third plugin module is ever added under `plugins/`, it joins this same list automatically — the
loops in Step 5 below use `plugins/*/*/go.mod` and require no manual list to maintain.

## Steps

### 1. Discover current state

```bash
git describe --tags --abbrev=0          # last tag (e.g. v1.1.3)
git log <last-tag>..HEAD --oneline      # commits to include in release
gh release view <last-tag>              # see format of the previous release notes
git tag -l 'plugins/*'                  # existing nested plugin-module tags — check for gaps
```

The last check matters: `plugins/*/go.mod` files are separate Go modules from the repo root. Tagging
the root (`v<version>`) does NOT make `go get .../plugins/web/echo@v<version>` resolve — each nested
module needs its own tag, `plugins/<path>/v<version>`, at the same commit. If a prior release's root
tag has no matching `plugins/*/v<that-version>` tags, that release is currently uninstallable for
anyone depending on a plugin module. Step 5 below tags every nested module automatically so this
can't recur, but check history here in case an earlier release still needs backfilling.

### 2. Determine the new version

If `$ARGUMENTS` was provided, use that as the version (normalise to have a `v` prefix).

If no argument was given, examine the commits since the last tag and suggest:
- `feat(...)` with breaking changes → bump **major**
- `feat(...)` without breaking changes → bump **minor**
- `fix(...)`, `chore(...)`, `docs(...)` only → bump **patch**

Ask the user to confirm the version before proceeding.

### 3. Generate release notes

Do NOT list commits one-per-line. Read every commit since the last tag (subject + body), group them
by what they mean to a consumer of the library, and write a short prose/bullet summary of the major
features and bug fixes. Pure tooling/CI/chore/internal-process commits (e.g. lint config, CI workflow
tweaks, adding this slash command) are still read but generally omitted from the summary unless they
affect consumers.

Format:

```
## What's Changed
* <summary of a feature or group of related commits, in your own words> by @<author>
* <summary of a bug fix> by @<author>
* ...

**Full Changelog**: https://github.com/SmrutAI/pedantigo/compare/<last-tag>...<new-tag>
```

To read the commits to summarize:
```bash
git log <last-tag>..HEAD --format="%h %an %s%n%b"
```

Map GitHub usernames from git author names (for this repo: Tushar Dwivedi → @tushar2708).

If a summarized change has an associated PR, link to the PR:
```
* Added `SchemaLLM` for LLM-friendly JSON schema generation by @tushar2708 in https://github.com/SmrutAI/pedantigo/pull/14
```

The `**Full Changelog**` compare link still covers the complete commit range, so no commit is lost —
only the prose summary is curated, not the underlying diff.

### 4. Show the user the plan

Before executing, show:
- The new version number
- The release notes draft

Get confirmation before creating the tag or release.

### 5. Release all three modules together, in this exact order

This step enforces the invariant above. Do not reorder, skip, or partially execute these sub-steps —
each one depends on the previous one having actually happened.

**5a. Bump both plugin modules' `go.mod` to require the new root version, and commit — before any
tag exists.** This has to come first because a tag is a permanent, immutable pointer to a commit; if
the `go.mod` bump happens after tagging, the mismatch is frozen into the release forever.

```bash
for gomod in plugins/*/*/v2/go.mod; do
  moddir=$(dirname "$gomod")
  sed -i.bak "s#github.com/SmrutAI/pedantigo/v2 v[0-9.]*#github.com/SmrutAI/pedantigo/v2 v<version>#" "$gomod"
  rm "$gomod.bak"
  make -C "$moddir" deps   # go mod tidy, regenerates go.sum for the bumped requirement
done
git add plugins/*/*/v2/go.mod plugins/*/*/v2/go.sum
git commit -m "chore(release): bump plugin modules to require pedantigo v<version>"
```

**5b. Tag the root module**, pointing at the commit from 5a (so the root tag and both plugin
`go.mod` bumps are the same commit):

```bash
git tag -a v<version> -m "Release v<version>"
git push origin v<version>
```

**5c. Tag both plugin modules, at that same commit**, using each module's own path-prefixed tag.
This is mandatory, not optional — skipping it silently breaks installation for that plugin, exactly
as happened with `plugins/web/echo` in `v2.0.0`/`v2.0.1`. The loop below covers every nested module
under `plugins/`, so this never needs to be updated by hand when a new plugin is added:

```bash
for gomod in plugins/*/*/v2/go.mod; do
  moddir=$(dirname "$gomod")
  tagprefix="${moddir%/v2}"
  git tag "$tagprefix/v<version>" "v<version>"
  git push origin "$tagprefix/v<version>"
done
```

**5d. Verify all three tags exist and point at the same commit** before creating the GitHub release:

```bash
git rev-list -n 1 v<version>
git rev-list -n 1 plugins/web/pedantigoecho/v<version>
git rev-list -n 1 plugins/web/pedantigogin/v<version>
# all three commit hashes above must be identical — if not, stop and investigate before continuing
```

**5e. Create the GitHub release** with the notes:
```bash
gh release create v<version> \
  --title "v<version>" \
  --notes "<release notes>"
```

Do NOT use `--generate-release-notes` — always write notes explicitly from the commits to ensure accuracy.

### 6. Confirm

After creation, verify:
```bash
gh release view v<version>
```

Show the user the release URL so they can confirm it looks correct.

## Notes

- The `notify-benchmarks.yml` workflow triggers automatically on `release: published` — no manual action needed after this.
- Do NOT edit release notes of past versions — create a new release instead.
- If the user wants to include breaking changes in the release notes, add a `## Breaking Changes` section above `## What's Changed`.
