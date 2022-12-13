# Metachart

`metachart` is a tool to generate [Helm](https://helm.sh/) Charts using
Kubernetes Resources [JSON Schema](https://json-schema.org/), user provided
preprocessing logic and some syntactic sugar.

Motivation: Use Helm without writing own Charts.

What `metachart` born charts can provide:
- Use complete Kubernetes resources API out of the box
- Generate Chart for any Kubernetes resource type using JSON Schema
- All string params support gotpl rendering
- Own resources preprocessors and custom schema definitions adding some
  resources build logic
- Features: `defaults`, `related`, `checksums`

## Installation

Download the latest binary from the [Releases](/releases) page.

On macos the application can be installed using [Homebrew](https://brew.sh/):

```shell
brew tap iponweb/tap
brew update
brew install metachart
```

## Quickstart

Please follow the [Quickstart](docs/quickstart.md) guide.

## Requirements

For `metachart` born charts minimal supported [Helm](https://helm.sh/) version
is `v3.2.0`.

## Documentation

For complete documentation see the [docs](docs) directory.

## Kicked off by

![](assets/iponweb-logo.png)

## License

Apache License 2.0, see [LICENSE](LICENCE).
