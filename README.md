# 127

*127* is a tool for mapping easy-to-remember hostnames to random [loopback
addresses](https://en.wikipedia.org/wiki/Localhost#Name_resolution).

The tool works as a front-end to the standard
[hosts-file](https://en.wikipedia.org/wiki/Hosts_(file)) on your system. It has
been tested on Linux, macOS and Windows.

## Installation

```
go get github.com/lende/127
```

You may also [download a binary release](https://github.com/lende/127/releases).

## Usage and options

```console
$ 127 -h
127 is a tool for mapping hostnames to random loopback addresses.

Usage: 127 [option ...] [hostname[:port]] [operation]

Prints an unassigned random IP if hostname is left out.

The operations are:

  set
        map hostname to random IP and print IP address (default)
  get
        print IP address associated with hostname
  remove
        remove hostname mapping

Options:

  -block string
        address block (default "127.0.0.0/8")
  -hosts string
        path to hosts file (default "/etc/hosts")
  -n    do not output a trailing newline
  -version
        print version information
```

### Notes

* `set` is the default operation and may be omitted
* When a hostname is already mapped, `set` will simply get its IP
* There is also a [code library](https://godoc.org/github.com/lende/127/lib)
  provided

## Examples

### A simple session

Let's map a hostname to a random loopback address:

```console
$ sudo 127 example.test
127.2.221.30
```

Running the command again simply returns the same IP address:

```console
$ 127 example.test
127.2.221.30
```

We can check that it works by pinging the new host:

```console
$ ping example.test
PING example.test (127.2.221.30) 56(84) bytes of data.
64 bytes from example.test (127.2.221.30): icmp_seq=1 ttl=64 time=0.042 ms
```

### Testing a third party service

Let's say you want to try out [ownCloud](https://owncloud.org/). Simply run:

```
docker run -d -p `sudo 127 -n owncloud.test:80`:80 owncloud
```

... and your *ownCloud* instance should be available at http://owncloud.test.

## Troubleshooting

* If there is a `secure_path`-entry in your `/etc/sudoers`-file, you may have to
  remove this or add your Go bin-path as an entry (otherwise `sudo` won't find
  the executable)

## Implementation details

* Internationalized domain names are converted and stored as [IDNA
  Punycode](https://en.wikipedia.org/wiki/Punycode) in the hosts-file (for
  compatibility)
* On macOS we automatically create loopback aliases for IPs in the
  "127.0.0.0/8"-block, as they are not routed to the local machine by default
