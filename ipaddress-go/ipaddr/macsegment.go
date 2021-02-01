package ipaddr

import (
	"math/big"
	"unsafe"
)

type MACSegInt uint8 //TODO consider changing to int16 later, because it makes arithmetic easier, in thigns like increment, or iterators, or spliterators

func ToMACSegInt(val SegInt) MACSegInt {
	return MACSegInt(val)
}

func newMACSegmentValues(value, upperValue MACSegInt) *macSegmentValues {
	return &macSegmentValues{value: value, upperValue: upperValue}
}

type macSegmentValues struct {
	value      MACSegInt
	upperValue MACSegInt
	cache      divCache
}

func (seg macSegmentValues) includesZero() bool {
	return seg.value == 0
}

func (seg macSegmentValues) includesMax() bool {
	return seg.upperValue == 0xff
}

func (seg macSegmentValues) isMultiple() bool {
	return seg.value != seg.upperValue
}

func (seg macSegmentValues) getCount() *big.Int {
	return big.NewInt(int64((seg.upperValue - seg.value)) + 1)
}

//func (seg macSegmentValues) GetSegmentPrefixLength() PrefixLen {
//	return nil
//}

func (seg macSegmentValues) GetBitCount() BitCount {
	return MACBitsPerSegment
}

func (seg macSegmentValues) GetByteCount() int {
	return MACBytesPerSegment
}

func (seg macSegmentValues) getValue() *big.Int {
	return big.NewInt(int64(seg.value))
}

func (seg macSegmentValues) getUpperValue() *big.Int {
	return big.NewInt(int64(seg.upperValue))
}

func (seg macSegmentValues) getDivisionValue() DivInt {
	return DivInt(seg.value)
}

func (seg macSegmentValues) getUpperDivisionValue() DivInt {
	return DivInt(seg.upperValue)
}

func (seg macSegmentValues) getDivisionPrefixLength() PrefixLen {
	//TODO for MAC this needs to be changed to getMinPrefixLengthForBlock
	return nil
}

func (seg macSegmentValues) getSegmentValue() SegInt {
	return SegInt(seg.value)
}

func (seg macSegmentValues) getUpperSegmentValue() SegInt {
	return SegInt(seg.upperValue)
}

func (seg macSegmentValues) calcBytesInternal() (bytes, upperBytes []byte) {
	bytes = []byte{byte(seg.value)}
	if seg.isMultiple() {
		upperBytes = []byte{byte(seg.upperValue)}
	} else {
		upperBytes = bytes
	}
	return
}

func (seg macSegmentValues) deriveNew(val, upperVal DivInt, prefLen PrefixLen) divisionValues {
	return newMACSegmentValues(MACSegInt(val), MACSegInt(upperVal))
}

func (seg macSegmentValues) deriveNewSeg(val, upperVal SegInt, prefLen PrefixLen) divisionValues {
	return newMACSegmentValues(MACSegInt(val), MACSegInt(upperVal))
}

func (seg macSegmentValues) getCache() *divCache {
	return &seg.cache
}

//func (seg macSegmentValues) getLower() (divisionValues, *divCache) {
//	return newMACSegmentValues(seg.value, seg.value)
//}
//
//func (seg macSegmentValues) getUpper() (divisionValues, *divCache) {
//	return newMACSegmentValues(seg.upperValue, seg.upperValue)
//}

var _ divisionValues = macSegmentValues{}

//var _ segmentValues = macSegmentValues{}

type MACAddressSegment struct {
	addressSegmentInternal
}

// We must override GetBitCount, GetByteCount and others for the case when we construct as the zero value

func (seg *MACAddressSegment) GetBitCount() BitCount {
	return IPv4BitsPerSegment
}

func (seg *MACAddressSegment) GetByteCount() int {
	return IPv4BytesPerSegment
}

func (seg *MACAddressSegment) GetMaxValue() MACSegInt {
	return 0xff
}

//func (seg *MACAddressSegment) ToAddressDivision() *AddressDivision {
//	return seg.ToAddressSegment().ToAddressDivision() xxx
//}

func (seg *MACAddressSegment) ToAddressSegment() *AddressSegment {
	if seg == nil {
		return nil
	}
	vals := seg.divisionValues
	if vals == nil {
		seg.divisionValues = macSegmentValues{}
	}
	return (*AddressSegment)(unsafe.Pointer(seg))
}

func NewMACSegment(val MACSegInt) *MACAddressSegment {
	return NewMACRangeSegment(val, val)
}

func NewMACRangeSegment(val, upperVal MACSegInt) *MACAddressSegment {
	return &MACAddressSegment{
		addressSegmentInternal{
			addressDivisionInternal{
				addressDivisionBase{newMACSegmentValues(val, upperVal)},
			},
		},
	}
}
