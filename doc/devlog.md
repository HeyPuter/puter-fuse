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
