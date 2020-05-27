// A bit array (also known as bit map, bit set, bit string, or bit vector)
// is an array data structure that compactly stores bits.
// https://en.wikipedia.org/wiki/Bit_array

package bitarray

import (
	"sync"
	"unsafe"

	"github.com/aermolaev/atomicvalue"
)

// BitArray array of binary values.
type BitArray struct {
	mu       sync.RWMutex
	blocks   []BitBlock
	curIndex int64
	size     int64
	capacity int64
	count    atomicvalue.Int
}

type BitBlock uint64

const (
	blockSize    = int64(unsafe.Sizeof(BitBlock(0)) * 8)
	bitBlockFull = (1 << blockSize) - 1

	bitBlockMark   = true
	bitBlockUnmark = false

	BitBlockNotFound = -1
)

// NewBitArray creates and initializes a new BitArray using capacity as its
// initial capacity.
func NewBitArray(capacity int64) *BitArray {
	size := (capacity / blockSize) + 1

	return &BitArray{
		blocks:   make([]BitBlock, size),
		capacity: capacity,
		size:     size,
	}
}

// HasRoom reports true if this BitArray contains bits that are set to true.
func (b *BitArray) HasRoom() bool {
	return b.count.Get64() < b.capacity
}

// IsEmpty reports true if this BitArray contains no bits that are set to true.
func (b *BitArray) IsEmpty() bool {
	return !b.HasRoom()
}

// Len returns the number of occupied bits.
func (b *BitArray) Len() int {
	return b.count.Get()
}

// Cap returns the BitArray capacity, that is, the total bits allocated
// for the data.
func (b *BitArray) Cap() int {
	return int(b.capacity)
}

// Reset resets BitArray to initial state.
func (b *BitArray) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := int64(0); i < b.size; i++ {
		b.blocks[i] = BitBlock(0)
	}

	b.count.Set(0)
}

// Set sets the bit at the specified index to the specified value.
func (b *BitArray) Set(index int64, mark bool) (changed bool) {
	if i, j := bitIndexAndNum(index); i < b.size {
		block := &b.blocks[i]

		b.mu.Lock()

		if mark == bitBlockMark {
			if changed = block.compareAndMark(j); changed {
				b.count.Inc()
			}
		} else {
			if changed = block.compareAndUnmark(j); changed {
				b.count.Dec()

				if i < b.curIndex {
					b.curIndex = i // move pointer closer to the beginning
				}
			}
		}

		b.mu.Unlock()
	}

	return
}

// Get returns the value of the bit with the specified index.
func (b *BitArray) Get(index int64) (res bool) {
	if i, j := bitIndexAndNum(index); i < b.size {
		block := &b.blocks[i]

		b.mu.RLock()
		res = block.value(j)
		b.mu.RUnlock()
	}

	return
}

// Mark sets the bit at the specified index to true.
func (b *BitArray) Mark(index int64) {
	b.Set(index, bitBlockMark)
}

// Unmark sets the bit at the specified index to false.
func (b *BitArray) Unmark(index int64) {
	b.Set(index, bitBlockUnmark)
}

// MarkFree finds the index of the first bit that is set to false and
// sets the bit to true. Returns index of changed bit. Returns BitBlockNotFound
// unless array has room.
func (b *BitArray) MarkFree() (index int64) {
	index = BitBlockNotFound

	if !b.HasRoom() { // fast check w/o lock
		return
	}

	b.mu.Lock()

	if b.HasRoom() {
		if block := b.nextFree(); block != nil {
			b.count.Inc()

			j := block.ffz()
			block.mark(j)

			index = (b.curIndex * blockSize) + j
		}
	}

	b.mu.Unlock()

	return
}

func (b *BitArray) nextFree() *BitBlock {
	for i := int64(0); i < b.size; i++ {
		if block := b.current(); block.hasRoom() {
			return block
		}

		b.curIndex = (b.curIndex + 1) % b.size
	}

	return nil
}

func (b *BitArray) current() *BitBlock {
	return &b.blocks[b.curIndex]
}

func bitIndexAndNum(i int64) (int64, int64) {
	return i / blockSize, i % blockSize
}

func (b BitBlock) value(bit int64) bool {
	return (b & mask(bit)) != 0
}

func (b *BitBlock) mark(bit int64) {
	*b |= mask(bit)
}

func (b *BitBlock) unmark(bit int64) {
	*b &^= mask(bit)
}

func (b *BitBlock) compareAndMark(bit int64) (changed bool) {
	n := mask(bit)
	changed = (*b & n) == 0

	if changed {
		*b |= n
	}

	return
}

func (b *BitBlock) compareAndUnmark(bit int64) (changed bool) {
	n := mask(bit)
	changed = (*b & n) != 0

	if changed {
		*b &^= n
	}

	return
}

func (b BitBlock) hasRoom() bool {
	return b != bitBlockFull
}

func (b BitBlock) ffz() int64 {
	v := (b & (^b - 1))

	switch blockSize {
	case 64:
		return popcount64(uint64(v))

	case 32:
		return popcount32(uint32(v))

	default:
		panic("wrong block size")
	}
}

func popcount64(b uint64) int64 {
	const (
		m1 = 0x5555555555555555 // binary: 0101...
		m2 = 0x3333333333333333 // binary: 00110011..
		m4 = 0x0f0f0f0f0f0f0f0f // binary: 4 zeros, 4 ones ...
	)

	b -= (b >> 1) & m1             // put count of each 2 bits into those 2 bits
	b = (b & m2) + ((b >> 2) & m2) // put count of each 4 bits into those 4 bits
	b = (b + (b >> 4)) & m4        // put count of each 8 bits into those 8 bits
	b += b >> 8                    // put count of each 16 bits into their lowest 8 bits
	b += b >> 16                   // put count of each 32 bits into their lowest 8 bits
	b += b >> 32                   // put count of each 64 bits into their lowest 8 bits
	return int64(b & 0x7f)
}

func popcount32(b uint32) int64 {
	const (
		m1 = 0x55555555 // binary: 0101...
		m2 = 0x33333333 // binary: 00110011..
		m4 = 0x0f0f0f0f // binary: 4 zeros, 4 ones ...
	)

	b -= (b >> 1) & m1
	b = (b & m2) + ((b >> 2) & m2)
	b = (b + (b >> 4)) & m4
	b = b + (b >> 8)
	b = b + (b >> 16)
	return int64(b & 0x3f)
}

func mask(bit int64) BitBlock {
	return 1 << bit
}
