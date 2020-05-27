package bitarray

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitArray(t *testing.T) {
	assert := assert.New(t)

	b := NewBitArray(1_000_000)

	b.Mark(4000)
	assert.True(b.Get(4000))

	b.Unmark(4000)
	assert.False(b.Get(4000))

	// other
	assert.False(b.Get(4001))
}

func TestBitArrayHasRoom(t *testing.T) {
	assert := assert.New(t)

	const count = 1_0
	b := NewBitArray(count)

	for i := 0; i < count-1; i++ {
		b.MarkFree()
		assert.True(b.HasRoom())
	}

	b.MarkFree()
	assert.False(b.HasRoom())
	assert.True(b.IsEmpty())

	b.Set(0, false)
	assert.True(b.HasRoom())

	b.MarkFree()
	assert.False(b.HasRoom())
}

func TestBitArrayReset(t *testing.T) {
	assert := assert.New(t)

	const count = 100
	b := NewBitArray(count)

	for i := 0; i < count; i++ {
		b.MarkFree()
	}

	assert.False(b.HasRoom())
	b.Reset()
	assert.True(b.HasRoom())
	assert.Zero(b.count.Get())
}

func TestBitArrayConcurent(t *testing.T) {
	assert := assert.New(t)

	const count = 5_000
	b := NewBitArray(count)

	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			i := b.MarkFree()
			go func() {
				defer wg.Done()
				assert.True(b.Get(i))
				b.Unmark(i)
			}()
		}()
	}
	wg.Wait()

	assert.Zero(b.Len())
}

func TestBitArrayMarkGet(t *testing.T) {
	assert := assert.New(t)

	const count = 100_000
	b := NewBitArray(count)

	// check for all blocks are sets to false
	for i := int64(0); i < count; i++ {
		assert.False(b.Get(i))
	}

	// mark
	for i := int64(0); i < count; i++ {
		b.Set(i, true)
	}

	for i := int64(0); i < count; i++ {
		assert.True(b.Get(i))
	}

	// unmark
	for i := int64(0); i < count; i++ {
		b.Set(i, false)
	}

	for i := int64(0); i < count; i++ {
		assert.False(b.Get(i))
	}
}

func TestBitArrayMarkFree(t *testing.T) {
	assert := assert.New(t)

	b := NewBitArray(1_000_000)
	for i := 0; i < 100; i++ {
		assert.Equal(int64(i), b.MarkFree())
	}
}

func BenchmarkBitIndexAndNum(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = bitIndexAndNum(int64(n))
	}
}

func BenchmarkBitArrayGet(b *testing.B) {
	const size = 10_000_000
	ba := NewBitArray(size)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = ba.Get(10000)
		ba.Set(size-1, true)
	}
}

func BenchmarkBitArrayMarkFree(b *testing.B) {
	const size = 100_000_000
	ba := NewBitArray(size)

	for i := 0; i < int(ba.size/2); i++ {
		ba.MarkFree()
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ba.MarkFree()
	}
}

func BenchmarkBlockType(b *testing.B) {
	bc := BitBlock(10)

	for n := 0; n < b.N; n++ {
		_ = bc.value(10)
	}
}

func BenchmarkCompareAndMark(b *testing.B) {
	bc := BitBlock(10)
	bi := int64(10)

	for n := 0; n < b.N; n++ {
		_ = bc.compareAndMark(bi)
	}
}
