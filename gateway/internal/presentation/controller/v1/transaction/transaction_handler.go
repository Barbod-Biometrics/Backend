package transaction

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	transactionUsecase usecase.TransactionUsecase
	profileRepository  repository.ProfileRepository
	logger             logger.Logger
}

// Create a new transaction handler
func NewTransactionHandler(transactionUsecase usecase.TransactionUsecase, logger logger.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionUsecase: transactionUsecase,
		profileRepository:  profileRepository,
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

	userID := h.getUserID(c)
	if c.IsAborted() {
		return
	}

	var req profile.GetUsageSummaryRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("invalid usage summary request format",
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request parameters"})
		return
	}

	userProfiles, err := h.profileRepository.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to fetch user profiles for security check",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal validation error"})
		return
	}

	isAuthorized := false
	for _, p := range userProfiles {
		if p.ProfileID == req.ProfileID {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		h.logger.Warn("unauthorized profile access attempt",
			logger.Field{Key: "user_id", Value: userID},
			logger.Field{Key: "target_profile_id", Value: req.ProfileID},
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this profile"})
		return
	}

	summary, err := h.transactionUsecase.GetUsageSummary(c.Request.Context(), req.ProfileID)
	if err != nil {
		h.logger.Error("failed to retrieve usage summary",
			logger.Field{Key: "profile_id", Value: req.ProfileID},
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve usage summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *TransactionHandler) getUserID(c *gin.Context) uint64 {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		h.logger.Warn("failed to extract user_id from context",
			logger.Field{Key: "error", Value: err},
		)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	return userID
}
