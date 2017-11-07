# 127

*127* is a tool for mapping easy-to-remember hostnames to random [loopback
addresses].

The tool works as a front-end to the standard [hosts-file] on your system. It has
been tested on Linux, macOS (but see the [troubleshooting section]) and Windows.

## Installation

```
go get github.com/lende/127
```

You may also [download a binary release].

## Usage and options

```console
$ 127 -h
127 is a tool for mapping hostnames to random loopback addresses.

Usage: 127 [option ...] [hostname]
Print IP mapped to hostname, assigning a random IP if no mapping exists.

Options:
  -block string
        address block (default "127.0.0.0/8")
  -d    delete hostname mapping
  -hosts string
        path to hosts file (default "/etc/hosts")
  -n    do not output a trailing newline
  -v    print version information
```

### Notes

* Internationalized domain names are converted and stored as [IDNA Punycode] in
  the hosts-file (for compatibility)
* There is also a [Go API] provided

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

Let's say you want to try out [ownCloud]. Simply run:

```
docker run -d -p `sudo 127 -n owncloud.test`:80:80 owncloud
```

... and your *ownCloud* instance should be available at `http://owncloud.test`.

## Troubleshooting

* If there is a `secure_path`-entry in your `/etc/sudoers`-file, you may have to
  remove this or add your Go bin-path as an entry (otherwise `sudo` won't find
  the executable)
    * Alternatively you could set the owner of `/etc/hosts` to a group you are
      member of (for instance the `sudo`-group), and give it write-access
    * A third alternative is to copy the `127`-executable to a global path (such
      as `/usr/local/bin`)
* On macOS loopback addresses are not routed to the local host by default, and
  aliases must be explicitly created for each IP address before binding to it
    * See this [Super User question] for details
    * Here is an [idea for a daemon] that would solve the issue

[loopback addresses]: https://en.wikipedia.org/wiki/Localhost#Name_resolution
[hosts-file]: https://en.wikipedia.org/wiki/Hosts_(file)
[troubleshooting section]: #troubleshooting
[download a binary release]: https://github.com/lende/127/releases
[IDNA Punycode]: https://en.wikipedia.org/wiki/Punycode
[Go API]: https://godoc.org/github.com/lende/127/lib
[ownCloud]: https://owncloud.org/
[Super User question]: https://superuser.com/questions/458875/
[idea for a daemon]: https://github.com/lende/127d
