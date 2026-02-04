package session

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	workflowsession "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/workflow_session"
	sessionpkg "github.com/Barbod-Biometrics/Backend/gateway/internal/application/session"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	svc        usecase.SessionUsecase
	mgr        *sessionpkg.SessionManager
	configRepo repository.WorkflowConfigRepository
	l          logger.Logger
}

func NewSessionHandler(svc usecase.SessionUsecase, mgr *sessionpkg.SessionManager, configRepo repository.WorkflowConfigRepository, l logger.Logger) *SessionHandler {
	return &SessionHandler{svc: svc, mgr: mgr, configRepo: configRepo, l: l}
}

// Start godoc
// @Summary Initialize a workflow session
// @Tags Workflow
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body workflowsession.StartWorkflowRequest true "Workflow configuration"
// @Success 202 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /workflow [post]
func (h *SessionHandler) Start(c *gin.Context) {

	var pid uint64
	if v, ok := c.Get("profile_id"); ok {
		if id, ok := v.(uint64); ok {
			pid = id
		}
	}

	var req workflowsession.StartWorkflowRequest
	if err := c.BindJSON(&req); err != nil {
		panic(exception.BindingError{Err: err})
	}

	clientIP := c.ClientIP()

	id, err := h.svc.StartAndRun(c.Request.Context(), pid, req.WorkflowConfigID, clientIP)
	if err != nil {
		h.l.Error("Failed to start session", logger.Field{Key: "error", Value: err})
		if _, ok := err.(*exception.AppError); ok {
			panic(err)
		}
		panic(exception.NewInternalError("ERR_FAILED_TO_START", "failed to start session", err))
	}

	sess, _ := h.mgr.GetSession(id)
	ui := map[string]interface{}{}
	if sess != nil && sess.Metadata != nil {
		if instr, ok := sess.Metadata["instruction"].(string); ok {
			ui["instructions"] = instr
		}
		if liveSent, ok := sess.Metadata["liveness_sentence"].(string); ok {
			ui["liveness_sentence"] = liveSent
		}
	}

	// prepare translated status/next_step/message when translator is available
	statusStr := "initialized"
	nextStepStr := "UPLOAD_IMAGE_OCR"
	messageStr := "initialized"
	if tr := middleware.GetTranslator(c); tr != nil {
		if t, err := tr.Translate("workflow.status.initialized"); err == nil && t != "" {
			statusStr = t
		}
		if t, err := tr.Translate("workflow.next_step.upload_image_ocr"); err == nil && t != "" {
			nextStepStr = t
		}
		if t, err := tr.Translate("successMessage.workflowInitialized"); err == nil && t != "" {
			messageStr = t
		}
	}

	resp := map[string]interface{}{
		"status":     statusStr,
		"message":    messageStr,
		"next_step":  nextStepStr,
		"ui_config":  ui,
		"session_id": id,
	}
	c.JSON(http.StatusAccepted, resp)
}

// Upload godoc
// @Summary Upload file for current workflow step
// @Description For OCR step: upload 'file' (image). The OCR image is automatically cropped for face verification. For face verification: upload 'base_image' (optional - will use auto-cropped OCR image if not provided) and 'video' (liveness video) together.
// @Tags Workflow
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Session ID"
// @Param file formData file false "Image file for OCR step"
// @Param base_image formData file false "Base/reference image for face verification (optional - uses auto-cropped OCR image if not provided)"
// @Param video formData file false "Liveness video for face verification"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /workflow/{id}/upload [post]
func (h *SessionHandler) Upload(c *gin.Context) {
	id := c.Param("id")

	s, err := h.mgr.GetSession(id)
	if err != nil {
		panic(exception.NewNotFoundError("session"))
	}

	var nextStep string
	var result interface{}
	var croppedImageBytes []byte

	switch s.State {
	case entity.StatePendingOCR:
		file, err := c.FormFile("file")
		if err != nil {
			panic(exception.ErrEmptyImage)
		}

		src, err := file.Open()
		if err != nil {
			panic(exception.NewBadRequestError("ERR_FAILED_TO_READ_FILE", "failed to read file", err))
		}
		defer src.Close()

		fileBytes, _ := io.ReadAll(src)

		ocrResult, err := h.svc.ProcessOCR(c.Request.Context(), s, fileBytes)
		if err != nil {
			if _, ok := err.(*exception.AppError); ok {
				panic(err)
			}
			panic(exception.NewInternalError("ERR_PROCESS_OCR", "failed to process ocr", err))
		}
		result = ocrResult

		s.Lock()
		currentState := s.State
		s.Unlock()

		if currentState == entity.StateOCRPendingAcceptance {
			nextStep = "ACCEPT_OCR_RESULT"
		} else {
			nextStep = "UPLOAD_FACE_VERIFICATION"
		}

		s.Lock()
		if autoCropped, ok := s.Metadata["auto_cropped_image_bytes"]; ok {
			if cropBytes, ok := autoCropped.([]byte); ok && len(cropBytes) > 0 {
				croppedImageBytes = cropBytes
			}
		}
		s.Unlock()

	case entity.StateOCRPendingAcceptance:
		panic(exception.NewBadRequestError("ERR_OCR_PENDING_ACCEPTANCE", "OCR result is pending acceptance. Please accept the OCR result first.", nil))

	case entity.StateOCRSuccess, entity.StatePendingLiveness:
		baseImageFile, _ := c.FormFile("base_image")

		var baseImageBytes []byte
		if baseImageFile != nil {
			baseSrc, err := baseImageFile.Open()
			if err != nil {
				panic(exception.NewBadRequestError("ERR_FAILED_TO_READ_BASE_IMAGE", "failed to read base image", err))
			}
			defer baseSrc.Close()
			baseImageBytes, _ = io.ReadAll(baseSrc)
		}

		videoFile, err := c.FormFile("video")
		if err != nil {
			panic(exception.ErrEmptyVideo)
		}

		videoSrc, err := videoFile.Open()
		if err != nil {
			panic(exception.NewBadRequestError("ERR_FAILED_TO_READ_VIDEO", "failed to read video", err))
		}
		defer videoSrc.Close()
		videoBytes, _ := io.ReadAll(videoSrc)

		h.l.Info("Processing face verification", logger.Field{Key: "session_id", Value: id})
		fvResult, err := h.svc.ProcessFaceVerification(c.Request.Context(), s, baseImageBytes, videoBytes)
		h.l.Info("Face verification service returned", logger.Field{Key: "session_id", Value: id}, logger.Field{Key: "has_error", Value: err != nil})

		if err != nil {
			h.l.Error("Face verification processing error", logger.Field{Key: "error", Value: err}, logger.Field{Key: "session_id", Value: id})
			if _, ok := err.(*exception.AppError); ok {
				panic(err)
			}
			panic(exception.NewInternalError("ERR_PROCESS_FACE_VERIFICATION", "failed to process face verification", err))
		}

		h.l.Info("Face verification completed successfully", logger.Field{Key: "session_id", Value: id}, logger.Field{Key: "result_type", Value: fmt.Sprintf("%T", fvResult)})
		result = fvResult
		nextStep = "COMPLETED"

	default:
		currentState := s.State
		ae := exception.ErrInvalidRequest
		panic(struct {
			exception.AppError
			CurrentState interface{}
		}{*ae, currentState})
	}

	s.Lock()
	currentState := s.State
	s.Unlock()

	if len(croppedImageBytes) > 0 {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		jsonPart, _ := writer.CreateFormField("metadata")
		jsonData := fmt.Sprintf(`{"status":"success","next_step":"%s","session_id":"%s","current_state":"%s"}`, nextStep, id, currentState)

		jsonPart.Write([]byte(jsonData))

		imagePart, _ := writer.CreateFormFile("cropped_image", "cropped.jpg")
		imagePart.Write(croppedImageBytes)

		writer.Close()

		c.Header("Content-Type", writer.FormDataContentType())
		c.Data(http.StatusOK, writer.FormDataContentType(), body.Bytes())
		return
	}

	resp := map[string]interface{}{
		"status":        "success",
		"next_step":     nextStep,
		"session_id":    id,
		"current_state": currentState,
	}

	if result != nil {
		resp["result"] = result
	}

	if len(s.Errors) > 0 {
		resp["errors"] = s.Errors
	}

	h.l.Info("Sending upload response", logger.Field{Key: "session_id", Value: id}, logger.Field{Key: "status", Value: resp["status"]})
	c.JSON(http.StatusOK, resp)
}

// Get returns session state
// @Summary Get workflow session state
// @Tags Workflow
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /workflow/{id} [get]
func (h *SessionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	s, err := h.mgr.GetSession(id)
	if err != nil {
		panic(exception.NewNotFoundError("session"))
	}

	nextStep := h.getNextStep(s.State)

	resp := map[string]interface{}{
		"id":                 s.ID,
		"profile_id":         s.ProfileID,
		"workflow_config_id": s.WorkflowConfigID,
		"client_ip":          s.ClientIP,
		"state":              s.State,
		"next_step":          nextStep,
		"created_at":         s.CreatedAt,
		"updated_at":         s.UpdatedAt,
		"expires_at":         s.ExpiresAt,
		"metadata":           s.Metadata,
	}

	if s.OCRResult != nil {
		resp["ocr_result"] = s.OCRResult
	}

	if s.FaceResult != nil {
		resp["face_result"] = s.FaceResult
	}

	if len(s.Errors) > 0 {
		resp["errors"] = s.Errors
	}

	c.JSON(http.StatusOK, resp)
}

func (h *SessionHandler) getNextStep(state entity.SessionState) string {
	switch state {
	case entity.StateCreated, entity.StatePendingOCR:
		return "UPLOAD_IMAGE_OCR"
	case entity.StateOCRPendingAcceptance:
		return "ACCEPT_OCR_RESULT"
	case entity.StateOCRSuccess:
		return "UPLOAD_FACE_VERIFICATION"
	case entity.StateOCRFailed:
		return "UPLOAD_IMAGE_OCR"
	case entity.StateBaseImageSuccess, entity.StatePendingLiveness:
		return "UPLOAD_FACE_VERIFICATION"
	case entity.StateLivenessSuccess:
		return "COMPLETED"
	case entity.StateLivenessFailed:
		return "UPLOAD_FACE_VERIFICATION"
	case entity.StateCompleted:
		return "COMPLETED"
	default:
		return "UNKNOWN"
	}
}

// Cancel cancels a session
// @Summary Cancel a workflow session
// @Tags Workflow
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /workflow/{id}/cancel [post]
func (h *SessionHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	s, err := h.mgr.GetSession(id)
	if err != nil {
		panic(exception.NewNotFoundError("session"))
	}
	if err := h.mgr.CancelSession(s); err != nil {
		panic(exception.NewInternalError("ERR_FAILED_TO_CANCEL", "failed to cancel session", err))
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

// AcceptOCR accepts the OCR result and moves the workflow forward
// @Summary Accept OCR result
// @Tags Workflow
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /workflow/{id}/accept-ocr [post]
func (h *SessionHandler) AcceptOCR(c *gin.Context) {
	id := c.Param("id")

	s, err := h.mgr.GetSession(id)
	if err != nil {
		panic(exception.NewNotFoundError("session"))
	}

	if err := h.svc.AcceptOCR(c.Request.Context(), s); err != nil {
		panic(exception.NewBadRequestError("ERR_ACCEPT_OCR", "failed to accept OCR result", err))
	}

	// Get updated session to return current state
	s, _ = h.mgr.GetSession(id)
	nextStep := h.getNextStep(s.State)

	c.JSON(http.StatusOK, gin.H{
		"message":   "OCR result accepted",
		"state":     s.State,
		"next_step": nextStep,
	})
}
