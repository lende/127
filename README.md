# 127

_127_ is a tool for mapping easy-to-remember hostnames to random
[loopback addresses].

The tool works as a front-end to the standard [hosts file] on your system. It
has been tested on Linux, macOS (but see the [troubleshooting section]) and
Windows.

## Installation

```console
go install github.com/lende/127
```

You may also [download a binary release].

## Usage and options

```console
$ 127 -h
127 is a tool for mapping hostnames to random loopback addresses.

Usage: 127 [option ...] [hostname]
Print IP mapped to hostname, assigning a random IP if no mapping exists.

Options:
  -d    delete hostname mapping
  -f string
        path to hosts file (default "/etc/hosts")
  -n    do not output a trailing newline
  -v    print version information
```

### Notes

- Internationalized domain names are converted and stored as [IDNA Punycode] in
  the hosts-file (for compatibility)
- There is also a [Go API] provided

## Examples

### A simple demonstration

```console
# Map example.test to a random loopback address:
$ sudo 127 example.test
127.2.221.30

# Running the command again simply returns the same IP address:
$ 127 example.test
127.2.221.30

# Ping the new host to check that it worked:
$ ping example.test
PING example.test (127.2.221.30) 56(84) bytes of data.
64 bytes from example.test (127.2.221.30): icmp_seq=1 ttl=64 time=0.042 ms

# Delete the mapping by specifying the -d flag:
$ 127 -d example.test
127.2.221.30
$ ping example.test
ping: example.test: Name or service not known

# Running the command without any arguments simply returns a random IP:
$ 127
127.167.166.218
```

### Testing a third party service

Let's say you want to try out [ownCloud]. Simply run:

```console
sudo docker run --rm -p `sudo 127 owncloud.test`:80:80 owncloud/server
```

... and your _ownCloud_ instance should be available at `http://owncloud.test`.

## Troubleshooting

- If there is a `secure_path`-entry in your `/etc/sudoers`-file, you may have to
  remove this or add your Go bin-path as an entry (otherwise `sudo` won't find
  the executable)
  - Alternatively you could set the owner of `/etc/hosts` to a group you are
    member of (for instance the `sudo`-group), and give it write-access
  - A third alternative is to copy the `127`-executable to a global path (such
    as `/usr/local/bin`)
- On macOS loopback addresses are not routed to the local host by default, and
  aliases must be explicitly created for each IP address before binding to it
  - See this [Super User question] for details
  - Here is an [idea for a daemon] that would solve the issue

[loopback addresses]: https://en.wikipedia.org/wiki/Localhost#Name_resolution
[hosts file]: https://en.wikipedia.org/wiki/Hosts_(file)
[troubleshooting section]: #troubleshooting
[download a binary release]: https://github.com/lende/127/releases
[IDNA Punycode]: https://en.wikipedia.org/wiki/Punycode
[Go API]: https://godoc.org/github.com/lende/127/lib
[ownCloud]: https://owncloud.org/
[Super User question]: https://superuser.com/questions/458875/
[idea for a daemon]: https://github.com/lende/127d
