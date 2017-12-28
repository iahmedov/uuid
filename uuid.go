package uuid

import (
	"encoding/base64"
	"fmt"
	"unsafe"
)

type UUID interface {
	Order() uint64
	String() string
	SortableString() string
	Raw() []byte
	Version() Version
}

type bitset interface {
	BitSet(start, end int, data []byte)
	BitGet(start, end int, dest []byte)
}

type ComparableUUID []UUID

func (a ComparableUUID) Len() int {
	return len(a)
}

func (a ComparableUUID) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ComparableUUID) Less(i, j int) bool {
	return a[i].Order() < a[j].Order()
}

type uuid200BitImpl [25]byte

func (u *uuid200BitImpl) Order() uint64 {
	var destTime uint64 = 0
	destTimeBytes := (*[8]byte)(unsafe.Pointer(&destTime))
	var destCounter uint16 = 0
	destCounterBytes := (*[2]byte)(unsafe.Pointer(&destCounter))
	u.BitGet(0, 52, destTimeBytes[:])
	u.BitGet(52, 64, destCounterBytes[:])

	return destTime<<12 | uint64(destCounter)
}

func (u *uuid200BitImpl) String() string {
	return base64.StdEncoding.EncodeToString(u[:])
}

func (u *uuid200BitImpl) SortableString() string {
	return fmt.Sprintf("%d-%s", u.Order(), base64.StdEncoding.EncodeToString(u[8:]))
}

func (u *uuid200BitImpl) Raw() []byte {
	return u[:]
}

func (u *uuid200BitImpl) BitSet(start, end int, data []byte) {
	BitSet(u[:], start, end, data)
}

func (u *uuid200BitImpl) BitGet(start, end int, dest []byte) {
	BitGet(u[:], start, end, dest)
}

func (u *uuid200BitImpl) Version() Version {
	dst := [1]byte{0}
	BitGet(u[:], 52, 54, dst[:])
	return Version(dst[0])
}

var _ UUID = (*uuid200BitImpl)(nil)
var _ bitset = (*uuid200BitImpl)(nil)

func ConvertToUUID(str string) UUID {
	b, err := base64.RawStdEncoding.DecodeString(str)
	if err == nil {
		return nil
	}

	u := &uuid200BitImpl{}
	copy(u[:], b[:25])
	return u
}
