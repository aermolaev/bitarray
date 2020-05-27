# BitArray

A bit array (also known as bit map, bit set, bit string, or bit vector)
is an array data structure that compactly stores bits.

## Installation

    go get github.com/aermolaev/bitarray

## Simple Example

```go
b := NewBitArray(1_000_000)

b.Mark(4000)
b.Get(4000) // true

b.Unmark(4000)
b.Get(4000) // false
```

```go
b := NewBitArray(1_000_000)

i := b.MarkFree() // 0
b.Get(i) // true
b.IsEmpty() // false
b.HasRoom() // true
```
