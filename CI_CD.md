# CI/CD Pipeline

This document describes the GitHub Actions workflows across the Pedantigo ecosystem.

## Repository Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              3 REPOSITORIES                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│  pedantigo              │  pedantigo-benchmarks   │  pedantigo-docs         │
│  (Go library)           │  (Benchmark runner)     │  (Docusaurus site)      │
│                         │                         │                         │
│  Workflows:             │  Workflows:             │  Workflows:             │
│  - ci.yml               │  - benchmark.yml        │  - deploy.yml           │
│  - notify-benchmarks.yml│  - benchmark-pr-self.yml│                         │
│  - notify-benchmarks-   │  - benchmark-pr-        │                         │
│      pr.yml             │      external.yml       │                         │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Workflow 1: CI (`ci.yml`)

**Triggers:**
- Push to `main` branch
- Pull request to `main` branch

**Job: test**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CI WORKFLOW                                     │
│                                                                             │
│  Trigger: push/PR to main                                                   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Step 1: Checkout (fetch-depth: 0 for gitleaks)                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Step 2: Setup Go 1.21                                               │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Step 3: Install tools                                               │   │
│  │   - go-junit-report                                                 │   │
│  │   - goimports                                                       │   │
│  │   - golangci-lint v1.62.2                                           │   │
│  │   - pre-commit                                                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Step 4: Run pre-commit checks (all files)                           │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Step 5: Run tests with coverage (make test-ci-cov)                  │   │
│  │   Outputs: test-results.xml, coverage.out                           │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Step 6: Upload test results (always runs)                           │   │
│  │   - Publishes to PR as check "Test Results"                         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Step 7: Generate coverage report                                    │   │
│  │   - Thresholds: 80% warning, 90% good                               │   │
│  │   - Fails if below 80%                                              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                          ┌────────┴────────┐                                │
│                          │                 │                                │
│                     (PR only)         (main only)                           │
│                          │                 │                                │
│                          ▼                 ▼                                │
│  ┌──────────────────────────────┐  ┌──────────────────────────────────┐   │
│  │ Step 8a: Add coverage        │  │ Step 8b: Extract coverage %      │   │
│  │ comment to PR                │  │ Step 9: Update coverage badge    │   │
│  │ (sticky-pull-request-comment)│  │ (dynamic-badges-action → Gist)   │   │
│  └──────────────────────────────┘  └──────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Secrets Required:**
- `GIST_TOKEN` - For updating coverage badge
- `GIST_ID` - Gist containing `pedantigo-coverage.json`

---

## Workflow 2: Notify Benchmarks (`notify-benchmarks.yml`)

**Triggers:**
- Push to `main` branch
- Release published

**Purpose:** Triggers benchmark run. On release, passes version for automatic doc versioning.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        NOTIFY BENCHMARKS WORKFLOW                            │
│                                                                             │
│  Trigger: push to main OR release published                                 │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Step 1: Trigger benchmark run                                       │   │
│  │   - Uses: peter-evans/repository-dispatch@v3                        │   │
│  │   - Target: SmrutAI/pedantigo-benchmarks                            │   │
│  │   - Event: "pedantigo-updated"                                      │   │
│  │   - Payload: {"version": "<tag_name>"} (on release only)            │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│                    ┌───────────────────────────────┐                        │
│                    │ pedantigo-benchmarks receives │                        │
│                    │   "pedantigo-updated" event   │                        │
│                    │   + version in payload        │                        │
│                    └───────────────────────────────┘                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Secrets Required:**
- `BENCHMARK_TRIGGER_TOKEN` - PAT with `repo` scope for cross-repo dispatch

---

## Workflow 3: Benchmark PR (`notify-benchmarks-pr.yml`)

**Triggers:**
- Pull request to `main` branch

**Purpose:** Triggers benchmark run for PRs. Results are posted as a comment on the PR (not committed to docs).

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      BENCHMARK PR WORKFLOW (pedantigo)                       │
│                                                                             │
│  Trigger: PR to main                                                        │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ Step 1: Trigger benchmark for PR                                    │   │
│  │   - Uses: peter-evans/repository-dispatch@v3                        │   │
│  │   - Target: SmrutAI/pedantigo-benchmarks                            │   │
│  │   - Event: "pedantigo-pr"                                           │   │
│  │   - Payload: {pr_number, branch, sha}                               │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│                    ┌───────────────────────────────┐                        │
│                    │ pedantigo-benchmarks receives │                        │
│                    │   "pedantigo-pr" event        │                        │
│                    │   + PR info in payload        │                        │
│                    └───────────────────────────────┘                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Secrets Required:**
- `BENCHMARK_TRIGGER_TOKEN` - Same token used for main branch dispatch

---

## Workflow 4: Benchmark PR Self (`benchmark-pr-self.yml`)

**Repository:** pedantigo-benchmarks

**Triggers:**
- Pull request to `main` branch (in pedantigo-benchmarks itself)

**Purpose:** Run benchmarks when the benchmark code itself changes. Comments results on own PR.

**Branch Logic:**
- Benchmarks: PR branch (automatic via checkout)
- Pedantigo: main branch (cloned)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                  BENCHMARK PR SELF WORKFLOW (pedantigo-benchmarks)           │
│                                                                             │
│  Trigger: PR to main in pedantigo-benchmarks                                │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 1. Checkout benchmarks (PR branch - automatic)                      │   │
│  │ 2. Clone pedantigo (main branch)                                    │   │
│  │ 3. Run benchmarks                                                   │   │
│  │ 4. Comment BENCHMARK.md on own PR                                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Workflow 5: Benchmark PR External (`benchmark-pr-external.yml`)

**Repository:** pedantigo-benchmarks

**Triggers:**
- `repository_dispatch` with event type: `pedantigo-pr`

**Purpose:** Run benchmarks for PRs in pedantigo. Comments results back on the pedantigo PR.

**Branch Logic:**
- Benchmarks: main branch (automatic via checkout)
- Pedantigo: PR branch (from payload)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│               BENCHMARK PR EXTERNAL WORKFLOW (pedantigo-benchmarks)          │
│                                                                             │
│  Trigger: repository_dispatch "pedantigo-pr"                                │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ 1. Checkout benchmarks (main branch - automatic)                    │   │
│  │ 2. Clone pedantigo (PR branch from payload)                         │   │
│  │ 3. Run benchmarks                                                   │   │
│  │ 4. Comment BENCHMARK.md on pedantigo PR                             │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Secrets Required:**
- `PEDANTIGO_PR_TOKEN` - PAT with `repo` scope to comment on pedantigo PRs

---

## Cross-Repository Event Flow

```
                              PUSH TO MAIN / RELEASE
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              pedantigo repo                                  │
│                                                                             │
│  ┌──────────────┐    ┌────────────────────────┐                             │
│  │   ci.yml     │    │ notify-benchmarks.yml  │                             │
│  │              │    │                        │                             │
│  │ Tests +      │    │ Dispatch:              │                             │
│  │ Coverage     │    │ "pedantigo-updated"    │                             │
│  │              │    │ + version (if release) │                             │
│  └──────────────┘    └───────────┬────────────┘                             │
│                                  │                                           │
└──────────────────────────────────┼───────────────────────────────────────────┘
                                   │
                                   ▼
              ┌──────────────────────────────┐
              │   pedantigo-benchmarks       │
              │                              │
              │  Listens for:                │
              │  - pedantigo-updated         │
              │                              │
              │  Actions:                    │
              │  - Run benchmarks            │
              │  - Generate BENCHMARK.md     │
              │  - Commit to repo            │
              │  - Dispatch to docs:         │
              │    "benchmarks-updated"      │
              │    + version (passed through)│
              └──────────────┬───────────────┘
                             │
                             ▼
              ┌──────────────────────────────┐
              │       pedantigo-docs         │
              │                              │
              │  Listens for:                │
              │  - benchmarks-updated        │
              │                              │
              │  Actions:                    │
              │  - Clone pedantigo           │
              │  - Clone pedantigo-benchmarks│
              │  - Copy BENCHMARK.md         │
              │  - If version: create        │
              │    versioned docs            │
              │  - Build Docusaurus          │
              │  - Deploy to GitHub Pages    │
              └──────────────────────────────┘
```

---

## PR Benchmark Flow

### PR in pedantigo (branch X)

```
pedantigo PR (branch X)
    │
    └── notify-benchmarks-pr.yml
            │
            └── Dispatch: pedantigo-pr
                payload: {pr_number, branch: X, sha}
                    │
                    └── pedantigo-benchmarks
                        benchmark-pr-external.yml
                            │
                            ├── Checkout benchmarks: main
                            ├── Clone pedantigo: branch X
                            ├── Run benchmarks
                            └── Comment on pedantigo PR #N
```

### PR in pedantigo-benchmarks (branch Y)

```
pedantigo-benchmarks PR (branch Y)
    │
    └── benchmark-pr-self.yml
            │
            ├── Checkout benchmarks: branch Y (automatic)
            ├── Clone pedantigo: main
            ├── Run benchmarks
            └── Comment on own PR
```

---

## pedantigo-docs Deploy Workflow

**Triggers:**
- Push to `main` branch (own repo changes)
- `repository_dispatch` with event type: `benchmarks-updated`
- `workflow_dispatch` (manual trigger)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    PEDANTIGO-DOCS DEPLOY WORKFLOW                            │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                           BUILD JOB                                   │  │
│  │                                                                       │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Step 1: Checkout pedantigo-docs                                │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                              │                                        │  │
│  │                              ▼                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Step 2: Clone pedantigo (--depth 1)                            │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                              │                                        │  │
│  │                              ▼                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Step 3: Clone pedantigo-benchmarks (--depth 1)                 │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                              │                                        │  │
│  │                              ▼                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Step 4: Copy benchmark report                                  │  │  │
│  │  │   pedantigo-benchmarks/BENCHMARK.md → pedantigo/docs/benchmarks│  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                              │                                        │  │
│  │                              ▼                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Step 5: Setup Node.js 20 + npm ci                              │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                              │                                        │  │
│  │                              ▼                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Step 6: Create versioned docs (on release only)                │  │  │
│  │  │   - If version in payload: npm run docusaurus docs:version     │  │  │
│  │  │   - Strips 'v' prefix (v0.2.0 → 0.2.0)                         │  │  │
│  │  │   - BENCHMARK.md already copied → included in version!         │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                              │                                        │  │
│  │                              ▼                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Step 7: Commit versioned docs (on release only)                │  │  │
│  │  │   - git add versioned_docs versioned_sidebars versions.json    │  │  │
│  │  │   - git commit + push to pedantigo-docs                        │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                              │                                        │  │
│  │                              ▼                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Step 8: npm run build                                          │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                              │                                        │  │
│  │                              ▼                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Step 9: Upload pages artifact (build/)                         │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                       │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                 │                                           │
│                                 ▼                                           │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                          DEPLOY JOB                                   │  │
│  │                                                                       │  │
│  │  needs: build                                                         │  │
│  │  environment: github-pages                                            │  │
│  │                                                                       │  │
│  │  ┌────────────────────────────────────────────────────────────────┐  │  │
│  │  │ Deploy to GitHub Pages → https://pedantigo.dev                 │  │  │
│  │  └────────────────────────────────────────────────────────────────┘  │  │
│  │                                                                       │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Event Timelines

### Push to main (no tag)

```
t+0s   Developer pushes to pedantigo:main
       └── notify-benchmarks.yml dispatches to pedantigo-benchmarks
           payload: {"version": ""}

t+1s   pedantigo-benchmarks receives event
       └── Runs benchmarks, commits BENCHMARK.md
       └── Dispatches to pedantigo-docs
           payload: {"version": ""}

t+2m   pedantigo-docs receives event
       └── version is empty → skips versioning step
       └── Builds "next" docs with fresh benchmarks
       └── Deploys to https://pedantigo.dev
```

### Create release v0.2.0

```
t+0s   Developer publishes release v0.2.0 on pedantigo
       └── notify-benchmarks.yml dispatches to pedantigo-benchmarks
           payload: {"version": "v0.2.0"}

t+1s   pedantigo-benchmarks receives event
       └── Runs benchmarks, commits BENCHMARK.md
       └── Dispatches to pedantigo-docs
           payload: {"version": "v0.2.0"}

t+2m   pedantigo-docs receives event
       └── version is "v0.2.0" → creates versioned_docs/version-0.2.0/
       └── Commits versioned docs
       └── Builds and deploys
       └── /docs/0.2.0/benchmarks available with fresh data
```

---

## Secrets Summary

| Secret | Repository | Purpose |
|--------|------------|---------|
| `GIST_TOKEN` | pedantigo | Update coverage badge gist |
| `GIST_ID` | pedantigo | Gist ID for coverage badge |
| `BENCHMARK_TRIGGER_TOKEN` | pedantigo | Dispatch to pedantigo-benchmarks |
| `DOCS_TRIGGER_TOKEN` | pedantigo-benchmarks | Dispatch to pedantigo-docs |
| `PEDANTIGO_PR_TOKEN` | pedantigo-benchmarks | Comment on pedantigo PRs |

---

## Local Development

For local development, use `setup.sh` in pedantigo-docs which replicates the CI behavior:

```bash
./setup.sh
# Clones pedantigo → pedantigo/
# Clones pedantigo-benchmarks → pedantigo-benchmarks/
# Copies BENCHMARK.md → pedantigo/docs/benchmarks.md

npm start
# Starts Docusaurus dev server
```
