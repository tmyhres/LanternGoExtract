package sound

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// EntryLengthInBytes is the size of each entry in an EffSounds file.
const EntryLengthInBytes = 84

// ErrInvalidEffFile is returned when the .eff file format is invalid.
var ErrInvalidEffFile = errors.New("invalid .eff file - size must be multiple of 84")

// Logger is an interface for logging messages.
type Logger interface {
	LogError(message string)
	LogWarning(message string)
	LogInfo(message string)
}

// DefaultLogger is a no-op logger implementation.
type DefaultLogger struct{}

func (d *DefaultLogger) LogError(message string)   {}
func (d *DefaultLogger) LogWarning(message string) {}
func (d *DefaultLogger) LogInfo(message string)    {}

// EffSounds parses binary files ending with "_sounds.eff" extension.
// It contains no header, just an array of entries each consisting of 84 bytes.
// Each entry describes an instance of a sound or music in the world.
type EffSounds struct {
	soundBank      *EffSndBnk
	envAudio       *EnvAudio
	soundFilePath  string
	audioInstances []AudioInstance
}

// NewEffSounds creates a new EffSounds parser for the given file path.
func NewEffSounds(soundFilePath string, soundBank *EffSndBnk, envAudio *EnvAudio) *EffSounds {
	return &EffSounds{
		soundFilePath:  soundFilePath,
		soundBank:      soundBank,
		envAudio:       envAudio,
		audioInstances: make([]AudioInstance, 0),
	}
}

// Initialize reads and parses the sound file.
// Returns an error if the file cannot be read or is invalid.
func (e *EffSounds) Initialize(logger Logger) error {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	if e.soundBank == nil {
		return nil
	}

	file, err := os.Open(e.soundFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	fileLength := stat.Size()
	if fileLength%EntryLengthInBytes != 0 {
		logger.LogError(fmt.Sprintf("Invalid .eff file - size must be multiple of %d", EntryLengthInBytes))
		return ErrInvalidEffFile
	}

	entryCount := int(fileLength / EntryLengthInBytes)
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	for i := 0; i < entryCount; i++ {
		basePosition := EntryLengthInBytes * i

		// Read position and radius (offset 16)
		posX := readFloat32(data, basePosition+16)
		posY := readFloat32(data, basePosition+20)
		posZ := readFloat32(data, basePosition+24)
		radius := readFloat32(data, basePosition+28)

		// Read type byte (offset 56)
		typeByte := data[basePosition+56]

		audioType := AudioType(typeByte)
		if !audioType.IsValid() {
			logger.LogError(fmt.Sprintf("Unable to parse sound type: %d", typeByte))
			continue
		}

		// Read sound ID 1 (offset 48)
		soundID1 := readInt32(data, basePosition+48)
		sound1 := e.getSoundName(soundID1)

		switch audioType {
		case AudioTypeMusic:
			soundID2 := readInt32(data, basePosition+52)
			loopCountDay := readInt32(data, basePosition+60)
			loopCountNight := readInt32(data, basePosition+64)
			fadeOutMs := readInt32(data, basePosition+68)

			musicInstance := NewMusicInstance(
				audioType, posX, posY, posZ, radius,
				soundID1, soundID2,
				loopCountDay, loopCountNight, fadeOutMs,
			)
			e.audioInstances = append(e.audioInstances, musicInstance)

		case AudioTypeSound2D:
			soundID2 := readInt32(data, basePosition+52)
			sound2 := e.getSoundName(soundID2)
			cooldown1 := readInt32(data, basePosition+32)
			cooldown2 := readInt32(data, basePosition+36)
			cooldownRandom := readInt32(data, basePosition+40)
			volume1Raw := readInt32(data, basePosition+60)
			volume1 := e.getEalSoundVolume(sound1, volume1Raw)
			volume2Raw := readInt32(data, basePosition+64)
			volume2 := e.getEalSoundVolume(sound2, volume2Raw)

			soundInstance := NewSoundInstance2D(
				audioType, posX, posY, posZ, radius, volume1, sound1, cooldown1,
				sound2, cooldown2, cooldownRandom, volume2,
			)
			e.audioInstances = append(e.audioInstances, soundInstance)

		case AudioTypeSound3D:
			cooldown1 := readInt32(data, basePosition+32)
			cooldownRandom := readInt32(data, basePosition+40)
			volumeRaw := readInt32(data, basePosition+60)
			volume := e.getEalSoundVolume(sound1, volumeRaw)
			multiplier := readInt32(data, basePosition+72)

			soundInstance := NewSoundInstance3D(
				audioType, posX, posY, posZ, radius, volume, sound1,
				cooldown1, cooldownRandom, multiplier,
			)
			e.audioInstances = append(e.audioInstances, soundInstance)
		}
	}

	return nil
}

// getSoundName returns the sound name for the given sound ID.
func (e *EffSounds) getSoundName(soundID int32) string {
	emissionType := e.getEmissionType(soundID)

	switch emissionType {
	case EmissionTypeNone:
		return ""
	case EmissionTypeEmit:
		return e.soundBank.GetEmitSound(int(soundID - 1))
	case EmissionTypeLoop:
		return e.soundBank.GetLoopSound(int(soundID - 162))
	case EmissionTypeInternal:
		return GetClientSound(int(soundID))
	default:
		return Unknown
	}
}

// getEalSoundVolume returns the linear volume for the given sound.
func (e *EffSounds) getEalSoundVolume(soundName string, volumeRaw int32) float32 {
	if volumeRaw > 0 {
		volumeRaw = -volumeRaw
	}

	if e.envAudio == nil {
		return 1.0
	}

	if volumeRaw == 0 {
		return e.envAudio.GetVolumeLinear(soundName)
	}
	return e.envAudio.GetVolumeLinearFromLevel(volumeRaw)
}

// getEmissionType determines the emission type based on sound ID.
func (e *EffSounds) getEmissionType(soundID int32) EmissionType {
	if soundID <= 0 {
		return EmissionTypeNone
	}
	if soundID < 32 {
		return EmissionTypeEmit
	}
	if soundID < 162 {
		return EmissionTypeInternal
	}
	return EmissionTypeLoop
}

// AudioInstances returns the parsed audio instances.
func (e *EffSounds) AudioInstances() []AudioInstance {
	return e.audioInstances
}

// ExportSoundData exports the sound data to text files in the specified folder.
func (e *EffSounds) ExportSoundData(zoneName, rootFolder string) error {
	var sound2dExport strings.Builder
	var sound3dExport strings.Builder
	var musicExport strings.Builder

	for _, entry := range e.audioInstances {
		switch entry.Type() {
		case AudioTypeMusic:
			music, ok := entry.(*MusicInstance)
			if !ok {
				continue
			}
			fmt.Fprintf(&musicExport, "%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
				music.PosX, music.PosZ, music.PosY, music.PosRadius,
				music.TrackIndexDay, music.TrackIndexNight,
				music.LoopCountDay, music.LoopCountNight, music.FadeOutMs)

		case AudioTypeSound2D:
			sound2d, ok := entry.(*SoundInstance2D)
			if !ok {
				continue
			}
			fmt.Fprintf(&sound2dExport, "%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
				sound2d.PosX, sound2d.PosZ, sound2d.PosY, sound2d.PosRadius,
				sound2d.Sound1, sound2d.Sound2,
				sound2d.Cooldown1, sound2d.Cooldown2, sound2d.CooldownRandom,
				sound2d.Volume1, sound2d.Volume2)

		case AudioTypeSound3D:
			sound3d, ok := entry.(*SoundInstance3D)
			if !ok {
				continue
			}
			fmt.Fprintf(&sound3dExport, "%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
				sound3d.PosX, sound3d.PosZ, sound3d.PosY, sound3d.PosRadius,
				sound3d.Sound1, sound3d.Cooldown1, sound3d.CooldownRandom,
				sound3d.Volume1, sound3d.Multiplier)
		}
	}

	exportPath := filepath.Join(rootFolder, zoneName, "Zone")

	if musicExport.Len() > 0 {
		header := ExportHeaderTitle + "Music Instances\n" +
			"# Format: PosX, PosY, PosZ, Radius, MusicIndexDay, MusicIndexNight, LoopCountDay, LoopCountNight, FadeOutMs\n"

		if err := os.MkdirAll(exportPath, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(exportPath, "music_instances.txt"),
			[]byte(header+musicExport.String()), 0644); err != nil {
			return err
		}
	}

	if sound2dExport.Len() > 0 {
		header := ExportHeaderTitle + "Sound 2D Instances\n" +
			"# Format: PosX, PosY, PosZ, Radius, SoundNameDay, SoundNameNight, CooldownDay, CooldownNight, CooldownRandom, VolumeDay, VolumeNight\n"

		if err := os.MkdirAll(exportPath, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(exportPath, "sound2d_instances.txt"),
			[]byte(header+sound2dExport.String()), 0644); err != nil {
			return err
		}
	}

	if sound3dExport.Len() > 0 {
		header := ExportHeaderTitle + "Sound 3D Instances\n" +
			"# Format: PosX, PosY, PosZ, Radius, SoundName, Cooldown, CooldownRandom, Volume, Multiplier\n"

		if err := os.MkdirAll(exportPath, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(exportPath, "sound3d_instances.txt"),
			[]byte(header+sound3dExport.String()), 0644); err != nil {
			return err
		}
	}

	return nil
}

// readFloat32 reads a little-endian float32 from the data at the given offset.
func readFloat32(data []byte, offset int) float32 {
	bits := binary.LittleEndian.Uint32(data[offset:])
	return math.Float32frombits(bits)
}

// readInt32 reads a little-endian int32 from the data at the given offset.
func readInt32(data []byte, offset int) int32 {
	return int32(binary.LittleEndian.Uint32(data[offset:]))
}
