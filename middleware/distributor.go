package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	relayconstant "one-api/relay/constant"
	"one-api/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ModelRequest struct {
	Model string `json:"model"`
}

type SunoModelRequest struct {
	Model string `json:"mv"`
}

type RunwayOptions struct {
	Seconds     int  `json:"seconds"`
	ExploreMode bool `json:"exploreMode"`
}

type RunwayModelRequest struct {
	Model   string        `json:"taskType"`
	Options RunwayOptions `json:"options"`
}

type RunwaymlModelRequest struct {
	Model    string `json:"model"`
	Duration int    `json:"duration"`
}

type KlingModelRequest struct {
	Model     string `json:"model"`
	ModelName string `json:"model_name"`
	Mode      string `json:"mode"`
	Duration  int    `json:"duration"`
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		allowIpsMap := c.GetStringMap("allow_ips")
		if len(allowIpsMap) != 0 {
			clientIp := c.ClientIP()
			if _, ok := allowIpsMap[clientIp]; !ok {
				abortWithOpenAiMessage(c, http.StatusForbidden, "您的 IP 不在令牌允许访问的列表中")
				return
			}
		}
		userId := c.GetInt("id")
		var channel *model.Channel
		channelId, ok := c.Get("specific_channel_id")
		modelRequest, shouldSelectChannel, err := getModelRequest(c)
		if err != nil {
			abortWithOpenAiMessage(c, http.StatusBadRequest, "Invalid request, "+err.Error())
			return
		}
		userGroup, _ := model.CacheGetUserGroup(userId)
		tokenGroup := c.GetString("token_group")
		if tokenGroup != "" {
			// check common.UserUsableGroups[userGroup]
			if _, ok := common.GetUserUsableGroups(userGroup)[tokenGroup]; !ok {
				abortWithOpenAiMessage(c, http.StatusForbidden, fmt.Sprintf("令牌分组 %s 已被禁用", tokenGroup))
				return
			}
			// check group in common.GroupRatio
			if _, ok := common.GroupRatio[tokenGroup]; !ok {
				abortWithOpenAiMessage(c, http.StatusForbidden, fmt.Sprintf("分组 %s 已被弃用", tokenGroup))
				return
			}
			userGroup = tokenGroup
		}
		c.Set("group", userGroup)
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithOpenAiMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithOpenAiMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != common.ChannelStatusEnabled {
				abortWithOpenAiMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
		} else {
			// Select a channel for the user
			// check token model mapping
			modelLimitEnable := c.GetBool("token_model_limit_enabled")
			if modelLimitEnable {
				s, ok := c.Get("token_model_limit")
				var tokenModelLimit map[string]bool
				if ok {
					tokenModelLimit = s.(map[string]bool)
				} else {
					tokenModelLimit = map[string]bool{}
				}
				if tokenModelLimit != nil {
					if _, ok := tokenModelLimit[modelRequest.Model]; !ok {
						abortWithOpenAiMessage(c, http.StatusForbidden, "该令牌无权访问模型 "+modelRequest.Model)
						return
					}
				} else {
					// token model limit is empty, all models are not allowed
					abortWithOpenAiMessage(c, http.StatusForbidden, "该令牌无权访问任何模型")
					return
				}
			}

			if shouldSelectChannel {
				channel, err = model.CacheGetRandomSatisfiedChannel(userGroup, modelRequest.Model, 0)
				if err != nil {
					message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, modelRequest.Model)
					// 如果错误，但是渠道不为空，说明是数据库一致性问题
					if channel != nil {
						common.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
						message = "数据库一致性已被破坏，请联系管理员"
					}
					// 如果错误，而且渠道为空，说明是没有可用渠道
					abortWithOpenAiMessage(c, http.StatusServiceUnavailable, message)
					return
				}
				if channel == nil {
					abortWithOpenAiMessage(c, http.StatusServiceUnavailable, fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道（数据库一致性已被破坏）", userGroup, modelRequest.Model))
					return
				}
			}
		}
		SetupContextForSelectedChannel(c, channel, modelRequest.Model)
		c.Next()
	}
}

func getModelRequest(c *gin.Context) (*ModelRequest, bool, error) {
	var modelRequest ModelRequest
	shouldSelectChannel := true
	var err error
	if strings.Contains(c.Request.URL.Path, "/mj/") {
		relayMode := relayconstant.Path2RelayModeMidjourney(c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeMidjourneyTaskFetch ||
			relayMode == relayconstant.RelayModeMidjourneyTaskFetchByCondition ||
			relayMode == relayconstant.RelayModeMidjourneyNotify ||
			relayMode == relayconstant.RelayModeMidjourneyTaskImageSeed {
			shouldSelectChannel = false
		} else {
			midjourneyRequest := dto.MidjourneyRequest{}
			err = common.UnmarshalBodyReusable(c, &midjourneyRequest)
			if err != nil {
				abortWithMidjourneyMessage(c, http.StatusBadRequest, constant.MjErrorUnknown, "无效的请求, "+err.Error())
				return nil, false, err
			}
			midjourneyModel, mjErr, success := service.GetMjRequestModel(relayMode, &midjourneyRequest)
			if mjErr != nil {
				abortWithMidjourneyMessage(c, http.StatusBadRequest, mjErr.Code, mjErr.Description)
				return nil, false, fmt.Errorf(mjErr.Description)
			}
			if midjourneyModel == "" {
				if !success {
					abortWithMidjourneyMessage(c, http.StatusBadRequest, constant.MjErrorUnknown, "无效的请求, 无法解析模型")
					return nil, false, fmt.Errorf("无效的请求, 无法解析模型")
				} else {
					// task fetch, task fetch by condition, notify
					shouldSelectChannel = false
				}
			}
			modelRequest.Model = midjourneyModel
		}
		c.Set("relay_mode", relayMode)
	} else if strings.Contains(c.Request.URL.Path, "/suno/") {
		relayMode := relayconstant.Path2RelaySuno(c.Request.Method, c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeSunoFetch ||
			relayMode == relayconstant.RelayModeSunoFetchByID ||
			relayMode == relayconstant.RelayModeSunoFeed ||
			relayMode == relayconstant.RelayModeSunoFeedLyrics {
			shouldSelectChannel = false
		} else if relayMode == relayconstant.RelayModeSunoSubmit {
			modelName := service.CoverTaskActionToModelName(constant.TaskPlatformSuno, c.Param("action"))
			modelRequest.Model = modelName
		} else if relayMode == relayconstant.RelayModeSunoGenerateLyrics {
			modelRequest.Model = "chirp-lyrics"
		} else {
			var sunoModelRequest SunoModelRequest
			err = common.UnmarshalBodyReusable(c, &sunoModelRequest)
			if err != nil {
				modelRequest.Model = "chirp-v3-5"
			} else {
				modelRequest.Model = sunoModelRequest.Model
			}
		}
		c.Set("platform", string(constant.TaskPlatformSuno))
		c.Set("relay_mode", relayMode)
	} else if strings.Contains(c.Request.URL.Path, "/luma/") {
		relayMode := relayconstant.Path2RelayLuma(c.Request.Method, c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeLumaTasks ||
			relayMode == relayconstant.RelayModeLumaSingleTask {
			modelRequest.Model = "luma_video_fetch_tasks"
		} else if relayMode == relayconstant.RelayModeLumaFetchHdUrl {
			modelRequest.Model = "luma_video_download_api"
		} else {
			modelRequest.Model = "luma_video_api"
		}
		c.Set("platform", string(constant.TaskPlatformLuma))
		c.Set("relay_mode", relayMode)
	} else if strings.Contains(c.Request.URL.Path, "/runway/") {
		relayMode := relayconstant.Path2RelayRunway(c.Request.Method, c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeRunwayTasks {
			modelRequest.Model = "runway_fetch_tasks"
		} else if relayMode == relayconstant.RelayModeRunwayUpload {
			modelRequest.Model = "runway_upload"
		} else {
			var runwayModelRequest RunwayModelRequest
			err = common.UnmarshalBodyReusable(c, &runwayModelRequest)
			if err != nil {
				modelRequest.Model = "runway_gen3_fast_10"
			} else {
				if runwayModelRequest.Model == "gen2" {
					modelRequest.Model = "runway_gen2"
				} else {
					var seconds = 10
					if runwayModelRequest.Options.Seconds <= 5 {
						seconds = 5
					}
					if runwayModelRequest.Model == "europa" && runwayModelRequest.Options.ExploreMode {
						modelRequest.Model = fmt.Sprintf("%s%d", "runway_gen3_fast_", seconds)
					} else if runwayModelRequest.Model == "europa" && !runwayModelRequest.Options.ExploreMode {
						modelRequest.Model = fmt.Sprintf("%s%d", "runway_gen3_", seconds)
					} else {
						modelRequest.Model = fmt.Sprintf("%s%d", "runway_gen3_turbo_", seconds)
					}
				}
			}
		}
		c.Set("platform", string(constant.TaskPlatformRunway))
		c.Set("relay_mode", relayMode)
	} else if strings.Contains(c.Request.URL.Path, "/runwayml/") {
		relayMode := relayconstant.Path2RelayRunway(c.Request.Method, c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeRunwayTasks {
			modelRequest.Model = "runway_fetch_tasks"
		} else {
			var runwaymlModelRequest RunwaymlModelRequest
			err = common.UnmarshalBodyReusable(c, &runwaymlModelRequest)
			if err != nil {
				modelRequest.Model = "runwayml_gen3a_turbo_10"
			} else {
				modelRequest.Model = fmt.Sprintf("%s_%s_%d", "runwayml", runwaymlModelRequest.Model, runwaymlModelRequest.Duration)
			}
		}
		c.Set("platform", string(constant.TaskPlatformRunway))
		c.Set("relay_mode", relayMode)
	} else if strings.Contains(c.Request.URL.Path, "/kling/") {
		relayMode := relayconstant.Path2RelayKling(c.Request.Method, c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeKlingTasks {
			modelRequest.Model = "kling_fetch_tasks"
		} else {
			var klingModelRequest KlingModelRequest
			err = common.UnmarshalBodyReusable(c, &klingModelRequest)
			if err != nil {
				modelRequest.Model = "kling_fetch_tasks"
			} else {
				if relayMode == relayconstant.RelayModeKlingImages {
					modelRequest.Model = klingModelRequest.Model + "_images"
				} else if relayMode == relayconstant.RelayModeKlingVirtual {
					modelRequest.Model = klingModelRequest.ModelName
				} else {
					modelRequest.Model = fmt.Sprintf("%s_videos_%s_%d", klingModelRequest.Model, klingModelRequest.Mode, klingModelRequest.Duration)
				}
			}
		}
		c.Set("platform", constant.TaskPlatformKling)
		c.Set("relay_mode", relayMode)
	} else if !strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") {
		err = common.UnmarshalBodyReusable(c, &modelRequest)
	}
	if err != nil {
		abortWithOpenAiMessage(c, http.StatusBadRequest, "无效的请求, "+err.Error())
		return nil, false, errors.New("无效的请求, " + err.Error())
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/realtime") {
		//wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-10-01
		modelRequest.Model = c.Query("model")
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "text-moderation-stable"
		}
	}
	if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
		if modelRequest.Model == "" {
			modelRequest.Model = c.Param("model")
		}
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
		_ = common.UnmarshalBodyReusable(c, &modelRequest)
		modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "dall-e-3")
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/audio") {
		relayMode := relayconstant.RelayModeAudioSpeech
		if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/speech") {
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "tts-1")
		} else if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations") {
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, c.PostForm("model"))
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "whisper-1")
			relayMode = relayconstant.RelayModeAudioTranslation
		} else if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") {
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, c.PostForm("model"))
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "whisper-1")
			relayMode = relayconstant.RelayModeAudioTranscription
		}
		c.Set("relay_mode", relayMode)
	}
	return &modelRequest, shouldSelectChannel, nil
}

func SetupContextForSelectedChannel(c *gin.Context, channel *model.Channel, modelName string) {
	c.Set("original_model", modelName) // for retry
	if channel == nil {
		return
	}
	c.Set("channel_id", channel.Id)
	c.Set("channel_name", channel.Name)
	c.Set("channel_type", channel.Type)
	if nil != channel.OpenAIOrganization && "" != *channel.OpenAIOrganization {
		c.Set("channel_organization", *channel.OpenAIOrganization)
	}
	c.Set("auto_ban", channel.GetAutoBan())
	c.Set("model_mapping", channel.GetModelMapping())
	c.Set("status_code_mapping", channel.GetStatusCodeMapping())
	c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
	c.Set("base_url", channel.GetBaseURL())
	// TODO: api_version统一
	switch channel.Type {
	case common.ChannelTypeAzure:
		c.Set("api_version", channel.Other)
	case common.ChannelTypeVertexAi:
		c.Set("region", channel.Other)
	case common.ChannelTypeXunfei:
		c.Set("api_version", channel.Other)
	case common.ChannelTypeGemini:
		c.Set("api_version", channel.Other)
	case common.ChannelTypeAli:
		c.Set("plugin", channel.Other)
	case common.ChannelCloudflare:
		c.Set("api_version", channel.Other)
	}
}
