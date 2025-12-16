package transaction

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	transactionUsecase usecase.TransactionUsecase
	logger             logger.Logger
}

func NewTransactionHandler(transactionUsecase usecase.TransactionUsecase, logger logger.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionUsecase: transactionUsecase,
		logger:             logger,
	}
}

// GetUsageSummary returns the usage summary of services
// @Summary Get billing usage summary
// @Description Retrieves the total cost and service-wise breakdown for a specific profile.
// @Tags Billing
// @Accept json
// @Produce json
// @Param profile_id query int true "Profile ID"
// @Param from_date query string false "Filter start date (YYYY-MM-DD)" format(date)
// @Param to_date query string false "Filter end date (YYYY-MM-DD)" format(date)
// @Success 200 {object} profile.UsageSummaryResponse "Successful response with cost breakdown"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/billing/summary [get]
func (h *TransactionHandler) GetUsageSummary(c *gin.Context) {
	var req profile.GetUsageSummaryRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("invalid usage summary request format",
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request parameters"})
		return
	}

	summary, err := h.transactionUsecase.GetUsageSummary(c.Request.Context(), req.ProfileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve usage summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}
