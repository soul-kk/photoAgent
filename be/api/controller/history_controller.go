package controller

import (
	"net/http"
	"strconv"

	"go-service-starter/core/libx"
	"go-service-starter/core/store/mysql"
	"go-service-starter/domain"
	"go-service-starter/repository"

	"github.com/gin-gonic/gin"
)

type HistoryController struct {
	repo domain.AnalysisHistoryRepository
}

func NewHistoryController() *HistoryController {
	db, err := mysql.InitMySQL()
	if err != nil {
		panic("mysql init failed: " + err.Error())
	}
	repo := repository.NewAnalysisHistoryRepo(db)
	return &HistoryController{repo: repo}
}

func (h *HistoryController) RegisterProtected(r *gin.RouterGroup) {
	r.GET("/history/analysis", h.ListAnalysisHistory)
	r.GET("/history/analysis/:id", h.GetAnalysisHistory)
}

type listHistoryQuery struct {
	Type   string `form:"type"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

func (h *HistoryController) ListAnalysisHistory(c *gin.Context) {
	uid := libx.Uid(c)

	var q listHistoryQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		libx.Err(c, http.StatusBadRequest, "参数无效", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	offset := (q.Page - 1) * q.Limit

	list, total, err := h.repo.ListByUserID(c.Request.Context(), uid, q.Type, q.Limit, offset)
	if err != nil {
		libx.Err(c, http.StatusInternalServerError, "查询历史记录失败", err)
		return
	}

	libx.Ok(c, "ok", gin.H{
		"list":  list,
		"total": total,
		"page":  q.Page,
		"limit": q.Limit,
	})
}

func (h *HistoryController) GetAnalysisHistory(c *gin.Context) {
	uid := libx.Uid(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		libx.Err(c, http.StatusBadRequest, "无效的 ID", err)
		return
	}

	history, err := h.repo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		libx.Err(c, http.StatusNotFound, "历史记录不存在", err)
		return
	}

	// 确保只能查看自己的记录
	if history.UserID != uid {
		libx.Err(c, http.StatusForbidden, "无权访问该记录", nil)
		return
	}

	libx.Ok(c, "ok", history)
}
