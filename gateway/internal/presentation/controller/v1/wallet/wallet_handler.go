package wallet

import (
	"net/http"
	"strconv"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/wallet"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletUsecase usecase.WalletUsecase
	logger        logger.Logger
}

func NewWalletHandler(u usecase.WalletUsecase, logger logger.Logger) *WalletHandler {
	return &WalletHandler{
		walletUsecase: u,
		logger:        logger,
	}
}

// getUserID extracts the authenticated user's ID from the request context.
func (h *WalletHandler) getUserID(c *gin.Context) uint64 {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		panic(exception.NewUnauthorizedError("", nil))
	}
	return userID
}

// GetWalletSummary returns wallet summary for a profile
// @Summary Get Wallet Summary
// @Security BearerAuth
// @Description Get wallet summary including balance, total deposits, total withdrawals, and transaction count
// @Tags Wallet
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Success 200 {object} wallet.WalletSummaryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/wallet/summary [get]
func (h *WalletHandler) GetWalletSummary(c *gin.Context) {
	profileIDParam := c.Param("id")
	profileID, err := strconv.ParseUint(profileIDParam, 10, 64)
	if err != nil {
		panic(exception.NewBadRequestError("ERR_INVALID_PROFILE_ID", "Invalid profile ID", err))
	}

	userID := h.getUserID(c)
	resp, err := h.walletUsecase.GetWalletSummary(c.Request.Context(), userID, profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTransactions returns paginated list of transactions for a profile
// @Summary Get Wallet Transactions
// @Security BearerAuth
// @Description Get paginated list of wallet transactions for a profile
// @Tags Wallet
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Success 200 {object} wallet.TransactionsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/wallet/transactions [get]
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	profileIDParam := c.Param("id")
	profileID, err := strconv.ParseUint(profileIDParam, 10, 64)
	if err != nil {
		panic(exception.NewBadRequestError("ERR_INVALID_PROFILE_ID", "Invalid profile ID", err))
	}

	// Parse pagination params
	page := 1
	pageSize := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	userID := h.getUserID(c)
	resp, err := h.walletUsecase.GetTransactions(c.Request.Context(), userID, profileID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Deposit handles deposit to a profile's wallet
// @Summary Deposit to Wallet
// @Security BearerAuth
// @Description Make a deposit to the profile's wallet. Amount is in Iranian Rials (no decimal points).
// @Tags Wallet
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Param request body wallet.DepositRequest true "Deposit Request"
// @Success 200 {object} wallet.DepositResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/wallet/deposit [post]
func (h *WalletHandler) Deposit(c *gin.Context) {
	profileIDParam := c.Param("id")
	profileID, err := strconv.ParseUint(profileIDParam, 10, 64)
	if err != nil {
		panic(exception.NewBadRequestError("ERR_INVALID_PROFILE_ID", "Invalid profile ID", err))
	}

	var req wallet.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		panic(exception.BindingError{Err: err})
	}

	userID := h.getUserID(c)
	resp, err := h.walletUsecase.Deposit(c.Request.Context(), userID, profileID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
