# 127

_127_ is a tool for mapping easy-to-remember hostnames to random
[loopback addresses].

The tool reads and modifies the standard [hosts file] on your system. There is
also a [Go API] available. Tested on Linux, but may work on other Unix-like
systems.

## Installation

To install the latest version from source:

```console
git clone https://github.com/lende/127.git
cd 127; make && sudo make install
```

You may also [download a binary release].

## Usage and options

```console
$ 127 -h
127 is a tool for mapping hostnames to random loopback addresses.

Usage: 127 [option ...] [hostname]
Print IP mapped to hostname, assigning a random IP if no mapping exists.

Options:
  -d    delete mapping
  -e    echo hostname
  -f string
        path to hosts file (default "/etc/hosts")
  -v    print version information
```

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
sudo docker run --rm -p `sudo 127 owncloud.test`:80:80 owncloud:latest
```

... and your _ownCloud_ instance should be available at `http://owncloud.test`.

[loopback addresses]: https://en.wikipedia.org/wiki/Localhost#Name_resolution
[hosts file]: https://en.wikipedia.org/wiki/Hosts_(file)
[download a binary release]: https://github.com/lende/127/releases
[Go API]: https://godoc.org/github.com/lende/127/lib127
[ownCloud]: https://owncloud.org/
