package sound

// clientSounds contains hardcoded client sounds.
// Verified that no other references exist in Trilogy client.
var clientSounds = map[int]string{
	39:  "death_me",
	143: "thunder1",
	144: "thunder2",
	158: "wind_lp1",
	159: "rainloop",
	160: "torch_lp",
	161: "watundlp",
}

// GetClientSound returns the sound name for the given client sound index.
// Returns Unknown if the index is not found.
func GetClientSound(index int) string {
	if name, ok := clientSounds[index]; ok {
		return name
	}
	return Unknown
}
