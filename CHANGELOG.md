# s3enum changelog

## Unreleased

- Validate that specified number of workers is positive
- Print usage message to stderr instead of stdout
- Remove duplicate candidate generation logic
- Exclude context cancellation from error stats
- Remove redundant context check in producer loop

## v1.1.0

### Added

- DNS connection pooling for improved throughput
- Retry with exponential backoff on DNS failures
- Run summary printed to stderr on completion

### Changed

- Renamed `-threads` flag to `-workers`
- Increased default number of workers from 10 to 50
- Bumped [github.com/miekg/dns](https://github.com/miekg/dns) to v1.1.65
- Bumped golang.org/x/net to v0.52.0

### Fixed

- Fixed nil pointer panic in DNS resolver
  [\#55](https://github.com/koenrh/s3enum/issues/55)
- Fixed graceful shutdown on interrupt

## v1.0.0

- Replaced the unmaintained [docopt](https://github.com/docopt/docopt.go) package
  with the [flag](https://pkg.go.dev/flag) package from the standard library
- Bumped [github.com/miekg/dns](https://github.com/miekg/dns) to v1.1.49
- Updated the tool to use different heuristics to determine whether a bucket exists

## v0.2.0

- Migrated CI from Travis CI to GitHub Actions [\#27](https://github.com/koenrh/s3enum/pull/27)
- Added GolangCI-Lint [\#28](https://github.com/koenrh/s3enum/pull/28)
- Added GitHub Actions release workflow [\#34](https://github.com/koenrh/s3enum/pull/34)
- Updated github.com/miekg/dns to v1.1.17 [\#35](https://github.com/koenrh/s3enum/pull/35)

## v0.1.0

- Added support for multple name arguments [\#20](https://github.com/koenrh/s3enum/pull/20)
- Changed default name server, and added option to override default name server [\#21](https://github.com/koenrh/s3enum/pull/21)
