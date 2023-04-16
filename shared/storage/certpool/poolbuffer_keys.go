package certpool

import (
	"container/list"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/x509utils"
)

type pbKeys struct {
	keys  *list.List
	count int
}

func (p *pbKeys) Reset() {
	p.keys = list.New()
	p.count = 0
}

func (p *pbKeys) Count() int {
	return p.count
}

type pbKeyData struct {
	filename string
	pk       x509utils.PrivateKey
}

func (d *pbKeyData) Public() x509utils.PublicKey {
	return d.pk.Public().(x509utils.PublicKey)
}

func (d *pbKeyData) Validate() error {
	if v, ok := d.pk.(interface {
		Validate() error
	}); ok {
		return v.Validate()
	}
	return nil
}

func (pb *PoolBuffer) addKeyUnlocked(fn string, pk x509utils.PrivateKey) error {
	if pk != nil {
		pd := &pbKeyData{
			filename: fn,
			pk:       pk,
		}

		if err := pb.printKey(fn, pk); err != nil {
			return err
		}

		if err := pd.Validate(); err != nil {
			return err
		}

		// store
		if pb.keys.keys == nil {
			pb.keys.Reset()
		}

		pb.keys.keys.PushBack(pd)
		pb.keys.count++
	}
	return nil
}

// Keys returns an array of all stored Private Keys
func (pb *PoolBuffer) Keys() []x509utils.PrivateKey {
	out := make([]x509utils.PrivateKey, 0, pb.keys.count)
	core.ListForEach(pb.keys.keys, func(pk x509utils.PrivateKey) bool {
		if pk != nil {
			out = append(out, pk)
		}

		return false // continue
	})
	return out
}
