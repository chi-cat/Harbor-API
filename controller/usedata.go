package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/model"
	"strconv"
)

func GetAllQuotaDates(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	username := c.Query("username")
	dates, err := model.GetAllQuotaDates(startTimestamp, endTimestamp, username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    dates,
	})
	return
}

// GetUserQuotaDates 获取用户配额数据
// 该函数通过用户的ID和指定的时间范围来获取用户的配额数据
// 参数:
//
//	c *gin.Context: Gin框架的上下文对象，用于处理HTTP请求和响应
func GetUserQuotaDates(c *gin.Context) {
	// 从上下文中获取用户ID
	userId := c.GetInt("id")
	// 解析开始时间戳
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	// 解析结束时间戳
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	// 判断时间跨度是否超过 1 个月
	if endTimestamp-startTimestamp > 2592000 {
		// 如果时间跨度超过1个月，返回错误信息
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "时间跨度不能超过 1 个月",
		})
		return
	}
	// 调用模型获取用户配额数据
	dates, err := model.GetQuotaDataByUserId(userId, startTimestamp, endTimestamp)
	if err != nil {
		// 如果获取数据失败，返回错误信息
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	// 如果获取数据成功，返回配额数据
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    dates,
	})
	return
}
