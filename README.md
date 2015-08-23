# haiku

古池や蛙飛び込む水の音

Haiku Matcher

## Usage

```go
if haiku.Match(arg, []int{5, 7, 5}) {
    log.Println(arg + " is haiku")
}
```

```go
for _, h := range haiku.Find(arg, []int{5, 7, 5}) {
    log.Println(h + "、575 じゃん")
}
```

## Installation

```
$ go get github.com/mattn/go-haiku/cmd/haiku
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a mattn)
