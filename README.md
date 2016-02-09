# shitpic [![GoDoc](https://godoc.org/github.com/carlmjohnson/shitpic?status.svg)](https://godoc.org/github.com/carlmjohnson/shitpic)
As explained in [The Awl](http://www.theawl.com/2014/12/the-triumphant-rise-of-the-shitpic), a **shitpic** happens, “when an image is put through some diabolical combination of uploading, screencapping, filtering, cropping, and reuploading. They are particularly popular on Instagram.”

`shitpic` is a utility for creating shitpics. It recompresses an input file a number of times (default 100) and saves the degraded output.

##Example
Input:

![Clean monkey](http://i.imgur.com/ULOm0le.png)

Output:

![Dirty monkey](http://i.imgur.com/pdgFU2d.jpg)

## Installation
First install [Go](http://golang.org) and set your `GOPATH` environmental variable to the directory you would like the project saved in. Then run `go get github.com/carlmjohnson/shitpic`. The binary will be installed in `$GOPATH/bin`. If you don't want to keep the source, you can instead run `GOPATH=/tmp/sp go get github.com/carlmjohnson/shitpic && cp /tmp/sp/bin/shitpic .` to install the binary to your current working directory.

## Usage
```bash
$ shitpic -h
Usage of shitpic:
	shitpic [options] input output
  -cycles uint
    	How many times to reprocess input (default 100)
  -quality int
    	Lower bound of quality (0–100) (default 75)
```
