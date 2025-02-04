# golang-lru

This provides the `lru` package which implements a fixed-size
thread safe LRU cache. It is based on the Hashicorp LRU cache, which itself is based on Groupcache.

Full docs are available on [Godoc](http://godoc.org/github.com/optakt/golang-lru)

## Example

Using the LRU is very simple:

```go
l, _ := New(128)
for i := 0; i < 256; i++ {
    l.Add(i, nil)
}
if l.Len() != 128 {
    panic(fmt.Sprintf("bad len: %v", l.Len()))
}
```
