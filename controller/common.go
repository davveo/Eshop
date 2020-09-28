package controller

import (
	"net/http"
	"orange/global/consts"
	"orange/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SiteShow(ctx *gin.Context) {
	config := model.CreateSettingFactory("").Get(consts.SITE)
	if config == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取数据失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, config)
}

func FocusPicture(ctx *gin.Context) {
	clientType := ctx.Query("client_type") //APP/WAP/PC

	data, err := model.CreateFocusPictureFactory("").List(clientType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func Page(ctx *gin.Context) {
	pageType := ctx.Query("page_type")     // APP/WAP/PC
	clientType := ctx.Query("client_type") // INDEX/SPECIAL
	data, err := model.CreatePageFactory("").GetByType(clientType, pageType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func PageSiteNavigationList(ctx *gin.Context) {
	queryParams := make(map[string]interface{})

	clientType := ctx.Query("client_type")
	pageNo, _ := strconv.Atoi(ctx.DefaultQuery("page_no", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	queryParams["page_no"] = pageNo
	queryParams["page_size"] = pageSize
	queryParams["client_type"] = clientType
	data, dataTotal := model.CreateSiteNavigationFactory("").List(queryParams)

	ctx.JSON(http.StatusOK, gin.H{
		"data":       data,
		"data_total": dataTotal,
		"page_no":    pageNo,
		"page_size":  pageSize,
	})
}
