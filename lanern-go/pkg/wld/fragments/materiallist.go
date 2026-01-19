package fragments

import (
	"strconv"
	"strings"
)

// MaterialList (0x31) contains a list of material fragments (0x30).
// Internal name: _MP
// This list is used in the rendering of a mesh via the list indices.
type MaterialList struct {
	BaseFragment

	// Materials is the list of materials.
	Materials []*Material

	// Slots is a mapping of slot names to alternate skins.
	Slots map[string]map[int]*Material

	// VariantCount is the number of alternate skins.
	VariantCount int

	// AdditionalMaterials holds extra materials added as variants.
	AdditionalMaterials []*Material

	// HasBeenExported prevents the material list from being exported multiple times.
	HasBeenExported bool
}

// FragmentType returns the fragment type ID (0x31).
func (m *MaterialList) FragmentType() uint32 {
	return 0x31
}

// Initialize parses the material list fragment data.
func (m *MaterialList) Initialize(index int, id int, size int, data []byte, fragments []Fragment,
	stringHash map[int]string, isNewFormat bool) error {
	m.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return err
	}
	m.Name = GetStringFromHash(stringHash, nameRef)

	// Read flags
	_, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read material count
	materialCount, err := r.ReadInt32()
	if err != nil {
		return err
	}

	m.Materials = make([]*Material, 0, materialCount)

	for i := int32(0); i < materialCount; i++ {
		reference, err := r.ReadInt32()
		if err != nil {
			return err
		}

		fragIdx := int(reference) - 1
		if fragIdx >= 0 && fragIdx < len(fragments) {
			if material, ok := fragments[fragIdx].(*Material); ok {
				m.Materials = append(m.Materials, material)
				// Materials that are referenced in the MaterialList are already handled
				material.IsHandled = true
			}
		}
	}

	return nil
}

// BuildSlotMapping builds the slot mapping for character skin variants.
func (m *MaterialList) BuildSlotMapping() {
	m.Slots = make(map[string]map[int]*Material)

	if len(m.Materials) == 0 {
		return
	}

	for _, material := range m.Materials {
		character, _, partName := parseCharacterSkin(cleanFragmentName(material.GetName()))
		if character == "" {
			continue
		}

		key := character + "_" + partName
		m.Slots[key] = make(map[int]*Material)
	}

	m.AdditionalMaterials = make([]*Material, 0)
}

// AddVariant adds a material variant for character skins.
func (m *MaterialList) AddVariant(material *Material) {
	character, skinID, partName := parseCharacterSkin(cleanFragmentName(material.GetName()))

	key := character + "_" + partName

	if _, exists := m.Slots[key]; !exists {
		m.Slots[key] = make(map[int]*Material)
	}

	skinIDNumber, err := strconv.Atoi(skinID)
	if err != nil {
		skinIDNumber = 0
	}

	m.Slots[key][skinIDNumber] = material
	material.IsHandled = true

	if skinIDNumber > m.VariantCount {
		m.VariantCount = skinIDNumber
	}

	m.AdditionalMaterials = append(m.AdditionalMaterials, material)
}

// GetMaterialVariants returns all material variants for a given material.
func (m *MaterialList) GetMaterialVariants(material *Material) []*Material {
	var additionalSkins []*Material

	if m.Slots == nil {
		return additionalSkins
	}

	character, _, partName := parseCharacterSkin(cleanFragmentName(material.GetName()))

	key := character + "_" + partName

	variants, exists := m.Slots[key]
	if !exists {
		return additionalSkins
	}

	for i := 0; i < m.VariantCount; i++ {
		if mat, exists := variants[i+1]; exists {
			additionalSkins = append(additionalSkins, mat)
		} else {
			additionalSkins = append(additionalSkins, nil)
		}
	}

	return additionalSkins
}

// GetMaterialPrefix returns the material prefix for a shader type.
func GetMaterialPrefix(shaderType ShaderType) string {
	switch shaderType {
	case ShaderTypeDiffuse:
		return "d_"
	case ShaderTypeInvisible:
		return "i_"
	case ShaderTypeBoundary:
		return "b_"
	case ShaderTypeTransparent25:
		return "t25_"
	case ShaderTypeTransparent50:
		return "t50_"
	case ShaderTypeTransparent75:
		return "t75_"
	case ShaderTypeTransparentAdditive:
		return "ta_"
	case ShaderTypeTransparentAdditiveUnlit:
		return "tau_"
	case ShaderTypeTransparentMasked:
		return "tm_"
	case ShaderTypeDiffuseSkydome:
		return "ds_"
	case ShaderTypeTransparentSkydome:
		return "ts_"
	case ShaderTypeTransparentAdditiveUnlitSkydome:
		return "taus_"
	default:
		return "d_"
	}
}

// parseCharacterSkin parses character skin info from a material name.
// Returns character, skinID, and partName.
func parseCharacterSkin(materialName string) (character, skinID, partName string) {
	if len(materialName) != 9 {
		return "", "", ""
	}

	character = materialName[0:3]
	skinID = materialName[5:7]
	partName = materialName[3:5] + materialName[7:9]
	return
}

// cleanFragmentName removes common prefixes/suffixes from fragment names.
func cleanFragmentName(name string) string {
	// Remove common prefixes
	name = strings.TrimPrefix(name, "_")

	// Remove _MDF suffix if present
	if strings.HasSuffix(name, "_MDF") {
		name = name[:len(name)-4]
	}

	return name
}
