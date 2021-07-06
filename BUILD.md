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
```
mkdir -p $HOME/go/bin
export GOPATH=$HOME/go
```

Install required libraries:
```
sudo apt-get update && sudo apt-get install -y gcc libgtk-3-dev libappindicator3-dev
```

Build steps:
```
git clone https://github.com/alttpo/sni.git

cd sni

go install ./cmd/sni
```

## Results
You should now have a single-file executable `sni` in `$GOPATH/bin`.

You can now copy this `sni` binary to your `/usr/local/bin` folder
or you can `export PATH=$PATH:$GOPATH/bin`.

