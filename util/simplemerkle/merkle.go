package simplemerkle

import (
	"math"

	"github.com/pactus-project/pactus/crypto/hash"
)

var hasher func([]byte) hash.Hash

func init() {
	hasher = hash.CalcHash
}

type Tree struct {
	merkles []*hash.Hash
}

// nextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two.  This is a helper function used during the
// calculation of a merkle tree.
func nextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1

	return 1 << exponent // 2^exponent
}

// HashMerkleBranches takes two hashes, treated as the left and right tree
// nodes, and returns the hash of their concatenation.  This is a helper
// function used to aid in the generation of a merkle tree.
func HashMerkleBranches(left, right *hash.Hash) *hash.Hash {
	// Concatenate the left and right nodes.
	var h [hash.HashSize * 2]byte
	copy(h[:hash.HashSize], left.Bytes())
	copy(h[hash.HashSize:], right.Bytes())

	newHash := hasher(h[:])

	return &newHash
}

func NewTreeFromSlices(slices [][]byte) *Tree {
	hashes := make([]hash.Hash, len(slices))
	for i, b := range slices {
		hashes[i] = hasher(b)
	}

	return NewTreeFromHashes(hashes)
}

func NewTreeFromHashes(hashes []hash.Hash) *Tree {
	if len(hashes) == 0 {
		return nil
	}
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	nextPoT := nextPowerOfTwo(len(hashes))
	arraySize := nextPoT*2 - 1
	merkles := make([]*hash.Hash, arraySize)

	for i := range hashes {
		merkles[i] = &hashes[i]
	}

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil too.
		case merkles[i] == nil:
			merkles[offset] = nil

		// When there is no right child, the parent is generated by
		// hashing the concatenation of the left child with itself.
		case merkles[i+1] == nil:
			newHash := HashMerkleBranches(merkles[i], merkles[i])
			merkles[offset] = newHash

		// The normal case sets the parent node to the double sha256
		// of the concatenation of the left and right children.
		default:
			newHash := HashMerkleBranches(merkles[i], merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return &Tree{merkles: merkles}
}

func (tree *Tree) Root() hash.Hash {
	if tree == nil {
		return hash.UndefHash
	}
	h := tree.merkles[len(tree.merkles)-1]
	if h != nil {
		return *h
	}

	return hash.UndefHash
}

func (tree *Tree) Depth() int {
	if tree == nil {
		return 0
	}

	return int(math.Log2(float64(len(tree.merkles))))
}
