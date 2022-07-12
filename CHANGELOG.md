# s3enum changelog

## Unreleased

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
