# go-bt
A small BitTorrent client using solely the standard library

# Usage

## Parsing bencoded files

```
❯ go run ./main.go bencode -decode=/tmp/files.torrent | cut -c 1-80                    main
{
  "announce": "http://localhost:8088",
  "created by": "go-bt",
  "info": {
    "files": [
      {
        "length": 7000000,
        "path": [
          "/tmp/files/file1"
        ]
      },
      {
        "length": 2000000,
        "path": [
          "/tmp/files/file2"
        ]
      },
      {
        "length": 3000000,
        "path": [
          "/tmp/files/file3"
        ]
      }
    ],
    "name": "foo",
    "piece length": 65536,
    "pieces": ")O\ufffdx9W\ufffdA\ufffd6A\ufffd\ufffd/\ufffd\\\ufffd\ufffdȚ\ufff
  }
}
```

## Creating a `.torrent` file

```
❯ go run ./main.go create -announce http://localhost:8088 -name foo -pieceLength 65536 -out /tmp/files.torrent /tmp/files/file* 
❯ head -c 120 /tmp/files.torrent
d8:announce21:http://localhost:808810:created by5:go-bt4:infod5:filesld6:lengthi7000000e4:pathl16:/tmp/files/file1eed6:l%   
```

## Getting a URL-encoded `info_hash`

```
❯ go run ./main.go infohash -file=/tmp/files.torrent
hex: b6e355aa9e2a9b510cf67f0b4be76d9da36ddbbf
url: %B6%E3U%AA%9E%2A%9BQ%0C%F6%7F%0BK%E7m%9D%A3m%DB%BF
```

## Starting your own tracker

```
❯ go run ./main.go tracker -dir=/tmp serve -port=8080
2024/10/09 17:43:13 serving torrents from /tmp on :8080
2024/10/09 17:43:13 torrent file found: files.torrent
```

## Querying a tracker

```
❯ go run ./main.go tracker -torrent=/Users/axiomiety/Downloads/ubuntu-24.10-desktop-amd64.iso.torrent
2024/10/19 21:27:07 querying tracker: https://torrent.ubuntu.com/announce?info_hash=%3F%9A%AC%15%8C%7D%E8%DF%CA%B1q%EAX%A1z%AB%DF%7F%BC%93&peer_id=%02%F9%5DB%B14%AEJ%A9A%89%A1%15%E7%E2%3D%E7%8D3q&port=6688&uploaded=0&downloaded=0&left=45536&numwant=100
{
  "complete": 626,
  "incomplete": 58,
  "interval": 1800,
  "peers": [
    {
      "ip": "2607:5300:60:8460::1",
      "peer id": "-lt0D80-\ufffd\ufffd\ufffd\u0016'MѓY\ufffd\u0011;",
      "port": 20757
    },
  ...
```