# s3enum

s3enum is a fast and stealthy Amazon S3 bucket enumeration tool. It leverages DNS
instead of HTTP, which means it does not hit AWS infrastructure directly.

It was originally built back in 2016 to [target GitHub](https://koen.io/2016/02/13/github-bug-bounty-hunting/).

## Installation

```console
go install github.com/koenrh/s3enum@v1
```

## Usage

You need to specify the base name of the target (e.g., `hackerone`), and a word list.
You could either use the example [`wordlist.txt`](examples/wordlist.txt) file from
this repository, or get a word list [elsewhere](https://github.com/bitquark/dnspop/tree/master/results).
Optionally, you could specify the number of threads (defaults to 5).

```
$ s3enum -wordlist examples/wordlist.txt -suffixlist examples/suffixlist.txt -threads 10 hackerone

hackerone
hackerone-attachment
hackerone-attachments
hackerone-static
hackerone-upload
```

By default, `s3enum` will use the name server as specified in `/etc/resolv.conf`.
Alternatively, you could specify a different name server using the `-nameserver`
option. Besides, you could test multiple names at the same time.

```
s3enum \
  -wordlist examples/wordlist.txt \
  -suffixlist examples/suffixlist.txt \
  -nameserver 1.1.1.1 \
  hackerone h1 roflcopter
```

## Known limitations

s3enum is currently unable to detect S3 buckets in the us-east-1 region.
