# haiku

古池や蛙飛び込む水の音

Haiku Matcher

## Interface

```go
if haiku.Match(text, []int{5, 7, 5}) {
    log.Println(arg + " is haiku")
}
```

```go
for _, h := range haiku.Find(text, []int{5, 7, 5}) {
    log.Println(h + "、575 じゃん")
}
```

## Usage

From argument

```
$ haiku あぁ柿くへば鐘が鳴るなり法隆寺
柿くへば 鐘が鳴るなり 法隆寺
```

From stdin

```
$ cat README.md | haiku
古池や 蛙飛び込む 水の音
```

Extract 俳句 from URL with `-u` option

```
$ haiku -u "https://ja.wikipedia.org/wiki/ハノイの塔"
中央に 穴の開いた 大きさの
円盤の 上に大きな 円盤を
有効な 問題として 有名で
円盤を 移動させると すると次
一回り 大きい物の 右隣
一回り 大きい物の 右隣
円盤を 対応付けた とき数字
円盤を 動かすことで 解答が
円盤を 別の柱に 移し替え
```

Extract 短歌 with `-r` option

```
$ haiku -r 5,7,5,7,7 -u "https://ja.wikipedia.org/wiki/フクロウ"
フクロウが 鳴くと明日は 晴れるので 洗濯物を 干せという意味
```

## Installation

```
$ go get github.com/mattn/go-haiku/cmd/haiku
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a mattn)
