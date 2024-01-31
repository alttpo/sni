## Requirements

Follow installation instructions for Go found at https://golang.org/doc/install

At least Go version 1.16 is required to build SNI due to use of new standard library
functionality introduced in 1.16.

## Windows

Set up `$GOPATH` environment variable to point to `$HOME/go`.
```
mkdir C:\go
mkdir C:\go\bin
export GOPATH=C:\go
```

Build steps:
```
git clone https://github.com/alttpo/sni.git

cd sni

go install ./cmd/sni
```

## MacOS

Set up `$GOPATH` environment variable to point to `$HOME/go`.
```
mkdir -p $HOME/go/bin
export GOPATH=$HOME/go
```

Build steps:
```
git clone https://github.com/alttpo/sni.git

cd sni

go install ./cmd/sni
```

## Linux
Set up `$GOPATH` environment variable to point to `$HOME/go`.
```sh
mkdir -p $HOME/go/bin
export GOPATH=$HOME/go
```

Install required libraries:
```sh
sudo apt-get update && sudo apt-get install -y gcc libgtk-3-dev libappindicator3-dev
# or
sudo apt-get update && sudo apt-get install -y gcc libgtk-3-dev libayatana-appindicator3-dev
```

Build steps:
```sh
git clone https://github.com/alttpo/sni.git

cd sni

go install -tags=legacy_appindicator ./cmd/sni # to use libappindicator3
# or
go install ./cmd/sni # to use libayatana-appindicator3
```

## Results
You should now have a single-file executable `sni` in `$GOPATH/bin`.

You can now copy this `sni` binary to your `/usr/local/bin` folder
or you can `export PATH=$PATH:$GOPATH/bin`.

