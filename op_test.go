package uuid

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
	"unsafe"
)

func BenchmarkTimeNow(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = time.Now().UnixNano() / int64(time.Microsecond)
	}
}

func BenchmarkBase64(b *testing.B) {
	bytes := []byte("1234567890123456789012345")
	for n := 0; n < b.N; n++ {
		a := base64.StdEncoding.EncodeToString(bytes)
		_ = a
	}
}

func BenchmarkBase64Granular(b *testing.B) {
	bytes := []byte("1234567890123456789012345")
	for n := 0; n < b.N; n++ {
		l := base64.StdEncoding.EncodedLen(25)
		buf := make([]byte, l)
		base64.StdEncoding.Encode(buf, bytes)
		a := string(buf)
		_ = a
	}
}

func BenchmarkCryptoRand(bench *testing.B) {
	for n := 0; n < bench.N; n++ {
		c := 15
		b := make([]byte, c)
		_, err := crand.Read(b)
		if err != nil {
			panic("failed")
		}
	}
}

func BenchmarkRandomInt64(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var max int64 = 1<<63 - 1
	for n := 0; n < b.N; n++ {
		r.Int63n(max)
		r.Int63n(max)
	}
}

func BenchmarkRandomInt32(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// var max int32 = 1<<31 - 1
	for n := 0; n < b.N; n++ {
		r.Int31()
		r.Int31()
	}
}

func BenchmarkRandomInt32n(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var max int32 = 1<<31 - 1
	for n := 0; n < b.N; n++ {
		r.Int31n(max)
		r.Int31n(max)
	}
}

func BenchmarkUnsafe(b *testing.B) {
	var x int64 = 134348800 //1<<60-1
	for n := 0; n < b.N; n++ {
		data := *(*[8]byte)(unsafe.Pointer(&x))
		_ = data
	}
}

func BenchmarkCastAlias(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = []byte{byte(V1)}
	}
}

func BenchmarkAccessTable(b *testing.B) {
	table := [8]byte{0x01, 0x03, 0x07, 0x0F, 0x1F, 0x3F, 0x7F, 0xFF}
	for n := 0; n < b.N; n++ {
		_ = table[0]
		_ = table[1]
		_ = table[2]
		_ = table[3]
		_ = table[4]
		_ = table[5]
		_ = table[6]
		_ = table[7]
	}
}

func BenchmarkAccessBitwise(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = mask(1)
		_ = mask(2)
		_ = mask(3)
		_ = mask(4)
		_ = mask(5)
		_ = mask(6)
		_ = mask(7)
		_ = mask(8)
	}
}

func BenchmarkBitset(b *testing.B) {
	dst := []byte{0, 0}
	for n := 0; n < b.N; n++ {
		BitSet(dst[:], 5, 13, []byte{0xFF, 0XFF})
	}
}

func BenchmarkBitget(b *testing.B) {
	dst := []byte{0, 0}
	src := []byte{0, 0}
	for n := 0; n < b.N; n++ {
		BitGet(src, 5, 13, dst)
	}
}

func BenchmarkUuid2Gen(b *testing.B) {
	gen := NewGenerator(nil)
	for n := 0; n < b.N; n++ {
		gen.Next()
	}
}

func BenchmarkUuid2GetOrder(b *testing.B) {
	gen := NewGenerator(nil)
	uuid := gen.Next()
	for n := 0; n < b.N; n++ {
		uuid.Order()
	}
}

func TestUuidBitset(t *testing.T) {
	now := time.Now().UnixNano() / int64(time.Microsecond)
	timeBytes := *(*[8]byte)(unsafe.Pointer(&now))
	counter := 1234 // atomic.AddInt32(&g.rotation, 1)
	counterBytes := *(*[4]byte)(unsafe.Pointer(&counter))

	uuid := &uuid200BitImpl{}
	uuid.BitSet(0, 52, timeBytes[:])
	uuid.BitSet(52, 64, counterBytes[:])

	res := uuid.Order()
	gotTime := res & ((1<<52 - 1) << 12) >> 12
	if gotTime != uint64(now) {
		t.Fatalf("stored values are not same, expected: %d, got: %d (%d)", now, gotTime, res)
	}
}

func TestUuidKindOfCollision(t *testing.T) {
	gen := NewGenerator(nil)
	loop := 100000

	items := make([]UUID, loop)
	for i := 0; i < loop; i++ {
		items[i] = gen.Next()
	}

	mp := map[string]UUID{}
	for i := 0; i < loop; i++ {
		// fmt.Println(items[i])
		mp[items[i].String()] = items[i]
	}

	if len(mp) != loop {
		t.Fatalf("COLLISION!!!!!")
	}

	for i := 0; i < loop; i++ {
		u0 := ConvertToUUID(items[i].String())
		if u0.Order() != items[i].Order() {
			t.Fatalf("could not parse from str to UUID")
		}
	}
}

func TestUuidGenerationTime(t *testing.T) {
	gen := NewGenerator(nil)
	loop := 250000

	items := make([]UUID, loop)
	start := time.Now()
	for i := 0; i < loop; i++ {
		items[i] = gen.Next()
	}

	duration := time.Since(start)
	if duration > time.Second {
		t.Fatalf("requirements not met :(")
	}

	fmt.Printf("time taken to create %d items: %s\n", loop, duration)
}

func TestUuidSortableString(t *testing.T) {
	gen := NewGenerator(nil)
	loop := 100

	items := make([]UUID, loop)
	for i := 0; i < loop; i++ {
		items[i] = gen.Next()
	}
	sort.Sort(ComparableUUID(items))
	// for i := 0; i < loop; i++ {
	// 	fmt.Printf("%s\n", items[i].SortableString())
	// }
}
