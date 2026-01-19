package fragments

// ParticleCloud (0x34) defines a particle system.
// Internal name: None
// Can be referenced from a skeleton bone.
type ParticleCloud struct {
	BaseFragment

	// ParticleSprite is the reference to the particle sprite fragment.
	ParticleSprite Fragment

	// Flags - always 4
	Flags int32

	// ParticleCount - values like 200, 30
	ParticleCount int32

	// Various particle parameters (partially decoded)
	Value08 int32 // always 3
	Value12 int32 // Values are 1, 3, or 4
	Value16 byte
	Value17 byte
	Value18 byte
	Value19 byte
	Value44 float32
	Value48 float32
	Value52 int32   // looks like an int. numbers like 1000, 100, 750, 500, 1600, 2500
	Value56 float32 // low numbers: 4, 5, 8, 10, 0
	Value60 float32 // 0 or 1
	Value64 float32 // 0 or -1
	Value68 float32 // 0 or -1
	Value72 int32   // probably int: 13, 15, 20, 600, 83
	Value76 float32 // confirmed float: 0.4, 0.5, 1.5, 0.1
	Value80 float32 // float: 0.4, 1.9
}

// FragmentType returns the fragment type ID (0x34).
func (p *ParticleCloud) FragmentType() uint32 {
	return 0x34
}

// Initialize parses the particle cloud fragment data.
func (p *ParticleCloud) Initialize(index int, id int, size int, data []byte, fragments []Fragment,
	stringHash map[int]string, isNewFormat bool) error {
	p.initBase(index, size)

	r := NewFragmentReader(data)

	// Read name reference
	nameRef, err := r.ReadInt32()
	if err != nil {
		return err
	}
	p.Name = GetStringFromHash(stringHash, nameRef)

	// Read flags - always 4
	p.Flags, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read value08 - always 3
	p.Value08, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read value12 - Values are 1, 3, or 4
	p.Value12, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read bytes 16-19
	p.Value16, err = r.ReadByte()
	if err != nil {
		return err
	}
	p.Value17, err = r.ReadByte()
	if err != nil {
		return err
	}
	p.Value18, err = r.ReadByte()
	if err != nil {
		return err
	}
	p.Value19, err = r.ReadByte()
	if err != nil {
		return err
	}

	// Read value20 - particle count (200, 30, etc)
	p.ParticleCount, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Skip values 24-40 (5 int32s, always 0)
	for i := 0; i < 5; i++ {
		_, err = r.ReadInt32()
		if err != nil {
			return err
		}
	}

	// Read value44 - confirmed float
	p.Value44, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read value48 - looks like a float
	p.Value48, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read value52 - looks like an int
	p.Value52, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read value56 - looks like a float, low numbers
	p.Value56, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read value60 - float 0 or 1
	p.Value60, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read value64 - float 0 or -1
	p.Value64, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read value68 - float 0 or -1
	p.Value68, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read value72 - probably int
	p.Value72, err = r.ReadInt32()
	if err != nil {
		return err
	}

	// Read value76 - confirmed float
	p.Value76, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read value80 - float
	p.Value80, err = r.ReadFloat32()
	if err != nil {
		return err
	}

	// Read particle sprite reference
	particleSpriteRef, err := r.ReadInt32()
	if err != nil {
		return err
	}

	fragIdx := int(particleSpriteRef) - 1
	if fragIdx >= 0 && fragIdx < len(fragments) {
		p.ParticleSprite = fragments[fragIdx]
	}

	return nil
}
