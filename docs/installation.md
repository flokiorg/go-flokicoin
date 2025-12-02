# Installation

The first step is to install lokid.  See one of the following sections for
details on how to install on the supported operating systems.

## Requirements

[Go](http://golang.org) 1.17 or newer.


## Windows Installation

Currently, there is no MSI installer available for flokicoin. The `lokid` executable can be directly downloaded from the latest release page:

[lokid Releases](https://github.com/flokiorg/go-flokicoin/releases)

### Instructions:

1. **Download** the executable (`lokid.exe`) from the provided release link.
2. **Place** the executable in your preferred directory.
3. **Launch** `lokid.exe` by double-clicking the file or executing it via Command Prompt.

> **Note:** An official MSI installer is planned and will be available in an upcoming release.

## Linux/BSD/MacOSX/POSIX Installation

* Install Go according to the [installation instructions](http://golang.org/doc/install)
* Ensure Go was installed properly and is a supported version:

```bash
go version
go env GOROOT GOPATH
```

NOTE: The `GOROOT` and `GOPATH` above must not be the same path.  It is
recommended that `GOPATH` is set to a directory in your home directory such as
`~/goprojects` to avoid write permission issues.  It is also recommended to add
`$GOPATH/bin` to your `PATH` at this point.

* Run the following commands to obtain lokid, all dependencies, and install it:

```bash
git clone https://github.com/flokiorg/go-flokicoin $GOPATH/src/github.com/flokiorg/go-flokicoin
cd $GOPATH/src/github.com/flokiorg/go-flokicoin
go install -v . ./cmd/...
```

* lokid (and utilities) will now be installed in ```$GOPATH/bin```.  If you did
  not already add the bin directory to your system path during Go installation,
  we recommend you do so now.


## Startup

Typically lokid will run and start downloading the block chain with no extra
configuration necessary, however, there is an optional method to use a
`bootstrap.dat` file that may speed up the initial block chain download process.

* [Using bootstrap.dat](https://github.com/flokiorg/go-flokicoin/blob/master/docs/configuration.md#using-bootstrapdat)
