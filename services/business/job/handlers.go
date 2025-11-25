package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/szjason72/zervigo/shared/core/response"
	"gorm.io/gorm"
)

type jobHandler struct {
	service *JobService
}

func newJobHandler(service *JobService) *jobHandler {
	return &jobHandler{service: service}
}

func (h *jobHandler) registerRoutes(r *gin.Engine) {
	jobGroup := r.Group("/api/v1/jobs")
	{
		jobGroup.GET("", h.handleListJobs)
		jobGroup.GET("/:id", h.handleGetJob)
		jobGroup.POST("", h.handleCreateJob)
		jobGroup.PUT("/:id", h.handleUpdateJob)
		jobGroup.GET("/search", h.handleSearchJobs)
		jobGroup.POST("/favorite", h.handleFavoriteJob)
		jobGroup.DELETE("/favorite/:id", h.handleUnfavoriteJob)
	}
}

func (h *jobHandler) handleListJobs(c *gin.Context) {
	filter := parseJobFilter(c)
	result, err := h.service.ListJobs(c.Request.Context(), filter)
	if err != nil {
		writeError(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	writeSuccess(c, "获取职位列表成功", result)
}

func (h *jobHandler) handleSearchJobs(c *gin.Context) {
	filter := parseJobFilter(c)
	filter.Keyword = strings.TrimSpace(c.Query("keyword"))
	result, err := h.service.ListJobs(c.Request.Context(), filter)
	if err != nil {
		writeError(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}
	writeSuccess(c, "搜索职位成功", result.Items)
}

func (h *jobHandler) handleGetJob(c *gin.Context) {
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, response.CodeInvalidParams, "无效的职位ID")
		return
	}

	detail, err := h.service.GetJob(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		code := response.CodeInternalError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			code = response.CodeNotFound
		}
		writeError(c, status, code, err.Error())
		return
	}
	writeSuccess(c, "获取职位详情成功", detail)
}

func (h *jobHandler) handleCreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, response.CodeInvalidParams, err.Error())
		return
	}

	detail, err := h.service.CreateJob(c.Request.Context(), req)
	if err != nil {
		writeError(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	writeSuccess(c, "创建职位成功", detail)
}

func (h *jobHandler) handleUpdateJob(c *gin.Context) {
	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, response.CodeInvalidParams, "无效的职位ID")
		return
	}

	var req UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, response.CodeInvalidParams, err.Error())
		return
	}

	detail, err := h.service.UpdateJob(c.Request.Context(), id, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := response.CodeInternalError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			code = response.CodeNotFound
		}
		writeError(c, status, code, err.Error())
		return
	}

	writeSuccess(c, "更新职位成功", detail)
}

func (h *jobHandler) handleFavoriteJob(c *gin.Context) {
	var req FavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, response.CodeInvalidParams, err.Error())
		return
	}
	if req.UserID == 0 {
		writeError(c, http.StatusBadRequest, response.CodeInvalidParams, "userId 不能为空")
		return
	}

	ctx := c.Request.Context()
	if err := h.service.ToggleFavorite(ctx, req.UserID, req.JobID, true); err != nil {
		writeError(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	writeSuccess(c, "收藏成功", gin.H{"favorite": true})
}

func (h *jobHandler) handleUnfavoriteJob(c *gin.Context) {
	jobID, err := parseUintParam(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, response.CodeInvalidParams, "无效的职位ID")
		return
	}
	userIDParam := strings.TrimSpace(c.DefaultQuery("userId", "0"))
	userID, err := strconv.ParseUint(userIDParam, 10, 64)
	if err != nil || userID == 0 {
		writeError(c, http.StatusBadRequest, response.CodeInvalidParams, "userId 参数缺失或无效")
		return
	}

	ctx := c.Request.Context()
	if err := h.service.ToggleFavorite(ctx, uint(userID), jobID, false); err != nil {
		writeError(c, http.StatusInternalServerError, response.CodeInternalError, err.Error())
		return
	}

	writeSuccess(c, "取消收藏成功", gin.H{"favorite": false})
}

func parseJobFilter(c *gin.Context) JobFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	categories := strings.Split(strings.TrimSpace(c.Query("categories")), ",")
	if len(categories) == 1 && categories[0] == "" {
		categories = []string{}
	}

	return JobFilter{
		Page:       page,
		PageSize:   size,
		Keyword:    strings.TrimSpace(c.Query("keyword")),
		WorkType:   strings.TrimSpace(c.Query("workType")),
		Experience: strings.TrimSpace(c.Query("experience")),
		Status:     strings.TrimSpace(c.Query("status")),
		Categories: categories,
	}
}

func parseUintParam(param string) (uint, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(param), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func writeSuccess(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, response.Success(message, data))
}

func writeError(c *gin.Context, status int, code int, message string) {
	c.JSON(status, response.Error(code, message))
}
