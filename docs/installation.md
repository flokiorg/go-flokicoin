# Installation

The first step is to install flokicoind.  See one of the following sections for
details on how to install on the supported operating systems.

## Requirements

[Go](http://golang.org) 1.17 or newer.

## GPG Verification Key

All official release tags are signed by Conformal so users can ensure the code
has not been tampered with and is coming from The btcsuite developers.  To
verify the signature perform the following:

* Download the Conformal public key:
  https://raw.githubusercontent.com/btcsuite/flokicoind/master/release/GIT-GPG-KEY-conformal.txt

* Import the public key into your GPG keyring:

  ```bash
  gpg --import GIT-GPG-KEY-conformal.txt
  ```

* Verify the release tag with the following command where `TAG_NAME` is a
  placeholder for the specific tag:

  ```bash
  git tag -v TAG_NAME
  ```

## Windows Installation

* Install the MSI available at: [flokicoind windows installer](https://github.com/flokiorg/go-flokicoin/releases)
* Launch flokicoind from the Start Menu

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

* Run the following commands to obtain flokicoind, all dependencies, and install it:

```bash
git clone https://github.com/flokiorg/go-flokicoin $GOPATH/src/github.com/flokiorg/go-flokicoin
cd $GOPATH/src/github.com/flokiorg/go-flokicoin
go install -v . ./cmd/...
```

* flokicoind (and utilities) will now be installed in ```$GOPATH/bin```.  If you did
  not already add the bin directory to your system path during Go installation,
  we recommend you do so now.

## Gentoo Linux Installation

* [Install Layman](https://gitlab.com/bitcoin/gentoo) and enable the Bitcoin overlay.
* Copy or symlink `/var/lib/layman/bitcoin/Documentation/package.keywords/flokicoind-live` to `/etc/portage/package.keywords/`
* Install flokicoind: `$ emerge net-p2p/flokicoind`

## Startup

Typically flokicoind will run and start downloading the block chain with no extra
configuration necessary, however, there is an optional method to use a
`bootstrap.dat` file that may speed up the initial block chain download process.

* [Using bootstrap.dat](https://github.com/flokiorg/go-flokicoin/blob/master/docs/configuration.md#using-bootstrapdat)
