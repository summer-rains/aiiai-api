package constant

type TaskPlatform string

const (
	TaskPlatformSuno       TaskPlatform = "suno"
	TaskPlatformMidjourney              = "mj"
	TaskPlatformLuma                    = "luma"
	TaskPlatformRunway                  = "runway"
	TaskPlatformKling                   = "runway"
)

const (
	SunoActionMusic  = "MUSIC"
	SunoActionLyrics = "LYRICS"
)

var SunoModel2Action = map[string]string{
	"suno_music":  SunoActionMusic,
	"suno_lyrics": SunoActionLyrics,
}
