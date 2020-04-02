mp# CS425, Distributed Systems: Fall 2018

Machine Programming 4 â€“ Crane

# Introduction

## Dependency

1. [Golang](https://golang.org/doc/install)

## Location of stored files

```
/home/mp3/files/*.*
```

# Usage

Default listening port is 50050.

### Launch

```shell
cd ./node
go run *.go
```

### Command

## debug

use to debug program

```
debug
```

## join

use to join the group

```
join
```

## leave

use to leave the group

```
leave
```

## introducer

use to claim this node as introducer

```
introducer
```

## master

use to claim this node as introducer

```
master
```

## show

use to output the membership list of this node

```
show
```

## show-f

use to output the file stored infomation of this node

```
show-f
```

## show-m

use to output the global machine running infomation

```
show-m
```

## show-v

use to output the version information of file stored in this machine

```
show-v
```

## put

use to output the global machine running infomation

```
put localfilename sdfsfilename
```

## get

```
get sdfsfilename localfilename
```

## get-versions

gets all the last num-versions versions of the file into the localfilename (use delimiters to mark out versions).

```
get-versions filename versionnumber
```

## delete

delete file in this SDFS

```
delete sdfsfilename.
```

## ls

list all machine (VM) addresses where this file is currently being stored;

```
ls sdfsfilename
```

## store

At any machine, list all files currently being stored at this machine.

```
store ip_address
```

## master start

start push tuples

```
start
```

## client start

start receive tuples

```
startreceive
```

## client stop

stop receive tuples and write result into SDFS

```
stopreceive
```

## Grep Log

Default listening port is 50055.

### Launch Server

```shell
cd ./query/server
go run server.go
```

### Open Client

```shell
cd ./query/client
go run client.go
```

### Search Command

After the client start, you can grep anything you want under the guidance of the prompt

1. normal grep

```shell
# search for the special word "aaaaaa"
Enter the string to grep: aaaaaa
```

2. arbitrary regexps grep

```shell
# search for the arbitrary regexps
# use the option filed "-P"
Enter the string to grep: -P '\d{10,}'
```

3. regular expressions

```shell
# search for the regular expressions '\d{5,}[a-z]{4}'
# use the option filed "-P"
Enter the string to grep: -P '\d{5,}[a-z]{4}'
```
