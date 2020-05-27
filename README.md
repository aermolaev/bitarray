# BitArray

A bit array (also known as bit map, bit set, bit string, or bit vector)
is an array data structure that compactly stores bits.

## Installation

    go get github.com/aermolaev/bitarray

## Simple Example

```go
package main

import (
	"fmt"

	"github.com/aermolaev/bitarray"
)

func main() {
	b := bitarray.NewBitArray(1_000_000)

	b.Mark(4000)
	fmt.Println(b.Get(4000)) // true

	b.Unmark(4000)
	fmt.Println(b.Get(4000)) // false

	i := b.MarkFree()        // 0
	fmt.Println(i)           // true
	fmt.Println(b.Get(i))    // true
	fmt.Println(b.IsEmpty()) // false
	fmt.Println(b.HasRoom()) // true
}
```
