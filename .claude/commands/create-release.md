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

## Steps

### 1. Discover current state

```bash
git describe --tags --abbrev=0          # last tag (e.g. v1.1.3)
git log <last-tag>..HEAD --oneline      # commits to include in release
gh release view <last-tag>              # see format of the previous release notes
```

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

### 5. Create the tag and release

```bash
git tag -a v<version> -m "Release v<version>"
git push origin v<version>
```

Then create the GitHub release with the notes:
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
