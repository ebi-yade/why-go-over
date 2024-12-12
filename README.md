# why-go-over

Identifies dependent modules that have raised the Go version.

## Installation

You can install `why-go-over` using Homebrew, from source, or by downloading the binary.

In case you're using Homebrew:

```shell
brew install ebi-yade/tap/why-go-over
```

<details>

<summary>Other ways</summary>

### From Source

```shell
go install github.com/ebi-yade/why-go-over/cmd/why-go-over@latest
```

### Downloading the Binary

You can download the binary from the [releases page](https://github.com/ebi-yade/why-go-over/releases/),

</details>

## Usage

Imagine you run `go get -u ./...` and notice that the `go` directive in your `go.mod` file has unexpectedly been updated to `1.23`.

To identify which modules caused the required Go version to increase, run:

```shell
why-go-over 1.23
```
