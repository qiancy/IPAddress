package ipaddr

import (
	"fmt"
	"math/big"
	"net"
	"unsafe"
)

const (
	HexPrefix                  = "0x"
	OctalPrefix                = "0"
	RangeSeparator             = '-'
	AlternativeRangeSeparator  = '\u00bb'
	SegmentWildcard            = '*'
	SegmentWildcardStr         = string(SegmentWildcard)
	AlternativeSegmentWildcard = '¿'
	SegmentSqlWildcard         = '%'
	SegmentSqlSingleWildcard   = '_'
)

type SegmentValueProvider func(segmentIndex int) SegInt

type addressCache struct {
	ip           net.IPAddr // lower converted (cloned when returned)
	lower, upper *Address
	//fromString   *HostIdentifierString xxxxx
	fromString unsafe.Pointer
	//fromString *IPAddressString
	fromHost *HostName
}

type addressInternal struct {
	section *AddressSection
	zone    Zone
	cache   *addressCache
}

func (addr *addressInternal) GetBitCount() BitCount {
	if addr.section == nil {
		return 0
	}
	return addr.section.GetBitCount()
}

func (addr *addressInternal) GetByteCount() int {
	if addr.section == nil {
		return 0
	}
	return addr.section.GetByteCount()
}

func (addr *addressInternal) GetCount() *big.Int {
	if addr.section == nil {
		return bigOne()
	}
	return addr.section.GetCount()
}

func (addr *addressInternal) GetPrefixCount() *big.Int {
	if addr.section == nil {
		return bigOne()
	}
	return addr.section.GetPrefixCount()
}

func (addr *addressInternal) GetPrefixCountLen(prefixLen BitCount) *big.Int {
	if addr.section == nil {
		return bigOne()
	}
	return addr.section.GetPrefixCountLen(prefixLen)
}

func (addr *addressInternal) IsMultiple() bool {
	return addr.section != nil && addr.section.IsMultiple()
}

func (addr *addressInternal) IsPrefixed() bool {
	return addr.section != nil && addr.section.IsPrefixed()
}

func (addr *addressInternal) GetPrefixLength() PrefixLen {
	if addr.section == nil {
		return nil
	}
	return addr.section.GetPrefixLength()
}

//func (addr *addressInternal) isMore(other *Address) int {
//	if addr.section == nil {
//		if other.IsMultiple() {
//			return -1
//		}
//		return 0
//	}
//	return addr.section.IsMore(other.GetSection())
//}

func (addr *addressInternal) IsMore(other AddressDivisionSeries) int {
	if addr.section == nil {
		if other.IsMultiple() {
			return -1
		}
		return 0
	}
	return addr.section.IsMore(other)
}

func (addr addressInternal) String() string { // using non-pointer receiver makes it work well with fmt
	if addr.zone != noZone {
		return fmt.Sprintf("%v%c%s", addr.section, IPv6ZoneSeparator, addr.zone)
	}
	return fmt.Sprintf("%v", addr.section)
}

func (addr *addressInternal) IsSequential() bool {
	if addr.section == nil {
		return true
	}
	return addr.section.IsSequential()
}

func (addr *addressInternal) getSegment(index int) *AddressSegment {
	return addr.section.GetSegment(index)
}

func (addr *addressInternal) GetValue() *big.Int {
	if addr.section == nil {
		return bigZero()
	}
	return addr.section.GetValue()
}

func (addr *addressInternal) GetUpperValue() *big.Int {
	if addr.section == nil {
		return bigZero()
	}
	return addr.section.GetUpperValue()
}

func (addr *addressInternal) GetIP() net.IP {
	return addr.GetBytes()
}

func (addr *addressInternal) CopyIP(bytes net.IP) net.IP {
	return addr.CopyBytes(bytes)
}

func (addr *addressInternal) GetUpperIP() net.IP {
	return addr.GetUpperBytes()
}

func (addr *addressInternal) CopyUpperIP(bytes net.IP) net.IP {
	return addr.CopyUpperBytes(bytes)
}

func (addr *addressInternal) GetBytes() []byte {
	if addr.section == nil {
		return emptyBytes
	}
	return addr.section.GetBytes()
}

func (addr *addressInternal) CopyBytes(bytes []byte) []byte {
	if addr.section == nil {
		if bytes != nil {
			return bytes
		}
		return emptyBytes
	}
	return addr.section.CopyBytes(bytes)
}

func (addr *addressInternal) GetUpperBytes() []byte {
	if addr.section == nil {
		return emptyBytes
	}
	return addr.section.GetUpperBytes()
}

func (addr *addressInternal) CopyUpperBytes(bytes []byte) []byte {
	if addr.section == nil {
		if bytes != nil {
			return bytes
		}
		return emptyBytes
	}
	return addr.section.CopyUpperBytes(bytes)
}

func (addr *addressInternal) checkIdentity(section *AddressSection) *Address {
	if section == addr.section {
		return addr.toAddress()
	}
	return &Address{addressInternal{section: section, zone: addr.zone, cache: &addressCache{}}}
}

func (addr *addressInternal) getLower() *Address {
	//TODO cache the result in the addressCache
	return addr.checkIdentity(addr.section.GetLower())
}

func (addr *addressInternal) getUpper() *Address {
	//TODO cache the result in the addressCache
	return addr.checkIdentity(addr.section.GetUpper())
}

func (addr *addressInternal) IsZero() bool {
	section := addr.section
	if section == nil {
		return true
	}
	return section.IsZero()
}

func (addr *addressInternal) IncludesZero() bool {
	section := addr.section
	if section == nil {
		return true
	}
	return section.IncludesZero()
}

func (addr *addressInternal) IsMax() bool {
	section := addr.section
	if section == nil {
		// when no bits, the only value 0 is the max value too
		return true
	}
	return section.IsMax()
}

func (addr *addressInternal) IncludesMax() bool {
	section := addr.section
	if section == nil {
		// when no bits, the only value 0 is the max value too
		return true
	}
	return section.IncludesMax()
}

func (addr *addressInternal) IsFullRange() bool {
	section := addr.section
	if section == nil {
		// when no bits, the only value 0 is the max value too
		return true
	}
	return section.IsFullRange()
}

func (addr *addressInternal) toAddress() *Address {
	return (*Address)(unsafe.Pointer(addr))
}

func (addr *addressInternal) hasNoDivisions() bool {
	return addr.section.hasNoDivisions()
}

func (addr *addressInternal) getDivision(index int) *AddressDivision {
	return addr.section.getDivision(index)
}

func (addr *addressInternal) getDivisionCount() int {
	if addr.section == nil {
		return 0
	}
	return addr.section.GetDivisionCount()
}

func (addr *addressInternal) toPrefixBlock() *Address {
	return addr.checkIdentity(addr.section.toPrefixBlock())
}

func (addr *addressInternal) toPrefixBlockLen(prefLen BitCount) *Address {
	return addr.checkIdentity(addr.section.toPrefixBlockLen(prefLen))
}

// isIPv4() returns whether this matches an IPv4 address.
// we allow nil receivers to allow this to be called following a failed conversion like ToIPAddress()
func (addr *addressInternal) isIPv4() bool {
	return addr != nil && addr.section != nil && addr.section.matchesIPv4Address()
}

// isIPv6() returns whether this matches an IPv6 address.
// we allow nil receivers to allow this to be called following a failed conversion like ToIPAddress()
func (addr *addressInternal) isIPv6() bool {
	return addr != nil && addr.section != nil && addr.section.matchesIPv6Address()
}

// isIPv6() returns whether this matches an IPv6 address.
// we allow nil receivers to allow this to be called following a failed conversion like ToIPAddress()
func (addr *addressInternal) isMAC() bool {
	return addr != nil && addr.section != nil && addr.section.matchesMACAddress()
}

// isIP() returns whether this matches an IP address.
// It must be IPv4, IPv6, or the zero IPAddress which has no segments
// we allow nil receivers to allow this to be called following a failed conversion like ToIPAddress()
func (addr *addressInternal) isIP() bool {
	return addr != nil && (addr.section == nil /* zero addr */ || addr.section.matchesIPAddress())
}

func (addr *addressInternal) CompareTo(item AddressItem) int {
	return CountComparator.Compare(addr, item)
}

func (addr *addressInternal) contains(other AddressType) bool {
	otherAddr := other.ToAddress()
	if addr.toAddress() == otherAddr {
		return true
	}
	otherSection := otherAddr.GetSection()
	if addr.section == nil {
		return otherSection.GetSegmentCount() == 0
	}
	return addr.section.Contains(otherSection) &&
		// if it is IPv6 and has a zone, then it does not contain addresses from other zones
		addr.isSameZone(other)
}

func (addr *addressInternal) equals(other AddressType) bool {
	otherAddr := other.ToAddress()
	if addr.toAddress() == otherAddr {
		return true
	}
	otherSection := other.ToAddress().GetSection()
	if addr.section == nil {
		return otherSection.GetSegmentCount() == 0
	}
	return addr.section.Equals(otherSection) &&
		// if it it is IPv6 and has a zone, then it does not equal addresses from other zones
		addr.isSameZone(other)
}

func (addr *addressInternal) isSameZone(other AddressType) bool {
	return addr.zone == other.ToAddress().zone
}

func (addr *addressInternal) getAddrType() addrType {
	if addr.section == nil {
		return zeroType
	}
	return addr.section.addrType
}

//TODO the four string methods at address level are toCanonicalString, toNormalizedString, toHexString, toCompressedString
// we also want toCanonicalWildcardString
// the code will need to check the addrtype in the section, in fact, the code should just defer to the section,
// although that's a bit problematic for ipv6.  So for ipv6, we need to scale up to ipv6 inside the address code,
// unfortunately, although this is just as messy for the java side where we had to make a special override for ipv6 everywhere
// And let's face it, we need to override all methods in addresses for the init() calls anyway
func (addr *addressInternal) ToCanonicalString() string {
	//TODO
	return ""
}

func (addr *addressInternal) ToCanonicalWildcardString() string {
	//TODO
	return ""
}

func (addr *addressInternal) ToNormalizedString() string {
	//TODO
	return ""
}

func (addr *addressInternal) ToNormalizedWildcardString() string {
	//TODO
	return ""
}

//
//
//protected abstract IPAddressStringParameters createFromStringParams();
//
//	protected IPAddressStringParameters createFromStringParams() {
//		return new IPAddressStringParameters.Builder().
//				getIPv4AddressParametersBuilder().setNetwork(getNetwork()).getParentBuilder().
//				getIPv6AddressParametersBuilder().setNetwork(getIPv6Network()).getParentBuilder().toParams();
//	}
//
//	protected IPAddressStringParameters createFromStringParams() {
//		return new IPAddressStringParameters.Builder().
//				getIPv4AddressParametersBuilder().setNetwork(getIPv4Network()).getParentBuilder().
//				getIPv6AddressParametersBuilder().setNetwork(getNetwork()).getParentBuilder().toParams();
//	}
//protected IPAddressString getAddressfromString() {
//	return (IPAddressString) fromString;
//}

var zeroAddr = &Address{
	addressInternal{
		section: zeroSection,
		cache:   &addressCache{},
	},
}

type Address struct {
	addressInternal
}

func (addr *Address) init() *Address {
	if addr.section == nil {
		return zeroAddr // this has a zero section rather that a nil section
	}
	return addr
}

func (addr *Address) Contains(other AddressType) bool {
	return addr.init().contains(other)
}

func (addr *Address) Equals(other AddressType) bool {
	return addr.init().equals(other)
}

func (addr *Address) String() string {
	return addr.init().addressInternal.String()
}

func (addr *Address) GetSection() *AddressSection {
	return addr.init().section
}

// Gets the subsection from the series starting from the given index
// The first segment is at index 0.
func (addr *Address) GetTrailingSection(index int) *AddressSection {
	return addr.GetSection().GetTrailingSection(index)
}

//// Gets the subsection from the series starting from the given index and ending just before the give endIndex
//// The first segment is at index 0.
func (addr *Address) GetSubSection(index, endIndex int) *AddressSection {
	return addr.GetSection().GetSubSection(index, endIndex)
}

// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
// into the given slice, as much as can be fit into the slice, returning the number of segments copied
func (addr *Address) CopySubSegments(start, end int, segs []*AddressSegment) (count int) {
	return addr.GetSection().CopySubSegments(start, end, segs)
}

// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
// into the given slice, as much as can be fit into the slice, returning the number of segments copied
func (addr *Address) CopySegments(segs []*AddressSegment) (count int) {
	return addr.GetSection().CopySegments(segs)
}

// GetSegments returns a slice with the address segments.  The returned slice is not backed by the same array as this section.
func (addr *Address) GetSegments() []*AddressSegment {
	return addr.GetSection().GetSegments()
}

// GetSegment returns the segment at the given index
func (addr *Address) GetSegment(index int) *AddressSegment {
	return addr.getSegment(index)
}

// GetSegmentCount returns the segment count
func (addr *Address) GetSegmentCount() int {
	return addr.getDivisionCount()
}

// GetGenericDivision returns the segment at the given index as an AddressGenericDivision
func (addr *Address) GetGenericDivision(index int) AddressGenericDivision {
	return addr.getDivision(index)
}

// GetDivision returns the segment count
func (addr *Address) GetDivisionCount() int {
	return addr.getDivisionCount()
}

func (addr *Address) GetLower() *Address {
	return addr.init().getLower()
}

func (addr *Address) GetUpper() *Address {
	return addr.init().getUpper()
}

func (addr *Address) ToPrefixBlock() *Address {
	return addr.init().toPrefixBlock()
}

func (addr *Address) ToAddressString() HostIdentifierString {
	if addr.isIP() {
		return addr.toAddress().ToIPAddress().ToAddressString()
	} else if addr.isMAC() {
		return addr.toAddress().ToMACAddress().ToAddressString()
	}
	return nil
}

func (addr *Address) IsIPv4() bool {
	return addr.isIPv4()
}

func (addr *Address) IsIPv6() bool {
	return addr.isIPv6()
}
func (addr *Address) IsIP() bool {
	return addr.isIP()
}
func (addr *Address) IsMAC() bool {
	return addr.isMAC()
}

func (addr *Address) ToAddress() *Address {
	return addr
}

func (addr *Address) ToIPAddress() *IPAddress {
	if addr.isIP() {
		return (*IPAddress)(unsafe.Pointer(addr))
	}
	return nil
}

func (addr *Address) ToIPv6Address() *IPv6Address {
	if addr.isIPv6() {
		return (*IPv6Address)(unsafe.Pointer(addr))
	}
	return nil
}

func (addr *Address) ToIPv4Address() *IPv4Address {
	if addr.isIPv4() {
		return (*IPv4Address)(unsafe.Pointer(addr))
	}
	return nil
}

func (addr *Address) ToMACAddress() *MACAddress {
	if addr.isMAC() {
		return (*MACAddress)(addr)
	}
	return nil
}
