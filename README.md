<h3 align="center"><img width="300" alt="HiTIDE logo" src="./doc/logo.png"></h3>
<h3 align="center">FUSE Driver for the Puter Internet OS</h3>
<h4 align="center">Access Puter's filesystem on your device</h4>
<hr>

## What does this do?

This program lets you mount your files from the [Puter Open-Source Internet OS](https://github.com/HeyPuter/puter) as though
they were another harddrive. This works on Linux and Mac OS
using the FUSE interface.

## How to use

[Make sure Go is installed.](https://go.dev/doc/install)
This has been tested on version `go1.22.0`.

Then, run this command to install `puter-fuse`:

```sh
go install github.com/HeyPuter/puter-fuse@v1.0.0
```

## Configuration

### First-time Configuration

The first-time configuration will ask you for your Puter username
and password. If you don't have an account on puter.com you'll need
one in order to use this FUSE driver. Note that once we release the
open-source Puter Kernel you'll be able to login to any instance of
that instead.

Entering your username and password, and accepting the default options
for all other questions, should be sufficient for most installations.

### Configuration file

Configuration is saved to:

- `$HOME/.config/puterfuse/config.json`

## Technical Information

### What's a FUSE?

Filesystem in USErspace (FUSE) is an interface for filesystem
drivers that are loaded as userspace programs rather than in
the kernel. It is available in most POSIX systems like Linux
and Mac OS.

Puter's FUSE driver allows access to a cloud filesystem in a
way that looks like an ordinary local filesystem.

### Performance and Caching

Currently directory trees support read and write-back caching.
The contents of files are not currently cached by default, but
you can set `experimental_cache` to `true` in the configuration
file to enable read and write-back caching for files.
