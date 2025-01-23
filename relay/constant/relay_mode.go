package constant

import (
	"net/http"
	"strings"
)

const (
	RelayModeUnknown = iota
	RelayModeChatCompletions
	RelayModeCompletions
	RelayModeEmbeddings
	RelayModeModerations
	RelayModeImagesGenerations
	RelayModeEdits

	RelayModeMidjourneyImagine
	RelayModeMidjourneyDescribe
	RelayModeMidjourneyBlend
	RelayModeMidjourneyChange
	RelayModeMidjourneySimpleChange
	RelayModeMidjourneyNotify
	RelayModeMidjourneyTaskFetch
	RelayModeMidjourneyTaskImageSeed
	RelayModeMidjourneyTaskFetchByCondition
	RelayModeMidjourneyAction
	RelayModeMidjourneyModal
	RelayModeMidjourneyShorten
	RelayModeSwapFace
	RelayModeMidjourneyUpload

	RelayModeAudioSpeech        // tts
	RelayModeAudioTranscription // whisper
	RelayModeAudioTranslation   // whisper

	RelayModeSunoFetch
	RelayModeSunoFetchByID
	RelayModeSunoSubmit
	RelayModeSunoGenerate
	RelayModeSunoGenerateLyrics
	RelayModeSunoFeed
	RelayModeSunoFeedLyrics

	RelayModeRerank

	RelayModeRealtime

	RelayModeLumaGenerations
	RelayModeLumaExtend
	RelayModeLumaTasks
	RelayModeLumaSingleTask
	RelayModeLumaFetchHdUrl

	RelayModeRunwayGenerations
	RelayModeRunwayTasks
	RelayModeRunwayUpload

	RelayModeKlingImages
	RelayModeKlingVirtual
	RelayModeKlingVideo
	RelayModeKlingTasks
)

func Path2RelayMode(path string) int {
	relayMode := RelayModeUnknown
	if strings.HasPrefix(path, "/v1/chat/completions") || strings.HasPrefix(path, "/pg/chat/completions") {
		relayMode = RelayModeChatCompletions
	} else if strings.HasPrefix(path, "/v1/completions") {
		relayMode = RelayModeCompletions
	} else if strings.HasPrefix(path, "/v1/embeddings") {
		relayMode = RelayModeEmbeddings
	} else if strings.HasSuffix(path, "embeddings") {
		relayMode = RelayModeEmbeddings
	} else if strings.HasPrefix(path, "/v1/moderations") {
		relayMode = RelayModeModerations
	} else if strings.HasPrefix(path, "/v1/images/generations") {
		relayMode = RelayModeImagesGenerations
	} else if strings.HasPrefix(path, "/v1/edits") {
		relayMode = RelayModeEdits
	} else if strings.HasPrefix(path, "/v1/audio/speech") {
		relayMode = RelayModeAudioSpeech
	} else if strings.HasPrefix(path, "/v1/audio/transcriptions") {
		relayMode = RelayModeAudioTranscription
	} else if strings.HasPrefix(path, "/v1/audio/translations") {
		relayMode = RelayModeAudioTranslation
	} else if strings.HasPrefix(path, "/v1/rerank") {
		relayMode = RelayModeRerank
	} else if strings.HasPrefix(path, "/v1/realtime") {
		relayMode = RelayModeRealtime
	}
	return relayMode
}

func Path2RelayModeMidjourney(path string) int {
	relayMode := RelayModeUnknown
	if strings.HasSuffix(path, "/mj/submit/action") {
		// midjourney plus
		relayMode = RelayModeMidjourneyAction
	} else if strings.HasSuffix(path, "/mj/submit/modal") {
		// midjourney plus
		relayMode = RelayModeMidjourneyModal
	} else if strings.HasSuffix(path, "/mj/submit/shorten") {
		// midjourney plus
		relayMode = RelayModeMidjourneyShorten
	} else if strings.HasSuffix(path, "/mj/insight-face/swap") {
		// midjourney plus
		relayMode = RelayModeSwapFace
	} else if strings.HasSuffix(path, "/submit/upload-discord-images") {
		// midjourney plus
		relayMode = RelayModeMidjourneyUpload
	} else if strings.HasSuffix(path, "/mj/submit/imagine") {
		relayMode = RelayModeMidjourneyImagine
	} else if strings.HasSuffix(path, "/mj/submit/blend") {
		relayMode = RelayModeMidjourneyBlend
	} else if strings.HasSuffix(path, "/mj/submit/describe") {
		relayMode = RelayModeMidjourneyDescribe
	} else if strings.HasSuffix(path, "/mj/notify") {
		relayMode = RelayModeMidjourneyNotify
	} else if strings.HasSuffix(path, "/mj/submit/change") {
		relayMode = RelayModeMidjourneyChange
	} else if strings.HasSuffix(path, "/mj/submit/simple-change") {
		relayMode = RelayModeMidjourneyChange
	} else if strings.HasSuffix(path, "/fetch") {
		relayMode = RelayModeMidjourneyTaskFetch
	} else if strings.HasSuffix(path, "/image-seed") {
		relayMode = RelayModeMidjourneyTaskImageSeed
	} else if strings.HasSuffix(path, "/list-by-condition") {
		relayMode = RelayModeMidjourneyTaskFetchByCondition
	}
	return relayMode
}

func Path2RelaySuno(method, path string) int {
	relayMode := RelayModeUnknown
	if method == http.MethodPost && strings.HasSuffix(path, "/fetch") {
		relayMode = RelayModeSunoFetch
	} else if method == http.MethodGet && strings.Contains(path, "/fetch/") {
		relayMode = RelayModeSunoFetchByID
	} else if strings.Contains(path, "/submit/") {
		relayMode = RelayModeSunoSubmit
	} else if method == http.MethodPost && strings.HasSuffix(path, "/generate") {
		relayMode = RelayModeSunoGenerate
	} else if method == http.MethodPost && strings.Contains(path, "/lyrics") {
		relayMode = RelayModeSunoGenerateLyrics
	} else if method == http.MethodGet && strings.Contains(path, "/feed") {
		relayMode = RelayModeSunoFeed
	} else if method == http.MethodGet && strings.Contains(path, "/lyrics") {
		relayMode = RelayModeSunoFeedLyrics
	}
	return relayMode
}

func Path2RelayLuma(method, path string) int {
	relayMode := RelayModeUnknown
	if method == http.MethodPost && strings.HasSuffix(path, "/generations") {
		relayMode = RelayModeLumaGenerations
	} else if method == http.MethodPost && strings.HasSuffix(path, "/extend") {
		relayMode = RelayModeLumaExtend
	} else if method == http.MethodPost && strings.HasSuffix(path, "/tasks") {
		relayMode = RelayModeLumaTasks
	} else if method == http.MethodGet && strings.Contains(path, "/generations") {
		relayMode = RelayModeLumaSingleTask
	} else if method == http.MethodGet && strings.HasSuffix(path, "/download_video_url") {
		relayMode = RelayModeLumaFetchHdUrl
	}
	return relayMode
}

func Path2RelayRunway(method, path string) int {
	relayMode := RelayModeUnknown
	if method == http.MethodPost && (strings.HasSuffix(path, "/tasks") || strings.HasSuffix(path, "/image_to_video")) {
		relayMode = RelayModeRunwayGenerations
	} else if method == http.MethodGet && strings.Contains(path, "/tasks") {
		relayMode = RelayModeRunwayTasks
	} else if method == http.MethodPost && strings.Contains(path, "/uploads") {
		relayMode = RelayModeRunwayUpload
	}
	return relayMode
}

func Path2RelayKling(method, path string) int {
	relayMode := RelayModeUnknown
	if method == http.MethodPost && strings.HasSuffix(path, "/generations") {
		relayMode = RelayModeKlingImages
	} else if method == http.MethodPost && strings.HasSuffix(path, "/kolors-virtual-try-on") {
		relayMode = RelayModeKlingVirtual
	} else if method == http.MethodPost && strings.Contains(path, "/videos") {
		relayMode = RelayModeKlingVideo
	} else if method == http.MethodGet && (strings.Contains(path, "/videos") || strings.Contains(path, "/images")) {
		relayMode = RelayModeKlingTasks
	}
	return relayMode
}
