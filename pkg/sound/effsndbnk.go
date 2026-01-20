package sound

import (
	"bufio"
	"os"
	"strings"
)

// EffSndBnk parses plain text files ending with "_sndbnk" which list the sounds in each zone.
// The indices of these sounds are used by the sound entries to determine which sound should play.
type EffSndBnk struct {
	emitSounds    []string
	loopSounds    []string
	soundFilePath string
}

// NewEffSndBnk creates a new EffSndBnk for the given file path.
func NewEffSndBnk(soundFilePath string) *EffSndBnk {
	return &EffSndBnk{
		soundFilePath: soundFilePath,
		emitSounds:    make([]string, 0),
		loopSounds:    make([]string, 0),
	}
}

// Initialize reads and parses the sound bank file.
// Returns an error if the file cannot be read.
func (e *EffSndBnk) Initialize() error {
	file, err := os.Open(e.soundFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File not existing is not an error
		}
		return err
	}
	defer file.Close()

	var currentList *[]string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Skip comment lines
		if strings.HasPrefix(line, "#") {
			continue
		}

		switch line {
		case SectionEmit:
			currentList = &e.emitSounds
		case SectionLoop:
			currentList = &e.loopSounds
		default:
			if currentList != nil {
				*currentList = append(*currentList, line)
			}
		}
	}

	return scanner.Err()
}

// GetEmitSound returns the emit sound at the given index.
// Returns Unknown if the index is out of bounds.
func (e *EffSndBnk) GetEmitSound(index int) string {
	return e.getValueFromList(index, e.emitSounds)
}

// GetLoopSound returns the loop sound at the given index.
// Returns Unknown if the index is out of bounds.
func (e *EffSndBnk) GetLoopSound(index int) string {
	return e.getValueFromList(index, e.loopSounds)
}

// getValueFromList returns the value at the given index from the list.
// Returns Unknown if the index is out of bounds.
func (e *EffSndBnk) getValueFromList(index int, list []string) string {
	if index < 0 || index >= len(list) {
		return Unknown
	}
	return list[index]
}

// EmitSoundCount returns the number of emit sounds.
func (e *EffSndBnk) EmitSoundCount() int {
	return len(e.emitSounds)
}

// LoopSoundCount returns the number of loop sounds.
func (e *EffSndBnk) LoopSoundCount() int {
	return len(e.loopSounds)
}
