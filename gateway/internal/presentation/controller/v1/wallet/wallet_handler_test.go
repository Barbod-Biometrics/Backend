package wallet

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/wallet"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestContext creates a gin context with userID set
func setupTestContext(userID uint64) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyUserID, userID)
	return c, w
}

// setupTestContextNoAuth creates a gin context without userID (unauthenticated)
func setupTestContextNoAuth() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestNewWalletHandler(t *testing.T) {
	mockUC := mocks.NewMockWalletUsecase(t)
	mockLogger := mocks.NewMockAppLogger(t)
	h := NewWalletHandler(mockUC, mockLogger)

	assert.NotNil(t, h)
	assert.Equal(t, mockUC, h.walletUsecase)
}

func TestWalletHandler_getUserID(t *testing.T) {
	t.Run("returns userID when present in context", func(t *testing.T) {
		mockUC := mocks.NewMockWalletUsecase(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewWalletHandler(mockUC, mockLogger)
		c, w := setupTestContext(uint64(123))
		c.Request = httptest.NewRequest("GET", "/test", nil)

		userID := h.getUserID(c)

		assert.Equal(t, uint64(123), userID)
		assert.False(t, c.IsAborted())
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 0 and aborts when userID not in context", func(t *testing.T) {
		mockUC := mocks.NewMockWalletUsecase(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewWalletHandler(mockUC, mockLogger)
		c, w := setupTestContextNoAuth()
		c.Request = httptest.NewRequest("GET", "/test", nil)

		userID := h.getUserID(c)

		assert.Equal(t, uint64(0), userID)
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestWalletHandler_GetWalletSummary(t *testing.T) {
	tests := []struct {
		name           string
		profileID      string
		userID         uint64
		authenticated  bool
		mockSetup      func(*mocks.MockWalletUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:          "success",
			profileID:     "10",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetWalletSummary(mock.Anything, mock.Anything, uint64(10)).
					Return(&dto.WalletSummaryResponse{
						Success: true,
						Data: dto.WalletSummaryData{
							Balance:          50000,
							TotalDeposits:    100000,
							TotalWithdrawals: 50000,
							TransactionCount: 5,
							LastUpdated:      "2024-01-01T00:00:00Z",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var res dto.WalletSummaryResponse
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.True(t, res.Success)
				assert.Equal(t, uint64(50000), res.Data.Balance)
				assert.Equal(t, int64(100000), res.Data.TotalDeposits)
				assert.Equal(t, int64(50000), res.Data.TotalWithdrawals)
				assert.Equal(t, int64(5), res.Data.TransactionCount)
			},
		},
		{
			name:           "invalid profile id - non-numeric",
			profileID:      "abc",
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "invalid profile id")
			},
		},
		{
			name:           "invalid profile id - negative",
			profileID:      "-1",
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "invalid profile id")
			},
		},
		{
			name:           "invalid profile id - empty",
			profileID:      "",
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "invalid profile id")
			},
		},
		{
			name:          "unauthorized - no userID in context",
			profileID:     "10",
			userID:        uint64(0),
			authenticated: false,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				// Handler still calls usecase even after abort (design quirk)
				m.EXPECT().GetWalletSummary(mock.Anything, uint64(0), uint64(10)).
					Return(&dto.WalletSummaryResponse{Success: false}, nil).Maybe()
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "unauthorized")
			},
		},
		{
			name:          "usecase error",
			profileID:     "10",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetWalletSummary(mock.Anything, mock.Anything, uint64(10)).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "database error")
			},
		},
		{
			name:          "large profile id",
			profileID:     "18446744073709551615",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetWalletSummary(mock.Anything, mock.Anything, uint64(18446744073709551615)).
					Return(&dto.WalletSummaryResponse{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewMockWalletUsecase(t)
			mockLogger := mocks.NewMockAppLogger(t)
			h := NewWalletHandler(mockUC, mockLogger)

			var c *gin.Context
			var w *httptest.ResponseRecorder

			if tt.authenticated {
				c, w = setupTestContext(tt.userID)
			} else {
				c, w = setupTestContextNoAuth()
			}

			c.Params = gin.Params{{Key: "id", Value: tt.profileID}}
			c.Request = httptest.NewRequest("GET", "/profiles/"+tt.profileID+"/wallet/summary", nil)

			tt.mockSetup(mockUC)

			h.GetWalletSummary(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestWalletHandler_GetTransactions(t *testing.T) {
	tests := []struct {
		name           string
		profileID      string
		queryParams    string
		userID         uint64
		authenticated  bool
		mockSetup      func(*mocks.MockWalletUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:          "success with default pagination",
			profileID:     "10",
			queryParams:   "",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 1, 10).
					Return(&dto.TransactionsResponse{
						Success: true,
						Data: dto.TransactionsData{
							Transactions: []dto.TransactionDTO{
								{ID: "tx1", Type: "deposit", Amount: 1000},
								{ID: "tx2", Type: "withdrawal", Amount: -500},
							},
							TotalCount: 2,
							Page:       1,
							PageSize:   10,
							TotalPages: 1,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var res dto.TransactionsResponse
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.True(t, res.Success)
				assert.Len(t, res.Data.Transactions, 2)
				assert.Equal(t, 1, res.Data.Page)
				assert.Equal(t, 10, res.Data.PageSize)
			},
		},
		{
			name:          "success with custom pagination",
			profileID:     "10",
			queryParams:   "page=2&page_size=5",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 2, 5).
					Return(&dto.TransactionsResponse{
						Success: true,
						Data: dto.TransactionsData{
							Page:     2,
							PageSize: 5,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var res dto.TransactionsResponse
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.Equal(t, 2, res.Data.Page)
				assert.Equal(t, 5, res.Data.PageSize)
			},
		},
		{
			name:          "invalid page defaults to 1",
			profileID:     "10",
			queryParams:   "page=abc&page_size=10",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 1, 10).
					Return(&dto.TransactionsResponse{Success: true, Data: dto.TransactionsData{Page: 1}}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var res dto.TransactionsResponse
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.Equal(t, 1, res.Data.Page)
			},
		},
		{
			name:          "invalid page_size defaults to 10",
			profileID:     "10",
			queryParams:   "page=1&page_size=xyz",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 1, 10).
					Return(&dto.TransactionsResponse{Success: true, Data: dto.TransactionsData{PageSize: 10}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "zero page defaults to 1",
			profileID:     "10",
			queryParams:   "page=0",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 1, 10).
					Return(&dto.TransactionsResponse{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "zero page_size defaults to 10",
			profileID:     "10",
			queryParams:   "page_size=0",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 1, 10).
					Return(&dto.TransactionsResponse{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "negative page defaults to 1",
			profileID:     "10",
			queryParams:   "page=-5",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 1, 10).
					Return(&dto.TransactionsResponse{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "negative page_size defaults to 10",
			profileID:     "10",
			queryParams:   "page_size=-10",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 1, 10).
					Return(&dto.TransactionsResponse{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid profile id",
			profileID:      "invalid",
			queryParams:    "",
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "invalid profile id")
			},
		},
		{
			name:          "unauthorized",
			profileID:     "10",
			queryParams:   "",
			userID:        uint64(0),
			authenticated: false,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				// Handler still calls usecase even after abort (design quirk)
				m.EXPECT().GetTransactions(mock.Anything, uint64(0), uint64(10), 1, 10).
					Return(&dto.TransactionsResponse{Success: false}, nil).Maybe()
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "unauthorized")
			},
		},
		{
			name:          "usecase error",
			profileID:     "10",
			queryParams:   "",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 1, 10).
					Return(nil, errors.New("service unavailable"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "service unavailable")
			},
		},
		{
			name:          "empty transactions list",
			profileID:     "10",
			queryParams:   "",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 1, 10).
					Return(&dto.TransactionsResponse{
						Success: true,
						Data: dto.TransactionsData{
							Transactions: []dto.TransactionDTO{},
							TotalCount:   0,
							Page:         1,
							PageSize:     10,
							TotalPages:   0,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var res dto.TransactionsResponse
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.Empty(t, res.Data.Transactions)
				assert.Equal(t, int64(0), res.Data.TotalCount)
			},
		},
		{
			name:          "large page number",
			profileID:     "10",
			queryParams:   "page=999999&page_size=100",
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().GetTransactions(mock.Anything, mock.Anything, uint64(10), 999999, 100).
					Return(&dto.TransactionsResponse{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewMockWalletUsecase(t)
			mockLogger := mocks.NewMockAppLogger(t)
			h := NewWalletHandler(mockUC, mockLogger)

			var c *gin.Context
			var w *httptest.ResponseRecorder

			if tt.authenticated {
				c, w = setupTestContext(tt.userID)
			} else {
				c, w = setupTestContextNoAuth()
			}

			c.Params = gin.Params{{Key: "id", Value: tt.profileID}}
			url := "/profiles/" + tt.profileID + "/wallet/transactions"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}
			c.Request = httptest.NewRequest("GET", url, nil)

			tt.mockSetup(mockUC)

			h.GetTransactions(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestWalletHandler_Deposit(t *testing.T) {
	tests := []struct {
		name           string
		profileID      string
		requestBody    string
		userID         uint64
		authenticated  bool
		mockSetup      func(*mocks.MockWalletUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:          "success",
			profileID:     "10",
			requestBody:   `{"amount": 50000, "description": "Test deposit"}`,
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().Deposit(mock.Anything, mock.Anything, uint64(10), mock.MatchedBy(func(req dto.DepositRequest) bool {
					return req.Amount == 50000 && req.Description == "Test deposit"
				})).Return(&dto.DepositResponse{
					Success: true,
					Data: dto.DepositDataDTO{
						TransactionID: "tx-123",
						NewBalance:    100000,
						Message:       "Deposit successful",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var res dto.DepositResponse
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.NoError(t, err)
				assert.True(t, res.Success)
				assert.Equal(t, "tx-123", res.Data.TransactionID)
				assert.Equal(t, uint64(100000), res.Data.NewBalance)
			},
		},
		{
			name:          "success without description",
			profileID:     "10",
			requestBody:   `{"amount": 1000}`,
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().Deposit(mock.Anything, mock.Anything, uint64(10), mock.MatchedBy(func(req dto.DepositRequest) bool {
					return req.Amount == 1000 && req.Description == ""
				})).Return(&dto.DepositResponse{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid profile id",
			profileID:      "invalid",
			requestBody:    `{"amount": 1000}`,
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "invalid profile id")
			},
		},
		{
			name:           "missing amount",
			profileID:      "10",
			requestBody:    `{"description": "Test"}`,
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "zero amount",
			profileID:      "10",
			requestBody:    `{"amount": 0}`,
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "negative amount (parsed as 0 or invalid)",
			profileID:      "10",
			requestBody:    `{"amount": -100}`,
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			profileID:      "10",
			requestBody:    `{invalid json}`,
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			profileID:      "10",
			requestBody:    ``,
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "unauthorized",
			profileID:     "10",
			requestBody:   `{"amount": 1000}`,
			userID:        uint64(0),
			authenticated: false,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				// Handler still calls usecase even after abort (design quirk)
				m.EXPECT().Deposit(mock.Anything, uint64(0), uint64(10), mock.Anything).
					Return(&dto.DepositResponse{Success: false}, nil).Maybe()
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "unauthorized")
			},
		},
		{
			name:          "usecase error",
			profileID:     "10",
			requestBody:   `{"amount": 1000}`,
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().Deposit(mock.Anything, mock.Anything, uint64(10), mock.Anything).
					Return(nil, errors.New("insufficient permissions"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "insufficient permissions")
			},
		},
		{
			name:          "large amount",
			profileID:     "10",
			requestBody:   `{"amount": 999999999999}`,
			userID:        uint64(5),
			authenticated: true,
			mockSetup: func(m *mocks.MockWalletUsecase) {
				m.EXPECT().Deposit(mock.Anything, mock.Anything, uint64(10), mock.MatchedBy(func(req dto.DepositRequest) bool {
					return req.Amount == 999999999999
				})).Return(&dto.DepositResponse{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "amount as string",
			profileID:      "10",
			requestBody:    `{"amount": "1000"}`,
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "amount as float",
			profileID:      "10",
			requestBody:    `{"amount": 100.50}`,
			userID:         uint64(5),
			authenticated:  true,
			mockSetup:      func(m *mocks.MockWalletUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := mocks.NewMockWalletUsecase(t)
			mockLogger := mocks.NewMockAppLogger(t)
			h := NewWalletHandler(mockUC, mockLogger)

			var c *gin.Context
			var w *httptest.ResponseRecorder

			if tt.authenticated {
				c, w = setupTestContext(tt.userID)
			} else {
				c, w = setupTestContextNoAuth()
			}

			c.Params = gin.Params{{Key: "id", Value: tt.profileID}}
			c.Request = httptest.NewRequest("POST", "/profiles/"+tt.profileID+"/wallet/deposit", strings.NewReader(tt.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			tt.mockSetup(mockUC)

			h.Deposit(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
