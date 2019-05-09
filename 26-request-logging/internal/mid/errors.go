package mid

import (
	"log"
	"net/http"

	"github.com/ardanlabs/garagesale/internal/platform/web"
	"github.com/pkg/errors"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *log.Logger) web.Middleware {

	// This is the actual middleware function to be executed.
	f := func(before web.Handler) web.Handler {

		h := func(w http.ResponseWriter, r *http.Request) error {

			// Run the handler chain and catch any propagated error.
			if err := before(w, r); err != nil {

				// If the error was of the type *web.Error, the handler has
				// a specific status code and error to return. If not, the
				// handler sent any arbitrary error value so use 500.
				webErr, ok := errors.Cause(err).(*web.Error)
				if !ok {
					webErr = &web.Error{
						Err:    err,
						Status: http.StatusInternalServerError,
						Fields: nil,
					}
				}

				// Log the error.
				log.Printf("ERROR : %+v", err)

				// Determine the error message service users will see. If the status
				// code is under 500 then it is a "human readable" error that was
				// intended for users to see. If the status code is 500 or higher (the
				// default) then use a generic error message.
				var errStr string
				if webErr.Status < http.StatusInternalServerError {
					errStr = webErr.Err.Error()
				} else {
					errStr = http.StatusText(webErr.Status)
				}

				// Respond with the error type we send to clients.
				res := web.ErrorResponse{
					Error:  errStr,
					Fields: webErr.Fields,
				}
				if err := web.Respond(r.Context(), w, res, webErr.Status); err != nil {
					return err
				}
			}

			// Return nil to indicate the error has been handled.
			return nil
		}

		return h
	}

	return f
}