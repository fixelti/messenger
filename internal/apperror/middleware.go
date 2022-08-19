package apperror

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// TODO: разобраться во всехблоках кода
type appHandler func(ctx *gin.Context) error

func Middleware(handler appHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var appErr *AppError
		var ok bool

		err := handler(ctx)
		if err != nil {
			if errors.As(err, &appErr) {
				if errors.Is(err, ErrNotFound) {
					ctx.IndentedJSON(http.StatusNotFound, ErrNotFound)
					return
				}
				if errors.Is(err, ErrNotAuth) {
					ctx.IndentedJSON(http.StatusUnauthorized, ErrNotAuth)
					return
				}

				appErr, ok = err.(*AppError)
				if !ok {
					panic("can't convert err to app error")
				}
				ctx.IndentedJSON(http.StatusBadRequest, appErr)
				return
			}
			ctx.IndentedJSON(http.StatusTeapot, appErr.systemError(err))
		}
	}
}