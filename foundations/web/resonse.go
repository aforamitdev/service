package web

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

func Response(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {

	v, ok := ctx.Value(KeyValue).(*Values)

	if !ok {
		return NewShutdownError("web value missing from context")
	}
	v.StatusCode = statusCode

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}
	return nil
}

func RespondError(ctx context.Context, w http.ResponseWriter, err error) error {
	// if its a trusted error responser with the custom message , since we know it is going to be correct
	if webErr, ok := errors.Cause(err).(*Error); ok {
		err := ErrorResponse{
			Error:  webErr.Err.Error(),
			Fields: webErr.Fields,
		}
		if err := Response(ctx, w, err, webErr.Status); err != nil {
			return err
		}

		return nil
	}
	// not a trusted error
	er := ErrorResponse{
		Error: http.StatusText(http.StatusInternalServerError),
	}

	if err := Response(ctx, w, er, http.StatusInternalServerError); err != nil {
		return err
	}

	return nil
}
