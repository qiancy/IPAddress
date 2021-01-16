package ipaddr

import (
	"fmt"
	"unsafe"
)

var zeroSection AddressSection

//var zeroSegs = []*AddressSegment{}
//var zeroDivs = []*AddressDivision{}

func createSection(segments []*AddressDivision, prefixLength PrefixLen, addrType addrType, startIndex uint8) *AddressSection {
	return &AddressSection{
		addressSectionInternal{
			addressDivisionGroupingInternal{
				divisions:           segments,
				prefixLength:        prefixLength,
				addrType:            addrType,
				addressSegmentIndex: startIndex,
				cache:               &valueCache{},
			},
		},
	}
}

func createMultipleSection(segments []*AddressDivision, prefixLength PrefixLen, addrType addrType, startIndex uint8, isMultiple bool) *AddressSection {
	result := createSection(segments, prefixLength, addrType, startIndex)
	result.isMultiple = isMultiple
	return result
}

func createInitializedSection(segments []*AddressDivision, prefixLength PrefixLen, addrType addrType, startIndex uint8) *AddressSection {
	result := createSection(segments, prefixLength, addrType, startIndex)
	result.init() // assigns isMultiple
	return result
}

/*
// TODO MAC will need something like this when calculating prefix length on creation
//func (grouping *addressDivisionGroupingInternal) getPrefixLengthCacheLocked() PrefixLen {
//		count := grouping.GetDivisionCount()
//		bitsSoFar, prefixBits := BitCount(0), BitCount(0)
//		hasPrefix := false
//		for i := 0; i < count; i++ {
//			div := grouping.getDivision(i)
//			divPrefLen := div.getDivisionPrefixLength() //TODO for MAC this needs to be changed to getMinPrefixLengthForBlock (optimize it to check for full range or single value first )
//			if hasPrefix = divPrefLen != nil; hasPrefix {
//				divPrefBits := *divPrefLen
//				if !hasPrefix || divPrefBits != 0 {
//					prefixBits = bitsSoFar + divPrefBits
//				}
//				if grouping.addrType.alignsPrefix() {
//					break
//				}
//			}
//			bitsSoFar += div.GetBitCount()
//		}
//		if hasPrefix {
//			res := &prefixBits
//			prefLen.value = res
//		}
//}
*/

//////////////////////////////////////////////////////////////////
//
//
//
type addressSectionInternal struct {
	addressDivisionGroupingInternal
}

func (section *addressSectionInternal) toAddressSection() *AddressSection {
	return (*AddressSection)(unsafe.Pointer(section))
}

func (section *addressSectionInternal) ToAddressDivisionGrouping() *AddressDivisionGrouping {
	return (*AddressDivisionGrouping)(unsafe.Pointer(section))
}

// error returned for nil sements, or inconsistent prefixes
func (section *addressSectionInternal) init() error {
	segCount := section.GetSegmentCount()
	if segCount == 0 {
		var previousSegmentPrefix PrefixLen
		isMultiple := false
		bitsPerSegment := section.GetBitsPerSegment()
		for i := 0; i < segCount; i++ {
			div := section.getDivision(i)
			if div == nil {
				return &addressException{"ipaddress.error.null.segment"}
			}
			// unnecessary since we can control the division type
			// new ipv4/6 sections are created from ipv4/6segment while derived sections come from existing segments
			// in all cases, no way to insert mimatched divisions
			//else if section.getDivision(i).GetBitCount() != bitsPerSegment {
			//	return &addressException{"ipaddress.error.mismatched.bit.size"}
			//}

			segment := section.GetSegment(i)
			if !isMultiple && segment.isMultiple() {
				isMultiple = true
				section.isMultiple = true
			}

			//Calculate the segment-level prefix
			//
			//Across an address prefixes are:
			//IPv6: (null):...:(null):(1 to 16):(0):...:(0)
			//or IPv4: ...(null).(1 to 8).(0)...
			//For MAC, all segs have nil prefix since prefix is not segment-level
			//For MAC, prefixes must be derived in other ways, not from individual segment prefix values,
			// either using
			segPrefix := segment.getDivisionPrefixLength()
			if previousSegmentPrefix == nil {
				if segPrefix != nil {
					section.prefixLength = getNetworkPrefixLength(bitsPerSegment, *segPrefix, i)
				}
			} else if segPrefix == nil || *segPrefix != 0 {
				return &inconsistentPrefixException{str: fmt.Sprintf("%v %v %v", section.GetSegment(i-1), segment, segPrefix), key: "ipaddress.error.inconsistent.prefixes"}
			}
			previousSegmentPrefix = segPrefix
		}
	}
	return nil
}

func (section *addressSectionInternal) GetBitsPerSegment() BitCount {
	if section.GetDivisionCount() == 0 {
		return 0
	}
	return section.getDivision(0).GetBitCount()
}

func (section *addressSectionInternal) GetBytesPerSegment() int {
	if section.GetDivisionCount() == 0 {
		return 0
	}
	return section.getDivision(0).GetByteCount()
}

func (section *addressSectionInternal) GetSegment(index int) *AddressSegment {
	return section.getDivision(index).ToAddressSegment()
}

func (section *addressSectionInternal) GetSegmentCount() int {
	return section.GetDivisionCount()
}

func (section *addressSectionInternal) GetBitCount() BitCount {
	divLen := len(section.divisions)
	if divLen == 0 {
		return 0
	}
	return section.getDivision(0).GetBitCount() * BitCount(section.GetSegmentCount())
}

func (section *addressSectionInternal) GetByteCount() int {
	return int((section.GetBitCount() + 7) >> 3)
}

func (section *addressSectionInternal) matchesIPv6Section() bool {
	return section.addrType.isIPv6() || section.addrType.isNil()
}

func (section *addressSectionInternal) matchesIPv4Section() bool {
	return section.addrType.isIPv4() || section.addrType.isNil()
}

func (section *addressSectionInternal) matchesIPSection() bool {
	return section.addrType.isIP() || section.addrType.isNil()
}

func (section *addressSectionInternal) matchesMACSection() bool {
	return section.addrType.isMAC() || section.addrType.isNil()
}

func (section *addressSectionInternal) matchesIPv6Address() bool {
	return section.addrType.isIPv6() && section.GetSegmentCount() == IPv6SegmentCount
}

func (section *addressSectionInternal) matchesIPv4Address() bool {
	return section.addrType.isIPv4() && section.GetSegmentCount() == IPv4SegmentCount
}

func (section *addressSectionInternal) matchesMACAddress() bool {
	segCount := section.GetSegmentCount()
	return section.addrType.isMAC() &&
		(segCount == MediaAccessControlSegmentCount || segCount == ExtendedUniqueIdentifier64SegmentCount)
}

// Gets the subsection from the series starting from the given index and ending just before the give endIndex
// The first segment is at index 0.
func (section *addressSectionInternal) getSubSection(index, endIndex int) *AddressSection {
	if index < 0 {
		index = 0
	}
	thisSegmentCount := section.GetSegmentCount()
	if endIndex < thisSegmentCount {
		endIndex = thisSegmentCount
	}
	segmentCount := endIndex - index
	if segmentCount <= 0 {
		if thisSegmentCount == 0 {
			return section.toAddressSection()
		}
		return &zeroSection
	}
	if index == 0 && endIndex == thisSegmentCount {
		return section.toAddressSection()
	}
	segs := createSegmentArray(segmentCount)
	section.copySubSegmentsToSlice(index, endIndex, segs)
	newPrefLen := section.GetPrefixLength()
	if newPrefLen != nil && index != 0 {
		newPrefLen = getPrefixedSegmentPrefixLength(section.GetBitsPerSegment(), *newPrefLen, index)
	}
	newStartIndex := section.addressSegmentIndex + uint8(index)
	addrType := section.addrType
	if !section.IsMultiple() {
		return createSection(segs, newPrefLen, addrType, newStartIndex)
	}
	return createInitializedSection(segs, newPrefLen, addrType, newStartIndex)
}

func (section *addressSectionInternal) copySegmentsToSlice(divs []*AddressDivision) (count int) {
	return section.visitSegments(func(index int, div *AddressDivision) bool { divs[index] = div; return false }, len(divs))
}

func (section *addressSectionInternal) visitSegments(target func(index int, div *AddressDivision) bool, targetLen int) (count int) {
	if section.hasNoDivisions() {
		return
	}
	divs := section.divisions
	count = len(divs)
	if count > targetLen {
		count = targetLen
	}
	for start := 0; start < count; start++ {
		if target(start, divs[start]) {
			break
		}
	}
	return
}

func (section *addressSectionInternal) copySubSegmentsToSlice(start, end int, divs []*AddressDivision) (count int) {
	return section.visitSubSegments(start, end, func(index int, div *AddressDivision) bool { divs[index] = div; return false }, len(divs))
}

func (section *addressSectionInternal) visitSubSegments(start, end int, target func(index int, div *AddressDivision) (stop bool), targetLen int) (count int) {
	if section.hasNoDivisions() {
		return
	}
	divs := section.divisions
	targetIndex := 0
	if start < 0 {
		targetIndex -= start
		start = 0
		if targetIndex >= targetLen {
			return
		}
	}
	// how many to copy?
	sourceLen := len(divs)
	if end > sourceLen {
		end = sourceLen
	}
	calcCount := end - start
	if calcCount <= 0 {
		return
	}
	// if not enough space, adjust count and end
	space := targetLen - targetIndex
	if calcCount > space {
		count = space
		end = start + space
	} else {
		count = calcCount
	}
	// now copy
	for start < end {
		if target(targetIndex, divs[start]) {
			break
		}
		targetIndex++
		start++
	}
	return
}

func (section *addressSectionInternal) getLowestOrHighestSection(lowest bool) (result *AddressSection) { //TODO move this too
	if !section.IsMultiple() {
		return section.toAddressSection()
	}
	cache := section.cache
	sectionCache := &cache.sectionCache
	cache.RLock()
	if lowest {
		result = sectionCache.lower
	} else {
		result = sectionCache.upper
	}
	cache.RUnlock()
	if result != nil {
		return
	}
	cache.Lock()
	if lowest {
		result = sectionCache.lower
		if result == nil {
			result = section.createLowestOrHighestSectionCacheLocked(lowest)
			sectionCache.lower = result
		}
	} else {
		result = sectionCache.upper
		if result == nil {
			result = section.createLowestOrHighestSectionCacheLocked(lowest)
			sectionCache.upper = result
		}
	}
	cache.Unlock()
	return
}

func (section *addressSectionInternal) createLowestOrHighestSectionCacheLocked(lowest bool) *AddressSection {
	segmentCount := section.GetSegmentCount()
	segs := createSegmentArray(segmentCount)
	if lowest {
		for i := 0; i < segmentCount; i++ {
			segs[i] = section.GetSegment(i).GetLower().ToAddressDivision()
		}
	} else {
		for i := 0; i < segmentCount; i++ {
			segs[i] = section.GetSegment(i).GetUpper().ToAddressDivision()
		}
	}
	return createSection(segs, section.prefixLength, section.addrType, section.addressSegmentIndex)
}

func (section *addressSectionInternal) toPrefixBlock() *AddressSection {
	prefixLength := section.GetPrefixLength()
	if prefixLength == nil {
		return section.toAddressSection()
	}
	return section.toPrefixBlockLen(*prefixLength)
}

func (section *addressSectionInternal) toPrefixBlockLen(prefLen BitCount) *AddressSection {
	bitCount := section.GetBitCount()
	if prefLen < 0 {
		prefLen = 0
	} else {
		if prefLen > bitCount {
			prefLen = bitCount
		}
	}
	segCount := section.GetSegmentCount()
	if segCount == 0 {
		return section.toAddressSection()
	}
	segmentByteCount := section.GetBytesPerSegment()
	segmentBitCount := section.GetBitsPerSegment()
	existingPrefixLength := section.GetPrefixLength()
	prefixMatches := existingPrefixLength != nil && *existingPrefixLength == prefLen
	if prefixMatches {
		prefixedSegmentIndex := getHostSegmentIndex(prefLen, segmentByteCount, segmentBitCount)
		if prefixedSegmentIndex >= segCount {
			return section.toAddressSection()
		}
		segPrefLength := *getPrefixedSegmentPrefixLength(segmentBitCount, prefLen, prefixedSegmentIndex)
		seg := section.GetSegment(prefixedSegmentIndex)
		if seg.containsPrefixBlock(segPrefLength) {
			i := prefixedSegmentIndex + 1
			for ; i < segCount; i++ {
				seg = section.GetSegment(i)
				if !seg.IsFullRange() {
					break
				}
			}
			if i == segCount {
				return section.toAddressSection()
			}
		}
	}
	prefixedSegmentIndex := 0
	newSegs := createSegmentArray(segCount)
	if prefLen > 0 {
		prefixedSegmentIndex = getNetworkSegmentIndex(prefLen, segmentByteCount, segmentBitCount)
		copy(newSegs, section.divisions[:prefixedSegmentIndex])
	}
	for i := prefixedSegmentIndex; i < segCount; i++ {
		segPrefLength := getPrefixedSegmentPrefixLength(segmentBitCount, prefLen, i)
		oldSeg := section.divisions[i]
		newSegs[i] = oldSeg.toPrefixedNetworkDivision(segPrefLength)
	}
	//TODO caching of prefLen?  we should map it to a global array - check what we have in the validation code
	return createMultipleSection(newSegs, &prefLen, section.addrType, section.addressSegmentIndex, section.isMultiple || prefLen < bitCount)
}

//
//
//
//
type AddressSection struct {
	addressSectionInternal
}

// Gets the subsection from the series starting from the given index
// The first segment is at index 0.
func (section *AddressSection) GetTrailingSection(index int) *AddressSection {
	return section.getSubSection(index, section.GetSegmentCount())
}

//// Gets the subsection from the series starting from the given index and ending just before the give endIndex
//// The first segment is at index 0.
func (section *AddressSection) GetSubSection(index, endIndex int) *AddressSection {
	return section.getSubSection(index, endIndex)
}

// ForEachSegment calls the given callback for each segment, terminating early if a callback returns true
func (section *AddressSection) ForEachSegment(callback func(index int, segment *AddressSegment) (stop bool)) {
	section.visitSegments(
		func(index int, div *AddressDivision) bool {
			return callback(index, div.ToAddressSegment())
		},
		section.GetSegmentCount())
}

// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
// into the given slice, as much as can be fit into the slice, returning the number of segments copied
func (section *AddressSection) CopySubSegments(start, end int, segs []*AddressSegment) (count int) {
	return section.visitSubSegments(start, end, func(index int, div *AddressDivision) bool { segs[index] = div.ToAddressSegment(); return false }, len(segs))
}

// CopySubSegments copies the existing segments from the given start index until but not including the segment at the given end index,
// into the given slice, as much as can be fit into the slice, returning the number of segments copied
func (section *AddressSection) CopySegments(segs []*AddressSegment) (count int) {
	return section.visitSegments(func(index int, div *AddressDivision) bool { segs[index] = div.ToAddressSegment(); return false }, len(segs))
}

// GetSegments returns a slice with the address segments.  The returned slice is not backed by the same array as this section.
func (section *AddressSection) GetSegments() (res []*AddressSegment) {
	res = make([]*AddressSegment, section.GetSegmentCount())
	section.CopySegments(res)
	return
}

func (section *AddressSection) GetLower() *AddressSection {
	return section.getLowestOrHighestSection(true)
}

func (section *AddressSection) GetUpper() *AddressSection {
	return section.getLowestOrHighestSection(false)
}

func (section *AddressSection) ToPrefixBlock() *AddressSection {
	return section.toPrefixBlock()
}

func (section *AddressSection) ToPrefixBlockLen(prefLen BitCount) *AddressSection {
	return section.toPrefixBlockLen(prefLen)
}

func (section *AddressSection) ToIPAddressSection() *IPAddressSection {
	if section == nil || !section.matchesIPSection() {
		return nil
	}
	return (*IPAddressSection)(unsafe.Pointer(section))
}

func (section *AddressSection) ToIPv6AddressSection() *IPv6AddressSection {
	if section == nil || !section.matchesIPv6Section() {
		return nil
	}
	return (*IPv6AddressSection)(unsafe.Pointer(section))
}

func (section *AddressSection) ToIPv4AddressSection() *IPv4AddressSection {
	if section == nil || !section.matchesIPv4Section() {
		return nil
	}
	return (*IPv4AddressSection)(unsafe.Pointer(section))
}

func (section *AddressSection) ToMACAddressSection() *MACAddressSection {
	if section == nil || !section.matchesMACSection() {
		return nil
	}
	return (*MACAddressSection)(unsafe.Pointer(section))
}