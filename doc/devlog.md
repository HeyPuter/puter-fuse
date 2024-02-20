## 2023-05-17

### Choosing the language for the FUSE driver

For the Puter FUSE driver, it is important to choose a FUSE library
that will be easy to work with and is actively maintained; we need
to be able to resolve potential issues. Choosing the wrong library
could result in us needing to switch, write one from scratch, or
even port the driver to a different language.

I selected a subset of programming languages to find a FUSE driver
for. Here is a list of those languages as well as the rationale
for selecting the language as a candidate.

- `node` - we use javascript a lot already. writing a FUSE driver
  in a high-level language is probably okay, since the performance
  bottleneck will be with the network rather than the language the
  driver is written in.
- `python` - commonly used to interact with low-level interfaces
- `rust` - commonly used in the WASM ecosystem
- `golang` - simple language, convenient HTTP and websockets libraries

For each of these languages I took a look at the most promising
FUSE driver I could find and compared them.

[Node's FUSE driver](https://github.com/fuse-friends/fuse-native)
- last update in 2020
- stats: 26 forks, 19 open issues, 8 closed issues

[Python's FUSE driver](https://github.com/libfuse/python-fuse)
- last update April 25th
- stats: 70 forks, 13 open issues, 20 closed issues

[Rust's FUSE driver](https://github.com/zargony/fuse-rs)
- last update in 2020
- stats: 130 forks, 45 open issues, 37 closed issues

[Go's FUSE driver](https://github.com/hanwen/go-fuse)
- last update 3 weeks ago
- stats: 286 forks, 14 open issues, 247 closed issues

The FUSE driver for Golang seems the most promising. It is actively
maintained and is the most popular among the ones I've found.

## 2024-02-16

### Caching Cache

#### Write Cache

Implementing write caching in the FUSE driver will prevent error
in programs which don't expect a network filesystem, and will allow
batch operations on Puter's API without blocking users of the
filesystem.

```
FUSE Write --> WriteCacheFAO --> PuterFAO
```

Terms:
- **branch**: a structure including:
  - SHA1 hash of the contents of the file at some point in time
  - a chain of mutations performed on the file

Details:
- Every SHA1 hash that is referenced by a **branch** will have a
  corresponding entry in a map of structs, with each struct
  containing the contents of the file at the respective hash as
  well as a count of how many branches are referencing the hash.

WriteCacheFAO will implement:
- `Write` will initialize a new branch if none exists,
  add a `write` mutation to the branch,
  start a goroutine that delegates,
  then return to the caller.
  Upon return of the delegate, the branch is rebased
- `Truncate` will initialize a new branch if none exists,
  add a `truncate` mutation to the branch,
  and follow the same proceedure as with write
- `Read` will check for a write cache entry before delegating
- `Stat` will delegate, then check for a cached entry.
  If a cached entry exists, the response from the delegate
  will be mutated accordingly.
