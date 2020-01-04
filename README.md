I have archived this project, because Amazon killed the DNS feature this tool used to enumerate S3 bucket names. For details, read [this GitHub issue](https://github.com/koenrh/s3enum/issues/45).

---

# s3enum

![](https://github.com/koenrh/s3enum/workflows/build/badge.svg)

s3enum is a tool to enumerate a target's Amazon S3 buckets. It is fast and leverages
DNS instead of HTTP, which means that requests don't hit AWS directly.

It was originally built back in 2016 to [target GitHub](https://koen.io/2016/02/13/github-bug-bounty-hunting/).

## Installation

### Binaries

Find the binaries on the [Releases page](https://github.com/koenrh/s3enum/releases).

### Go

```console
go get github.com/koenrh/s3enum
```

## Usage

You need to specify the base name of the target (e.g. `hackerone`), and a word list.
You could either use the example [`wordlist.txt`](examples/wordlist.txt) file from
this repository, or get a word list [elsewhere](https://github.com/bitquark/dnspop/tree/master/results).
Optionally, you could specify the number of threads (defaults to 10).

```
$ s3enum --wordlist examples/wordlist.txt --suffixlist examples/suffixlist.txt --threads 10 hackerone

hackerone
hackerone-attachment
hackerone-attachments
hackerone-static
hackerone-upload
```

By default `s3enum` will use the name server as specified in `/etc/resolv.conf`.
Alternatively, you could specify a different name server using the `--nameserver`
option. Besides, you could test multiple names at the same time.

```
s3enum \
  --wordlist examples/wordlist.txt \
  --suffixlist examples/suffixlist.txt \
  --nameserver 1.1.1.1 \
  hackerone h1 roflcopter
```
