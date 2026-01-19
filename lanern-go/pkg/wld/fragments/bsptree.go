package fragments

import (
	"fmt"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/datatypes"
)

// BspTree (0x21)
// Internal Name: None
// Binary tree with each leaf node containing a BspRegion fragment.
type BspTree struct {
	BaseFragment
	// Nodes contains the BSP nodes within the tree
	Nodes []*datatypes.BspNode
}

// FragmentType returns the fragment type ID.
func (f *BspTree) FragmentType() uint32 {
	return 0x21
}

// Initialize parses the fragment data.
func (f *BspTree) Initialize(index int, id int, size int, data []byte, fragments []Fragment, stringHash map[int]string, isNewFormat bool) error {
	f.initBase(index, size)
	reader := NewFragmentReader(data)

	nameRef, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read name reference: %w", err)
	}
	f.Name = GetStringFromHash(stringHash, nameRef)

	nodeCount, err := reader.ReadInt32()
	if err != nil {
		return fmt.Errorf("failed to read node count: %w", err)
	}

	f.Nodes = make([]*datatypes.BspNode, nodeCount)
	for i := int32(0); i < nodeCount; i++ {
		normalX, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read normal x: %w", err)
		}
		normalY, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read normal y: %w", err)
		}
		normalZ, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read normal z: %w", err)
		}
		splitDistance, err := reader.ReadFloat32()
		if err != nil {
			return fmt.Errorf("failed to read split distance: %w", err)
		}
		regionID, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read region id: %w", err)
		}
		leftNode, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read left node: %w", err)
		}
		rightNode, err := reader.ReadInt32()
		if err != nil {
			return fmt.Errorf("failed to read right node: %w", err)
		}

		f.Nodes[i] = &datatypes.BspNode{
			NormalX:       normalX,
			NormalY:       normalY,
			NormalZ:       normalZ,
			SplitDistance: splitDistance,
			RegionID:      int(regionID),
			LeftNode:      int(leftNode) - 1,
			RightNode:     int(rightNode) - 1,
		}
	}

	return nil
}

// LinkBspRegions links BSP nodes to their corresponding BSP Regions.
// The RegionId is not a fragment index but instead an index in a list of BSP Regions.
func (f *BspTree) LinkBspRegions(regions []*BspRegion) {
	for _, node := range f.Nodes {
		if node.RegionID == 0 {
			continue
		}
		if node.RegionID-1 >= 0 && node.RegionID-1 < len(regions) {
			node.Region = regions[node.RegionID-1]
		}
	}
}
