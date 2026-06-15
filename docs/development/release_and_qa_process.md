# Release and QA Process

This page describes how the NVIDIA Infra Controller (NICo) project is branched,
versioned, tested, and released. It is intended for both contributors and
operators who want to understand which version of NICo they should be running.

## Where Releases Live

- **GitHub releases:** <https://github.com/NVIDIA/infra-controller/releases>
- **Issue tracker:** <https://github.com/NVIDIA/infra-controller/issues>
- **Source:** <https://github.com/NVIDIA/infra-controller>

Every published minor and patch release is available on the GitHub releases page
above, tagged with its semver version (see [Tag Naming](#tag-naming) below).

## Branches

NICo uses a small number of long-lived branches together with semver tags. The
table below summarizes them; each is described in detail in the following
sections.

| Branch                    | Purpose                                     | Stability                  |
|---------------------------|---------------------------------------------|----------------------------|
| `main`                    | Ongoing development                         | No stability guarantee     |
| `releases/vX.Y-rc`        | Release candidate, under active QA          | Improving, may be broken   |
| `releases/vX.Y`           | QA signed-off release branch                | Stable, recommended        |

### `main` — Ongoing Development

All changes land on `main` first. There is **no expectation of stability** on
`main`; it is not QA tested. The only tests that gate changes to `main` are the
automated tests that run in CI. Features may be incomplete and bugs may be
present at any commit.

Use `main` if you want early access to in-progress features and you accept that
things will sometimes be broken.

### `releases/vX.Y-rc` — Release Candidate Branches

When development for a minor version is feature-complete, a new long-lived
release branch is cut from `main`. While QA is in progress, this branch carries
an `-rc` suffix (for example, `releases/v2.1-rc`). "rc" stands for **release
candidate**: the code is being stabilized and exercised by QA, but has not yet
been signed off.

Use the latest commit on a release candidate branch if you want code that is a
little more stable than `main` (more so as QA progresses), with the
understanding that QA may still uncover problems. Use at your own risk.

### `releases/vX.Y` — QA Signed-Off Release Branches

Once QA has signed off on a release candidate, the `-rc` suffix is dropped and
the branch becomes the long-lived release branch for that minor version (for
example, `releases/v2.1`). All subsequent patch releases for that minor version
are made from this branch.

This branch — specifically, its latest tagged commit — is what most users should
deploy. See [Tag Naming](#tag-naming) below.

### Tag Naming

NICo uses [semantic versioning](https://semver.org/) of the form `vX.Y.Z`:

- `X` — major version
- `Y` — minor version
- `Z` — patch version

The following tag forms appear in the repository:

- **`vX.Y.0`** — A minor release. Published as a GitHub release from
  `releases/vX.Y`.
- **`vX.Y.Z`** (where `Z > 0`) — A patch release on top of `vX.Y.0`. Also
  published as a GitHub release from `releases/vX.Y`.
- **`vX.Y-rc`** — Applied to commits on the `-rc` branch during the QA period.
  Not a final release.
- **`vX.Y-pr`** — Applied to `main` immediately after a release branch is cut,
  to indicate that `main` is now the **prerelease** for the next minor version.
  For example, the day `releases/v2.1` is cut, `main` is tagged `v2.2-pr`,
  signaling that `main` is now pre-v2.2.

## Release Cadence

NICo follows a fixed quarterly cadence with a one-month QA window.

### Minor Releases (`X.Y.0`)

Every quarter:

1. **Code complete** (last day of December, March, June, September): a new
   release branch (e.g. `releases/v2.1-rc`) is cut from `main`.
2. Immediately after the cut, `main` is tagged with `vX.(Y+1)-pr` to mark the
   start of the next prerelease cycle on `main`.
3. The release candidate branch is **stabilized and QA tested for one month**.
4. **Final minor release** (last day of January, April, July, October): when QA
   signs off, the `-rc` suffix is dropped, the branch becomes `releases/vX.Y`,
   and a `vX.Y.0` tag is cut and published as a GitHub release.

In short: minor releases ship one month after code complete.

### Patch Releases (`X.Y.Z`)

Patch releases happen on the `releases/vX.Y` branch after the corresponding
`vX.Y.0` has shipped. Patch releases are cut as needed, but at most **once per
month**, on the last day of the month. Each patch release is published on
GitHub with a `vX.Y.Z` tag.

## Which Version Should I Use?

| Goal                                                    | What to run                                                |
|---------------------------------------------------------|------------------------------------------------------------|
| Early access to in-progress features                    | Latest `main`                                              |
| Slightly more stable, willing to help shake out bugs    | Latest commit on `releases/vX.Y-rc`                        |
| Most stable, production-style use                       | Latest non-`-rc`, non-`-pr` tag on `releases/vX.Y`         |

Bugs found on a tagged release (`vX.Y.Z` with no `-rc` or `-pr` suffix) are
treated with the highest priority and are tracked as **QA test escapes** —
defects that slipped past the QA window and require a follow-up fix, typically
in the next patch release.

## Backward Compatibility

Breaking changes are **not allowed** anywhere in the codebase for anything that
falls under our API guarantees.

### What Is Guaranteed to Remain Backward Compatible

- The **NICo REST API**.
- The **NICo CLI** (`nicocli`) — command names, arguments, flags, values, and
  exit codes.
- **Configuration file structures** — keys, values, filenames, and locations.
- **Environment variable names and values** consumed by NICo components.

If you depend on any of the above, you can rely on them not changing
incompatibly within and across releases.

### What Is Explicitly *Not* Guaranteed

The following are considered internal and may change without notice between
releases:

- The **gRPC API** and protobuf message contents.
- The **admin CLI** (also referred to as the *debug CLI*) — a lower-level tool
  intended for operators and developers, not end users.
- The **admin UI** (also referred to as the *debug UI*) — same audience as the
  admin CLI.
- The **Vault data model** — how secrets are laid out inside HashiCorp Vault.
- The **PostgreSQL database schema** used by NICo services. See the
  [tracking issue for database backward compatibility](https://github.com/NVIDIA/infra-controller/issues)
  on GitHub for the current state of this guarantee.
- Any other internal API contract between NICo services, or persistent data
  formats used only by NICo itself.

If you build automation that depends on any of the unguaranteed items above,
expect to update it across NICo releases.

## Glossary

A few terms used on this page that may not be obvious:

- **Code complete** — the point in the cycle at which feature work for a minor
  version stops and stabilization begins. On this date, the release branch is
  cut from `main`.
- **Release candidate (rc)** — a build or branch that is a candidate for
  release, pending QA sign-off. Identified by the `-rc` suffix on the branch
  name.
- **Prerelease (pr)** — a build of `main` that is on its way to becoming the
  next minor release. Identified by the `-pr` suffix on a tag (e.g.
  `v2.2-pr`).
- **QA sign-off** — the formal acknowledgment from QA that a release candidate
  has passed its test plan and may be promoted to a final release.
- **QA test escape** — a defect discovered in a tagged, signed-off release that
  was not caught during the QA window. These are treated as high-priority and
  typically fixed in a subsequent patch release.
- **Semver** — [semantic versioning](https://semver.org/), the `vX.Y.Z` scheme
  used by NICo where `X` is major, `Y` is minor, and `Z` is patch.
