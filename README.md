# shitpic [![GoDoc](https://godoc.org/github.com/carlmjohnson/shitpic?status.svg)](https://godoc.org/github.com/carlmjohnson/shitpic)
As explained in [The Awl](http://www.theawl.com/2014/12/the-triumphant-rise-of-the-shitpic), a **shitpic** happens, “when an image is put through some diabolical combination of uploading, screencapping, filtering, cropping, and reuploading. They are particularly popular on Instagram.”

`shitpic` is a utility for creating shitpics. It recompresses an input file a number of times (default 100) and saves the degraded output.

## Example
Input:

![Clean monkey](http://i.imgur.com/ULOm0le.png)

Output:

![Dirty monkey](http://i.imgur.com/pdgFU2d.jpg)

## Installation
First install [Go](http://golang.org).

If you just want to install the binary to your current directory, and don't care about the source code, run

```bash
GOBIN=$(pwd) GOPATH=/tmp/gobuild go get github.com/carlmjohnson/shitpic
```

## Usage
```bash
$ shitpic -h
Usage of shitpic:
    shitpic [options] input output

Shitpic accepts and can output JPEG, GIF, and PNG files.

  -cycles uint
        How many times to reprocess input (default 100)
  -quality int
        Lower bound of quality (0–100) (default 75)
  -reduce-colors
        Reduce to 256 colors
```
