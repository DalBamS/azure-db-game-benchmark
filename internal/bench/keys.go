package bench

import "math/rand/v2"

// KeyPicker draws account ids with a "hot set" model:
//   with probability HotProb  -> uniform over the first HotKeys ids (the "active players")
//   otherwise                 -> uniform over [1, N]
// This is documented, reproducible and keeps the top-1 key share tiny while still
// concentrating a configurable share of traffic on a small working set.
type KeyPicker struct {
	N       int64
	HotKeys int64
	HotProb float64
	rng     *rand.Rand
}

func NewKeyPicker(n, hotKeys int64, hotProb float64, seed uint64) *KeyPicker {
	if hotKeys < 1 {
		hotKeys = 1
	}
	if hotKeys > n {
		hotKeys = n
	}
	return &KeyPicker{N: n, HotKeys: hotKeys, HotProb: hotProb, rng: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))}
}

func (k *KeyPicker) Pick() int64 {
	if k.rng.Float64() < k.HotProb {
		return 1 + k.rng.Int64N(k.HotKeys)
	}
	return 1 + k.rng.Int64N(k.N)
}

func (k *KeyPicker) Rng() *rand.Rand { return k.rng }

// TopShare returns the analytical share of traffic received by the single hottest key
// and by the hot set as a whole, for documentation in results.
func (k *KeyPicker) TopShare() (top1 float64, hotSet float64) {
	top1 = k.HotProb/float64(k.HotKeys) + (1-k.HotProb)/float64(k.N)
	hotSet = k.HotProb + (1-k.HotProb)*float64(k.HotKeys)/float64(k.N)
	return
}
