# s3enum

[![Build Status](https://travis-ci.com/koenrh/s3enum.svg?branch=master)](https://travis-ci.com/koenrh/s3enum)

s3enum is a fast enumeration tool built to enumerate a target's Amazon S3 buckets.
It helps security researchers, and penetration testers to collect bucket names
for further inspection. This tool uses DNS instead of HTTP, which means it doesn't
hit Amazon's infrastructure (directly).

## Installation

```bash
go get -u github.com/koenrh/s3enum
```

## Usage

You need to specify the base name of the target (e.g. `hackerone`), and a word list.
You could either use the `words.txt` file from this repository, or get a word list
[elsewhere](https://github.com/bitquark/dnspop/tree/master/results). Optionally,
you could specify the number of threads (defaults to 10).

```
$ s3enum -w words.txt -n hackerone

hackerone-attachment
hackerone-attachments
hackerone-static
hackerone-upload
```
