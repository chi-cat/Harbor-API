package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"one-api/common"
	"strings"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Ability struct {
	Group      string  `json:"group" gorm:"type:varchar(64);primaryKey;autoIncrement:false"`
	Model      string  `json:"model" gorm:"type:varchar(64);primaryKey;autoIncrement:false"`
	ChannelId  int     `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	Enabled    bool    `json:"enabled"`
	Priority   *int64  `json:"priority" gorm:"bigint;default:0;index"`
	Weight     uint    `json:"weight" gorm:"default:0;index"`
	Tag        *string `json:"tag" gorm:"index"`
	ModelAlias *string `json:"model_alias" gorm:"type:varchar(64);index"`
}

func GetGroupModels(group string) []string {
	var models []string
	// Find distinct models
	groupCol := "`group`"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
	}
	DB.Table("abilities").Where(groupCol+" = ? and enabled = ?", group, true).Distinct("model").Pluck("model", &models)
	return models
}

func GetEnabledModels() []string {
	var models []string
	// Find distinct models
	DB.Table("abilities").Where("enabled = ?", true).Distinct("model").Pluck("model", &models)
	return models
}

func GetAllEnableAbilities() []Ability {
	var abilities []Ability
	DB.Find(&abilities, "enabled = ?", true)
	return abilities
}

func getPriority(group string, model string, retry int) (int, error) {
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}

	var priorities []int
	err := DB.Model(&Ability{}).
		Select("DISTINCT(priority)").
		Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, model).
		Order("priority DESC").              // 按优先级降序排序
		Pluck("priority", &priorities).Error // Pluck用于将查询的结果直接扫描到一个切片中

	if err != nil {
		// 处理错误
		return 0, err
	}

	if len(priorities) == 0 {
		// 如果没有查询到优先级，则返回错误
		return 0, errors.New("数据库一致性被破坏")
	}

	// 确定要使用的优先级
	var priorityToUse int
	if retry >= len(priorities) {
		// 如果重试次数大于优先级数，则使用最小的优先级
		priorityToUse = priorities[len(priorities)-1]
	} else {
		priorityToUse = priorities[retry]
	}
	return priorityToUse, nil
}

func getChannelQuery(group string, model string, retry int) *gorm.DB {
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}
	maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? and (model = ? or model_alias = ?) and enabled = "+trueVal, group, model, model)
	channelQuery := DB.Where(groupCol+" = ? and (model = ? or model_alias = ?)and enabled = "+trueVal+" and priority = (?)", group, model, model, maxPrioritySubQuery)
	if retry != 0 {
		priority, err := getPriority(group, model, retry)
		if err != nil {
			common.SysError(fmt.Sprintf("Get priority failed: %s", err.Error()))
		} else {
			channelQuery = DB.Where(groupCol+" = ? and ( model = ? or model_alias = ?) and enabled = "+trueVal+" and priority = ?", group, model, model, priority)
		}
	}

	return channelQuery
}

func GetRandomSatisfiedChannel(group string, model string, limitsMap map[string]bool, retry int) (*Channel, error) {
	var abilities []Ability

	var err error = nil
	channelQuery := getChannelQuery(group, model, retry)
	if common.UsingSQLite || common.UsingPostgreSQL {
		err = channelQuery.Order("weight DESC").Find(&abilities).Error
	} else {
		err = channelQuery.Order("weight DESC").Find(&abilities).Error
	}
	if err != nil {
		return nil, err
	}
	copyAbilities := make([]Ability, 0)
	for _, ability := range abilities {
		if limitsMap == nil || limitsMap[ability.Model] {
			copyAbilities = append(copyAbilities, ability)
		}
	}
	abilities = copyAbilities
	channel := Channel{}
	if len(abilities) > 0 {
		// 平滑系数
		smoothingFactor := 10
		// Randomly choose one
		weightSum := 0
		for _, ability_ := range abilities {
			weightOfAbility := int(ability_.Weight) + smoothingFactor
			weightSum += weightOfAbility - common.ChannelWeights.GetPenaltyWeight(ability_.ChannelId, weightOfAbility-1)
		}
		// Randomly choose one
		weight := common.GetRandomInt(int(weightSum))
		for _, ability_ := range abilities {
			weightOfAbility := int(ability_.Weight) + smoothingFactor
			weight -= weightOfAbility - common.ChannelWeights.GetPenaltyWeight(ability_.ChannelId, weightOfAbility-1)
			//log.Printf("weight: %d, ability weight: %d", weight, *ability_.Weight)
			if weight <= 0 {
				channel.Id = ability_.ChannelId
				break
			}
		}
	} else {
		return nil, errors.New("channel not found")
	}
	err = DB.First(&channel, "id = ?", channel.Id).Error
	return &channel, err
}

func (channel *Channel) AddAbilities() error {
	models_ := strings.Split(channel.Models, ",")
	groups_ := strings.Split(channel.Group, ",")
	var modelMapping map[string]string
	if channel.ModelMapping != nil && (*channel.ModelMapping) != "" {
		s := *channel.ModelMapping
		if err := json.Unmarshal([]byte(s), &modelMapping); err != nil {
			return err
		}
	} else {
		modelMapping = make(map[string]string)
	}
	invModelMapping := make(map[string]string)
	for key, val := range modelMapping {
		invModelMapping[val] = key
	}
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			s := invModelMapping[model]
			ability := Ability{
				Group:      group,
				Model:      model,
				ChannelId:  channel.Id,
				Enabled:    channel.Status == common.ChannelStatusEnabled,
				Priority:   channel.Priority,
				Weight:     uint(channel.GetWeight()),
				Tag:        channel.Tag,
				ModelAlias: &s,
			}
			abilities = append(abilities, ability)
		}
	}
	if len(abilities) == 0 {
		return nil
	}
	for _, chunk := range lo.Chunk(abilities, 50) {
		err := DB.Create(&chunk).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (channel *Channel) DeleteAbilities() error {
	return DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities() error {
	// A quick and dirty way to update abilities
	// First delete all abilities of this channel
	err := channel.DeleteAbilities()
	if err != nil {
		return err
	}
	// Then add new abilities
	err = channel.AddAbilities()
	if err != nil {
		return err
	}
	return nil
}

func UpdateAbilityStatus(channelId int, status bool) error {
	return DB.Model(&Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}

func UpdateAbilityStatusByTag(tag string, status bool) error {
	return DB.Model(&Ability{}).Where("tag = ?", tag).Select("enabled").Update("enabled", status).Error
}

func UpdateAbilityByTag(tag string, newTag *string, priority *int64, weight *uint) error {
	ability := Ability{}
	if newTag != nil {
		ability.Tag = newTag
	}
	if priority != nil {
		ability.Priority = priority
	}
	if weight != nil {
		ability.Weight = *weight
	}
	return DB.Model(&Ability{}).Where("tag = ?", tag).Updates(ability).Error
}

func FixAbility() (int, error) {
	var channelIds []int
	count := 0
	// Find all channel ids from channel table
	err := DB.Model(&Channel{}).Pluck("id", &channelIds).Error
	if err != nil {
		common.SysError(fmt.Sprintf("Get channel ids from channel table failed: %s", err.Error()))
		return 0, err
	}
	// Delete abilities of channels that are not in channel table
	err = DB.Where("channel_id NOT IN (?)", channelIds).Delete(&Ability{}).Error
	if err != nil {
		common.SysError(fmt.Sprintf("Delete abilities of channels that are not in channel table failed: %s", err.Error()))
		return 0, err
	}
	common.SysLog(fmt.Sprintf("Delete abilities of channels that are not in channel table successfully, ids: %v", channelIds))
	count += len(channelIds)

	// Use channelIds to find channel not in abilities table
	var abilityChannelIds []int
	err = DB.Table("abilities").Distinct("channel_id").Pluck("channel_id", &abilityChannelIds).Error
	if err != nil {
		common.SysError(fmt.Sprintf("Get channel ids from abilities table failed: %s", err.Error()))
		return 0, err
	}
	var channels []Channel

	if len(abilityChannelIds) == 0 {
		err = DB.Find(&channels).Error
	} else {
		err = DB.Where("id NOT IN (?)", abilityChannelIds).Find(&channels).Error
	}
	if err != nil {
		return 0, err
	}
	for _, channel := range channels {
		err := channel.UpdateAbilities()
		if err != nil {
			common.SysError(fmt.Sprintf("Update abilities of channel %d failed: %s", channel.Id, err.Error()))
		} else {
			common.SysLog(fmt.Sprintf("Update abilities of channel %d successfully", channel.Id))
			count++
		}
	}
	InitChannelCache()
	return count, nil
}

func GetAlias2Models() map[string][]string {
	var abilities []*Ability
	var alias2models = make(map[string][]string)
	DB.Find(&abilities)
	for _, ability := range abilities {
		if ability.Enabled && ability.ModelAlias != nil && (*ability.ModelAlias) != "" {
			alias := *ability.ModelAlias
			if alias2models[alias] == nil {
				alias2models[alias] = make([]string, 0)
			}
			alias2models[alias] = append(alias2models[alias], ability.Model)
		}
	}
	return alias2models
}
