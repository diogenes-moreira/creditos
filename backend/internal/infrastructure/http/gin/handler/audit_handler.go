package handler

import (
	"net/http"

	"github.com/diogenes-moreira/creditos/backend/internal/application/dto"
	"github.com/diogenes-moreira/creditos/backend/internal/application/service"
	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditService *service.AuditService
}

func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// GetAuditLogs godoc
// @Summary Get audit logs
// @Description Returns a paginated list of audit log entries recording all system operations
// @Tags Audit
// @Produce json
// @Param offset query int false "Pagination offset" default(0)
// @Param limit query int false "Pagination limit" default(50)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /admin/audit [get]
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	var req dto.PaginationRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit <= 0 {
		req.Limit = 50
	}
	logs, total, err := h.auditService.FindAll(c.Request.Context(), req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse{Data: dto.ToAuditLogResponses(logs), Total: total, Offset: req.Offset, Limit: req.Limit})
}
