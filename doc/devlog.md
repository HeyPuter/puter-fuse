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

## 2024-02-19

### Mapping inode numbers

Considering this scenario:
- Create a file
  - inode number is assigned to a write cache entry
- Create operation is completed
  - server responds with a UUID for the file
- Associate inode number with UUID

In the current implementation of `puterfs.Filesystem` in this repo,
there is no write cache entry and inode numbers have a 1-to-1 relationship
with UUIDs; this will not be the case anymore.
As such, a few questions need to be considered:
- Is `puterfs.Filesystem` or a layer of the `FAO` interface responsible
  for associating UUIDs with inode numbers?
- Use temporary UUIDs or something else for write cache entries?

Since UUIDs are stored as strings this allows for the UUID of a write
cache entry to be set to something like `write-cache:<path of file>`.
However, I forsee a rare and subtle bug caused from this. Imagine a
file is created at `/a/my_file.txt`, and the file is then moved to
`/b/my_file.txt` before the real underlying `create` operation has
been completed. This is fine if we remember that the write-cache UUID
for the file is still `write-cache:/a/my_file.txt`, but what do we set
as the write cache UUID of a new file created at `/a/my_file.txt` before
that same previous `create` operation is completed? One way or another
we lose information about one of these files and don't associate its
inode correctly.

Since UUIDs are universally unique, we can create a temporary UUID for
the write cache entry and record a table that maps write-cache UUIDs to
remote UUIDs. I'd say take this one step further and assign every entry
a local UUID (including entries from remote); this reducecs branching
conditions that would need to check which UUID to use, since for certian
purposes (like assigning the inode number) we always use the local UUID
regardless if a file has a write-cache entry or an entry that's synced with
the remote filesystem.

Next question is: can we store all the UUID associations and inode associations
in memory? The former we didn't have before, and the latter is currently stored
in memory. I have on my own computer approximately 2,000,000 files in my home
directory and 20,000,000 million (user-accessible) files from `/`.
Taking the worst case scenario (20M files, string representation of UUID),
[a chat with GPT](https://chat.openai.com/share/1d7adfb1-d3a7-424f-be8d-bface27c18ed)
followed by some manual calculations suggests this scenario would cost 2GB
of memory.

I think for now, this case is incredibly unlikely and instead the user will have
at most a few large projects in the filesystem. For us, all source files across
all Puter repos (at least, ones that I have cloned) come to `150658` files,
or 14 MB of memory required to store UUID associations. So yes, for the
foreseeable future it is absolutely acceptable to store all this information
in memory.

Note the calculation for UUID-to-inode association is omitted since it's
guarenteed to be smaller.

### Let's take a look at NodeInfo (stat object) storage

Considering the above calculations it should be pretty simple
to consider the storage of `NodeInfo` (formerly `CloudItem`)
in memory as well.

Again I [delegated to GPT](https://chat.openai.com/share/0bc8644c-2209-4fb5-b20e-0fd0685d1b66)
and the analysis it ran suggests `1417` bytes, which is within the range
of my expectations (considering we're storing a decent amount of information
in that object as compared to a POSIX stat object).

This comes to 203 MB of storage using the same number of files used in the
previous analysis for storing UUID associations. (150658)
This is acceptable considering the performance advantage.

### Verbosity in Golang

While this isn't a deal breaker, I'm finding I have to repeat myself
a lot while writing code in Go.

The `gen.js` core generator in this repository, which creates `FAO`, `ProxyFAO`,
and `BaseFAO` (eventually `LoggingFAO` as well) from a meta interface
definition has _already_ saved me time as I've continually tweaked the
interface to account for oversights.
A proxy implementation, and implementations that would extend proxies
like loggers and profilers, are such behaviours that can be implied from
a class definition. It's possible to achieve this in javascript, but in
Golang it seems only to be possible by creating an abstraction that
provides prototypal inheritence, or a solution involving code generation
(perhaps combining reflection to avoid the need for an external
representation of the class definition)

Now that I'm adding `AssociationService` and `NodeInfoStore`, I'm about
to place more `RWMutex` and related logic around each of the maps in
these structs. This is really a job for the decorator pattern; i.e.
an implementor of an `IMap` interface that applies mutexes and delegates
to an underlying implementor of `IMap`.
Since golang doesn't support operator overloading, the existing interface
with map objects is incompatible with an implementation that allows for
decorators.

Note: after writing the above I decided to implement `IMap` using
generics, which were introduced in Go 1.18.

### Data Structures

#### Directory Tree

- map from Local UID to entries containing:
  - pointer to list of child local UIDs
  - lock for mutating the list of child local UIDs
  - timestamp of last readdir

#### File cache

- map of SHA1 hashes to entries containing:
  - pointer to cached data
  - timestamp of read
  - list of references (keeps data in cache)
  - lock for mutating list of references
  - priority queue for TTL of cached entries?
    - for now the TTLCacheFAO will add a reference on
      the entry and call `<-time.After` before removing it.

#### Write back cache

- map from Local UID to "branches"; a branch contains:
  - SHA1 hash of a cache entry, or an empty string if the file
    is new (doesn't exist on remote yet)
  - a list of mutations to apply to that cache entry
    - a mutation can be:
      - write mutation (data []byte, offset int64)
      - truncate mutation
      - delete mutation

When a delete mutation is applied, it would be acceptable to
remove the cached data and other mutations in the write-back cache
