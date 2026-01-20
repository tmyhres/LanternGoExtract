package fragments

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/datatypes"
)

// BspRegionType (0x29)
// Internal Name: None
// Associates a list of regions with a specified region type (Water, Lava, PvP or Zoneline).
type BspRegionType struct {
	BaseFragment
	// RegionTypes is the region type associated with the region
	RegionTypes []datatypes.RegionType
	// BspRegionIndices contains the indices of BSP regions
	BspRegionIndices []int
	// RegionString is the region type string
	RegionString string
	// Zoneline contains zoneline information if applicable
	Zoneline *datatypes.ZonelineInfo
}

// FragmentType returns the fragment type ID.
func (f *BspRegionType) FragmentType() uint32 {
	return 0x29
}

// Initialize parses the fragment data.
func (f *BspRegionType) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	_, err = reader.ReadInt32() // flags
	if err != nil {
		return fmt.Errorf("failed to read flags: %w", err)
	}

	regionCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read region count: %w", err)
	}

	f.BspRegionIndices = make([]int, regionCount)
	for i := int32(0); i < regionCount; i++ {
		idx, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read region index: %w", err)
		}
		f.BspRegionIndices[i] = int(idx)
	}

	regionStringSize, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read region string size: %w", err)
	}

	var regionTypeString string
	if regionStringSize == 0 {
		regionTypeString = strings.ToLower(f.Name)
	} else {
		encodedBytes, err := reader.ReadBytes(int(regionStringSize))
		if err != nil {
			return fmt.Errorf("failed to read region string: %w", err)
		}
		regionTypeString = strings.ToLower(DecodeString(encodedBytes))
	}

	f.RegionTypes = make([]datatypes.RegionType, 0)

	if strings.HasPrefix(regionTypeString, "wtn_") || strings.HasPrefix(regionTypeString, "wt_") {
		// Ex: wt_zone, wtn_XXXXXX
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeWater)
	} else if strings.HasPrefix(regionTypeString, "wtntp") {
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeWater)
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeZoneline)
		f.decodeZoneline(regionTypeString)
		f.RegionString = regionTypeString
	} else if strings.HasPrefix(regionTypeString, "lan_") || strings.HasPrefix(regionTypeString, "la_") {
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeLava)
	} else if strings.HasPrefix(regionTypeString, "lantp") {
		// TODO: Figure this out - soldunga
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeLava)
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeZoneline)
		f.decodeZoneline(regionTypeString)
		f.RegionString = regionTypeString
	} else if strings.HasPrefix(regionTypeString, "drntp") {
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeZoneline)
		f.decodeZoneline(regionTypeString)
		f.RegionString = regionTypeString
	} else if strings.HasPrefix(regionTypeString, "drp_") {
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypePvp)
	} else if strings.HasPrefix(regionTypeString, "drn_") {
		if strings.Contains(regionTypeString, "_s_") {
			f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeSlippery)
		} else {
			f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeUnknown)
		}
	} else if strings.HasPrefix(regionTypeString, "sln_") {
		// gukbottom, cazicthule (gumdrop), runnyeye, velketor
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeWaterBlockLos)
	} else if strings.HasPrefix(regionTypeString, "vwn_") {
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeFreezingWater)
	} else {
		// All trilogy client region types are accounted for
		// This is here in case newer clients have newer types
		// tox - "wt_zone" - Possible legacy water zonepoint for boat?
		f.RegionTypes = append(f.RegionTypes, datatypes.RegionTypeNormal)
	}

	return nil
}

func (f *BspRegionType) decodeZoneline(regionTypeString string) {
	f.Zoneline = &datatypes.ZonelineInfo{}

	// TODO: Verify this
	if regionTypeString == "drntp_zone" {
		f.Zoneline.Type = datatypes.ZonelineTypeReference
		f.Zoneline.Index = 0
		return
	}

	if len(regionTypeString) < 10 {
		return
	}

	zoneID, err := strconv.Atoi(regionTypeString[5:10])
	if err != nil {
		return
	}

	if zoneID == 255 {
		if len(regionTypeString) >= 16 {
			zonelineID, err := strconv.Atoi(regionTypeString[10:16])
			if err == nil {
				f.Zoneline.Type = datatypes.ZonelineTypeReference
				f.Zoneline.Index = zonelineID
			}
		}
		return
	}

	f.Zoneline.ZoneIndex = zoneID

	if len(regionTypeString) >= 31 {
		x := f.getValueFromRegionString(regionTypeString[10:16])
		y := f.getValueFromRegionString(regionTypeString[16:22])
		z := f.getValueFromRegionString(regionTypeString[22:28])
		rot, _ := strconv.Atoi(regionTypeString[28:31])

		f.Zoneline.Type = datatypes.ZonelineTypeAbsolute
		f.Zoneline.Position = datatypes.Vec3{X: x, Y: y, Z: z}
		f.Zoneline.Heading = rot
	}
}

func (f *BspRegionType) getValueFromRegionString(substring string) float32 {
	if strings.HasPrefix(substring, "-") {
		val, err := strconv.ParseFloat(substring[1:], 32)
		if err != nil {
			return 0
		}
		return -float32(val)
	}
	val, err := strconv.ParseFloat(substring, 32)
	if err != nil {
		return 0
	}
	return float32(val)
}

// LinkRegionType links this region type to BSP regions.
func (f *BspRegionType) LinkRegionType(bspRegions []*BspRegion) {
	for _, regionIndex := range f.BspRegionIndices {
		if regionIndex >= 0 && regionIndex < len(bspRegions) {
			bspRegions[regionIndex].SetRegionFlag(f)
		}
	}
}
