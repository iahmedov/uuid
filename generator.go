package uuid

import (
	"math/rand"
	"sync/atomic"
	"time"
	"unsafe"
)

type Generator interface {
	Next() UUID
}

type Version byte

const (
	V1 Version = 0x01
	V2 Version = 0x02 // 2bit
)

type uuidGeneratorV1 struct {
	// Format (in bits):
	// 52  - time in microseconds, enough for ~1.3*47 years
	// 12  - counter, which rotates in range (0-4095) while generating UUIDs
	//		 since our generator takes almost 1microsecond to generate one ID
	//       we will have enough time to increment this value, in order to
	//       distinguish 2 UUID generated at one moment
	// 2   - version identifier
	// 16  - prefix, used for distributed/multiprocess generation of UUIDs,
	//		 adds some collision to UUID generated in different machines/processes
	// 118 - randomly generated data
	prefix   [2]byte
	rotation int32
	max59bit int64
	r        *rand.Rand
}

type uuidGeneratorV2 struct {
	// Format (in bits):
	// 52  - time in microseconds, enough for yet another ~1.3*47 years
	// 12  - counter, which rotates in range (0-4095) while generating UUIDs
	//		 since our generator takes almost 1microsecond to generate one ID
	//       we will have enough time to increment this value, in order to
	//       distinguish 2 UUID generated at one moment
	// 2   - version identifier
	// 134 - randomly generated data (60,60,14)
	rotation int32
	max60bit int64
	max14bit int32
	r        *rand.Rand
}

func NewGenerator(prefix *[2]byte) Generator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if prefix != nil {
		return &uuidGeneratorV1{
			prefix:   *prefix,
			rotation: 0,
			max59bit: 1<<59 - 1,
			r:        r,
		}
	} else {
		return &uuidGeneratorV2{
			rotation: 0,
			max60bit: 1<<60 - 1,
			max14bit: 1<<14 - 1,
			r:        r,
		}
	}
}

func (g *uuidGeneratorV1) Next() UUID {
	now := time.Now().UnixNano() / int64(time.Microsecond)
	timeBytes := *(*[8]byte)(unsafe.Pointer(&now))
	counter := atomic.AddInt32(&g.rotation, 1)
	counterBytes := *(*[4]byte)(unsafe.Pointer(&counter))

	uuid := &uuid200BitImpl{}
	uuid.BitSet(0, 52, timeBytes[:])
	uuid.BitSet(52, 64, counterBytes[:])
	uuid.BitSet(64, 66, []byte{byte(V1)})
	uuid.BitSet(66, 82, g.prefix[:])
	// 118 bits, could be generated in two rands
	rand1 := g.r.Int63n(g.max59bit)
	rand2 := g.r.Int63n(g.max59bit)

	rand1Bytes := *(*[8]byte)(unsafe.Pointer(&rand1))
	rand2Bytes := *(*[8]byte)(unsafe.Pointer(&rand2))
	uuid.BitSet(82, 141, rand1Bytes[:])
	uuid.BitSet(141, 200, rand2Bytes[:])

	return uuid
}

func (g *uuidGeneratorV2) Next() UUID {
	now := time.Now().UnixNano() / int64(time.Microsecond)
	timeBytes := *(*[8]byte)(unsafe.Pointer(&now))
	counter := atomic.AddInt32(&g.rotation, 1)
	counterBytes := *(*[4]byte)(unsafe.Pointer(&counter))

	uuid := &uuid200BitImpl{}
	uuid.BitSet(0, 52, timeBytes[:])
	uuid.BitSet(52, 64, counterBytes[:])
	uuid.BitSet(64, 66, []byte{byte(V2)})
	// 134 bits, could be generated in 3 rands
	rand1 := g.r.Int63n(g.max60bit)
	rand2 := g.r.Int63n(g.max60bit)
	rand3 := g.r.Int31n(g.max14bit)

	rand1Bytes := *(*[12]byte)(unsafe.Pointer(&rand1))
	rand2Bytes := *(*[12]byte)(unsafe.Pointer(&rand2))
	rand3Bytes := *(*[4]byte)(unsafe.Pointer(&rand3))
	uuid.BitSet(66, 126, rand1Bytes[:])
	uuid.BitSet(126, 186, rand2Bytes[:])
	uuid.BitSet(186, 200, rand3Bytes[:])

	return uuid
}

var _ Generator = (*uuidGeneratorV1)(nil)
var _ Generator = (*uuidGeneratorV2)(nil)
