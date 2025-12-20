package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/gin-gonic/gin"
)

const (
	genericError = "errors.generic"
)

type RecoveryMiddleware struct {
	constants *bootstrap.Constants
}

func NewRecovery(constants *bootstrap.Constants) *RecoveryMiddleware {
	return &RecoveryMiddleware{constants: constants}
}

func (recovery RecoveryMiddleware) Recovery(ctx *gin.Context) {
	defer func() {
		if rec := recover(); rec != nil {
			if err, ok := rec.(error); ok {
				recovery.handleRecoveredError(ctx, err)
			} else {
				recovery.handleRecoveredError(ctx, errors.New("panic"))
			}
			ctx.Abort()
		}
	}()

	ctx.Next()
}

func (recovery RecoveryMiddleware) handleRecoveredError(ctx *gin.Context, err error) {
	switch e := err.(type) {
	case exception.ValidationErrors:
		handleValidationError(ctx, e)
		return
	case exception.BindingError:
		handleBindingError(ctx, e)
		return
	case *exception.AuthError:
		handleAuthError(ctx, e)
		return
	case *exception.AppError:
		handleAppError(ctx, e)
		return
	case exception.NotFoundError:
		handleNotFoundError(ctx, e)
		return
	case exception.ForbiddenError:
		handleForbiddenError(ctx, e)
		return
	default:
		trans := GetTranslator(ctx)
		var msg string
		if trans != nil {
			m, _ := trans.Translate(genericError)
			msg = m
		} else {
			msg = "internal server error"
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": "ERR_INTERNAL", "message": msg})
	}

}

func handleAppError(ctx *gin.Context, appErr *exception.AppError) {
	trans := GetTranslator(ctx)
	var message string
	if trans != nil {
		if t, err := trans.Translate(appErr.Code); err == nil && t != appErr.Code {
			message = t
		} else {
			message = appErr.Message
		}
	} else {
		message = appErr.Message
	}

	status := appErr.HTTPStatus
	if status == 0 {
		status = 500
	}

	ctx.JSON(status, gin.H{"code": appErr.Code, "message": message})
}

func handleValidationError(ctx *gin.Context, validationErrors exception.ValidationErrors) {
	trans := GetTranslator(ctx)
	errorMessages := make(map[string]map[string]string)

	for _, validationError := range validationErrors.Errors {
		if _, ok := errorMessages[validationError.Field]; !ok {
			errorMessages[validationError.Field] = make(map[string]string)
		}
		fieldName, _ := trans.Translate(validationError.Field)
		message, _ := trans.Translate(fmt.Sprintf("errors.%s", validationError.Tag), fieldName)
		errorMessages[validationError.Field][validationError.Tag] = message
	}

	ctx.JSON(422, gin.H{"code": "ERR_VALIDATION", "errors": errorMessages})
}

func handleBindingError(ctx *gin.Context, bindingErr exception.BindingError) {
	trans := GetTranslator(ctx)
	message, _ := trans.Translate(genericError)

	if numError, ok := bindingErr.Err.(*strconv.NumError); ok {
		message, _ = trans.Translate("errors.numeric", numError.Num)
	} else if bindingErr.Err == http.ErrMissingFile {
		message, _ = trans.Translate("errors.fileRequired")
	}

	ctx.JSON(400, gin.H{"code": "ERR_BINDING", "message": message})
}

func handleAuthError(ctx *gin.Context, authErr *exception.AuthError) {
	trans := GetTranslator(ctx)

	message, _ := trans.Translate(genericError)
	switch authErr.Type {
	case exception.ErrorTypeInvalidCredentials:
		message, _ = trans.Translate("errors.invalidAuthCredentials")
	case exception.ErrorTypeExpiredToken:
		message, _ = trans.Translate("errors.expiredAuthToken")
	case exception.ErrorTypeInvalidToken:
		message, _ = trans.Translate("errors.invalidAuthToken")
	case exception.ErrorTypeUnauthorized:
		message, _ = trans.Translate("errors.unauthorized")
	}

	code := "ERR_UNAUTHORIZED"
	switch authErr.Type {
	case exception.ErrorTypeInvalidCredentials:
		code = "ERR_INVALID_CREDENTIALS"
	case exception.ErrorTypeExpiredToken:
		code = "ERR_EXPIRED_TOKEN"
	case exception.ErrorTypeInvalidToken:
		code = "ERR_INVALID_TOKEN"
	case exception.ErrorTypeUnauthorized:
		code = "ERR_UNAUTHORIZED"
	}

	ctx.JSON(401, gin.H{"code": code, "message": message})
}

func handleNotFoundError(ctx *gin.Context, notFoundErr exception.NotFoundError) {
	trans := GetTranslator(ctx)
	itemName, _ := trans.Translate(notFoundErr.Item)
	message, _ := trans.Translate("errors.notFound", itemName)
	ctx.JSON(404, gin.H{"code": "ERR_NOT_FOUND", "message": message})
}

func handleForbiddenError(ctx *gin.Context, forbiddenErr exception.ForbiddenError) {
	trans := GetTranslator(ctx)
	resourceName, _ := trans.Translate(forbiddenErr.Resource)
	message, _ := trans.Translate("errors.forbiddenError", resourceName)
	switch forbiddenErr.Type {
	case exception.ForbiddenTypeBannedUser:
		message, _ = trans.Translate("errors.bannedUser")
	}

	code := "ERR_FORBIDDEN"
	switch forbiddenErr.Type {
	case exception.ForbiddenTypeBannedUser:
		code = "ERR_BANNED_USER"
	}
	ctx.JSON(403, gin.H{"code": code, "message": message})
}
