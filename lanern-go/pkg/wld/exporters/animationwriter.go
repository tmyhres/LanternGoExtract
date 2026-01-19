package exporters

import (
	"fmt"
	"strings"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/datatypes"
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/wld/fragments"
)

// AnimationWriter exports animation data to a text format.
type AnimationWriter struct {
	TextAssetWriter
	targetAnimation      string
	isCharacterAnimation bool
}

// NewAnimationWriter creates a new AnimationWriter.
func NewAnimationWriter(isCharacterAnimation bool) *AnimationWriter {
	w := &AnimationWriter{
		isCharacterAnimation: isCharacterAnimation,
	}
	w.AppendLine(ExportHeaderTitle + "Animation")
	return w
}

// SetTargetAnimation sets the target animation name to export.
func (w *AnimationWriter) SetTargetAnimation(animationName string) {
	w.targetAnimation = animationName
}

// AddFragmentData adds skeleton hierarchy animation data to the export buffer.
func (w *AnimationWriter) AddFragmentData(data fragments.Fragment) {
	skeleton, ok := data.(*fragments.SkeletonHierarchy)
	if !ok || skeleton == nil {
		return
	}

	if w.targetAnimation == "" {
		return
	}

	anim, exists := skeleton.Animations[w.targetAnimation]
	if !exists || anim == nil {
		return
	}

	w.AppendLine("# Animation: " + w.targetAnimation)
	w.AppendLine(fmt.Sprintf("framecount,%d", anim.FrameCount))
	w.AppendLine(fmt.Sprintf("totalTimeMs,%d", anim.AnimationTimeMs))

	for i := 0; i < len(skeleton.Skeleton); i++ {
		boneName := datatypes.CleanBoneAndStripBase(skeleton.BoneMapping[i], skeleton.ModelBase)
		fullPath := skeleton.Skeleton[i].CleanedFullPath

		trackArray := anim.TracksCleanedStripped
		var poseArray map[string]datatypes.TrackFragment
		if posAnim, exists := skeleton.Animations["pos"]; exists && posAnim != nil {
			poseArray = posAnim.TracksCleanedStripped
		}

		track, trackExists := trackArray[boneName]
		if !trackExists {
			// Fall back to pose animation
			if poseArray == nil {
				continue
			}
			poseTrack, poseExists := poseArray[boneName]
			if !poseExists {
				continue
			}

			trackDef := poseTrack.GetTrackDefFragment()
			if trackDef == nil {
				continue
			}

			// Get the first frame from pose track
			if trackDefFrag, ok := trackDef.(*fragments.TrackDefFragment); ok && len(trackDefFrag.Frames) > 0 {
				bt := trackDefFrag.Frames[0]
				w.createTrackString(fullPath, 0, bt, anim.AnimationTimeMs)
			}
		} else {
			trackDef := track.GetTrackDefFragment()
			if trackDef == nil {
				continue
			}

			trackDefFrag, ok := trackDef.(*fragments.TrackDefFragment)
			if !ok {
				continue
			}

			for j := 0; j < anim.FrameCount; j++ {
				if j >= len(trackDefFrag.Frames) {
					break
				}

				boneTransform := trackDefFrag.Frames[j]
				var delay int
				if w.isCharacterAnimation {
					if anim.FrameCount > 0 {
						delay = anim.AnimationTimeMs / anim.FrameCount
					}
				} else {
					delay = skeleton.Skeleton[i].Track.GetFrameMs()
				}
				w.createTrackString(fullPath, j, boneTransform, delay)
			}
		}
	}
}

// createTrackString creates a track string for a single bone transform.
func (w *AnimationWriter) createTrackString(fullPath string, frame int, boneTransform datatypes.BoneTransform, delay int) {
	w.AppendString(fullPath)
	w.AppendString(",")

	w.AppendString(fmt.Sprintf("%d", frame))
	w.AppendString(",")

	// Note: Coordinate system conversion (x, z, y) and rotation sign flipping
	w.AppendString(fmt.Sprintf("%g", boneTransform.Translation.X))
	w.AppendString(",")

	w.AppendString(fmt.Sprintf("%g", boneTransform.Translation.Z))
	w.AppendString(",")

	w.AppendString(fmt.Sprintf("%g", boneTransform.Translation.Y))
	w.AppendString(",")

	w.AppendString(fmt.Sprintf("%g", -boneTransform.Rotation.X))
	w.AppendString(",")

	w.AppendString(fmt.Sprintf("%g", -boneTransform.Rotation.Z))
	w.AppendString(",")

	w.AppendString(fmt.Sprintf("%g", -boneTransform.Rotation.Y))
	w.AppendString(",")

	w.AppendString(fmt.Sprintf("%g", boneTransform.Rotation.W))
	w.AppendString(",")

	w.AppendString(fmt.Sprintf("%g", boneTransform.Scale))
	w.AppendString(",")

	w.AppendLine(fmt.Sprintf("%d", delay))
}

// stripModelBase strips the model base from a bone name.
func stripModelBase(boneName, modelBase string) string {
	if strings.HasPrefix(boneName, modelBase+"/") {
		boneName = boneName[len(modelBase):]
		boneName = "root" + boneName
	}
	return boneName
}
