package exporters

import (
	"fmt"
	"strings"

	"github.com/tmyhres/LanternGoExtract/pkg/wld/datatypes"
	"github.com/tmyhres/LanternGoExtract/pkg/wld/fragments"
)

// BspTreeWriter exports BSP tree data to a text format.
type BspTreeWriter struct {
	TextAssetWriter
}

// NewBspTreeWriter creates a new BspTreeWriter.
func NewBspTreeWriter() *BspTreeWriter {
	w := &BspTreeWriter{}
	w.AppendLine(ExportHeaderTitle + "BSP Tree")
	w.AppendLine(ExportHeaderFormat + "Normal nodes: NormalX, NormalY, NormalZ, SplitDistance, LeftNodeId, RightNodeId")
	w.AppendLine(ExportHeaderFormat + "Leaf nodes: BSPRegionId, RegionType")
	return w
}

// AddFragmentData adds BSP tree fragment data to the export buffer.
func (w *BspTreeWriter) AddFragmentData(data fragments.Fragment) {
	tree, ok := data.(*fragments.BspTree)
	if !ok || tree == nil {
		return
	}

	for _, node := range tree.Nodes {
		// Normal node
		if node.Region == nil {
			w.AppendString(formatFloat(node.NormalX))
			w.AppendString(",")
			w.AppendString(formatFloat(node.NormalZ))
			w.AppendString(",")
			w.AppendString(formatFloat(node.NormalY))
			w.AppendString(",")
			w.AppendString(formatFloat(node.SplitDistance))
			w.AppendString(",")
			w.AppendString(fmt.Sprintf("%d", node.LeftNode))
			w.AppendString(",")
			w.AppendLine(fmt.Sprintf("%d", node.RightNode))
		} else {
			// Leaf node
			w.AppendString(fmt.Sprintf("%d", node.RegionID))
			w.AppendString(",")

			var types string
			region, ok := node.Region.(*fragments.BspRegion)
			if ok && region != nil && region.RegionType != nil {
				typeStrs := make([]string, len(region.RegionType.RegionTypes))
				for i, rt := range region.RegionType.RegionTypes {
					typeStrs[i] = regionTypeToString(rt)
				}
				types = strings.Join(typeStrs, ";")
			} else {
				types = regionTypeToString(datatypes.RegionTypeNormal)
			}

			w.AppendString(types)

			if region == nil || region.RegionType == nil {
				w.AppendString("\n")
				continue
			}

			// Check for zoneline
			hasZoneline := false
			for _, rt := range region.RegionType.RegionTypes {
				if rt == datatypes.RegionTypeZoneline {
					hasZoneline = true
					break
				}
			}

			if hasZoneline && region.RegionType.Zoneline != nil {
				zoneline := region.RegionType.Zoneline

				w.AppendString(",")
				w.AppendString(zonelineTypeToString(zoneline.Type))
				w.AppendString(",")

				if zoneline.Type == datatypes.ZonelineTypeReference {
					w.AppendString(fmt.Sprintf("%d", zoneline.Index))
				} else {
					w.AppendString(fmt.Sprintf("%d", zoneline.ZoneIndex))
					w.AppendString(",")
					w.AppendString(formatFloat(zoneline.Position.X))
					w.AppendString(",")
					w.AppendString(formatFloat(zoneline.Position.Y))
					w.AppendString(",")
					w.AppendString(formatFloat(zoneline.Position.Z))
					w.AppendString(",")
					w.AppendString(fmt.Sprintf("%d", zoneline.Heading))
				}
			}

			w.AppendString("\n")
		}
	}
}

// formatFloat formats a float32 with decimal point separator.
func formatFloat(f float32) string {
	return fmt.Sprintf("%g", f)
}

// regionTypeToString converts a RegionType to its string representation.
func regionTypeToString(rt datatypes.RegionType) string {
	switch rt {
	case datatypes.RegionTypeNormal:
		return "Normal"
	case datatypes.RegionTypeWater:
		return "Water"
	case datatypes.RegionTypeLava:
		return "Lava"
	case datatypes.RegionTypePvp:
		return "Pvp"
	case datatypes.RegionTypeZoneline:
		return "Zoneline"
	case datatypes.RegionTypeWaterBlockLos:
		return "WaterBlockLos"
	case datatypes.RegionTypeFreezingWater:
		return "FreezingWater"
	case datatypes.RegionTypeSlippery:
		return "Slippery"
	case datatypes.RegionTypeUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// zonelineTypeToString converts a ZonelineType to its string representation.
func zonelineTypeToString(zt datatypes.ZonelineType) string {
	switch zt {
	case datatypes.ZonelineTypeReference:
		return "Reference"
	case datatypes.ZonelineTypeAbsolute:
		return "Absolute"
	default:
		return "Unknown"
	}
}
