package gwu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

var (
	// ErrDecodeRequest failed to decode request. Is safe to display to the client. Log the error for debugging.
	ErrDecodeRequest = errors.New("failed to decode request")
	// ErrEncodeResponse failed to encode response. Is safe to display to the client. Log the error for debugging.
	ErrEncodeResponse = errors.New("failed to encode response")
)

// Logger defines gwu's minimally required logger.
type Logger interface {
	Debug(string, ...any)
	Info(string, ...any)
}

// IntoJSON writes the data as JSON with Content-Type `application/json` and given status code to the response.
// If the JSON encoding fails, it logs the error and writes ErrEncodeResponse to the response.
//
// Example usage:
//
//	web.IntoJSON(w, log, data, http.StatusOK)
func IntoJSON(w http.ResponseWriter, log Logger, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Info(fmt.Errorf("%w: %w", ErrEncodeResponse, err).Error())
		http.Error(w, ErrEncodeResponse.Error(), http.StatusInternalServerError)
	}
}

// HandleOpts are options for the Handle, CnIn, and Exec functions, use HandleOptsFunc to set the options.
// Use the HandleOpts to retrieve a contextual logger.
type HandleOpts struct {
	Log Logger
}

// HandleOptsFunc sets a HandleOpts option.
type HandleOptsFunc func(opt *HandleOpts)

// Log sets the logger for the HandleOpts.
func Log(log Logger) HandleOptsFunc {
	return func(opt *HandleOpts) {
		opt.Log = log
	}
}

// CnIn constructs the input of an Exec function.
// Commonly used are JSON, PathVal, and Empty.
//
// Important: Return only safe to display errors, Handle writes a CnIn function's error to the response with
// http.StatusBadRequest.
type CnIn[In any] func(*http.Request, HandleOpts) (In, error)

// Exec executes the endpoint logic. Pass it to Handle to retrieve an http.Handler.
// The Exec function should return the output, status code, and error.
//
// It is not recommended to make your service function Exec funcs, instead use controllers or some other sort of
// intermediate. An Exec is aware of its HTTP context and should only return client-safe error messages.
// Services contain business logic and may leak internal information.
//
// Important: Return only safe to display errors, Handle writes an Exec function's error to the response.
type Exec[In, Out any] func(context.Context, In, HandleOpts) (Out, int, error)

// JSON CnIn decodes the request body into the given data type In.
func JSON[In any]() CnIn[In] {
	return func(r *http.Request, _ HandleOpts) (In, error) {
		var in In
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			return in, ErrDecodeRequest
		}

		return in, nil
	}
}

// PathVal CnIn reads a path value with the given key.
func PathVal(key string) CnIn[string] {
	return func(r *http.Request, _ HandleOpts) (string, error) {
		return r.PathValue(key), nil
	}
}

// Empty CnIn always returns nil and no error.
// Use Empty for endpoints that do not require input.
func Empty() CnIn[any] {
	return func(_ *http.Request, _ HandleOpts) (any, error) {
		return nil, nil
	}
}

// ValIn Exec validates the input with the given validation function.
// If the validation fails, it returns an http.StatusBadRequest and the validation error.
// Afterward, it calls the given Exec function.
//
// Use ValIn to validate the input before executing the logic.
//
// ValIn expects the validation function to return an error that is safe to display to the client.
func ValIn[In, Out any](fn Exec[In, Out], fnVal func(in In) error) Exec[In, Out] {
	var out Out
	return func(ctx context.Context, in In, opts HandleOpts) (Out, int, error) {
		err := fnVal(in)
		if err != nil {
			return out, http.StatusBadRequest, err
		}

		return fn(ctx, in, opts)
	}
}

// Handle returns an http.Handler that executes the endpoint's logic with the given CnIn and Exec functions.
// Handle abstracts the HTTP boilerplate.
//
// If no Log option provides a logger, Handle instantiates a new slog.Logger with slog.TextHandler.
func Handle[In, Out any](inFn CnIn[In], fn Exec[In, Out], optFns ...HandleOptsFunc) http.Handler {
	var opts HandleOpts
	for _, fn := range optFns {
		fn(&opts)
	}

	if opts.Log == nil {
		opts.Log = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in, err := inFn(r, opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		out, code, err := fn(r.Context(), in, opts)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}

		IntoJSON(w, opts.Log, out, code)
	})
}
