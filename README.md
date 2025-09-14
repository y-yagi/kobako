kobako
======

kobako is a small CLI wrapper that runs commands inside Docker containers with your current directory mounted. It's useful to run language toolchains (Go, Node, Python, etc.) without installing them locally.

Usage
-----

Basic usage (mounts current directory into container by default):

```shell
kobako <command> [args...]
```

Examples
--------

Run `go version` in a Go container:

```shell
kobako go version
```

Run `npx create-react-app` (uses `node:alpine` by default):

```shell
kobako npx create-react-app my-app
```

Run Python tests:

```shell
kobako python -m pytest
```

Environment variables
---------------------

- `KOBAKO_IMAGE`: override the container image used.
- `KOBAKO_NODE_IMAGE`: override the Node image used when command selection defaults to Node.
- `KOBAKO_HOST_DIR`: override the host directory mounted into the container (defaults to current directory).
- `KOBAKO_WORKDIR`: override the target workdir inside the container (defaults to `/work`).

Notes
-----

- By default the tool selects Alpine-based images for lightweight containers.
