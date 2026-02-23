package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/gurodrigues-dev/tcp-server-client/session"
)

type RequestHandler struct {
	authManager       *AuthManager
	jobManager        *JobManager
	submissionManager *SubmissionManager
}

func NewRequestHandler(authManager *AuthManager, jobManager *JobManager, submissionManager *SubmissionManager) *RequestHandler {
	return &RequestHandler{
		authManager:       authManager,
		jobManager:        jobManager,
		submissionManager: submissionManager,
	}
}

func (rh *RequestHandler) HandleRequest(session *session.Session, req *Request) (*Response, error) {
	switch req.Method {
	case "authorize":
		return rh.handleAuthorize(session, req)
	case "submit":
		return rh.handleSubmit(session, req)
	default:
		return nil, fmt.Errorf("unknown method: %s", req.Method)
	}
}

func (rh *RequestHandler) handleAuthorize(session *session.Session, req *Request) (*Response, error) {
	authParams, err := parseParams[*AuthParams](req.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth params: %w", err)
	}

	if err := rh.authManager.Authorize(session, authParams); err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	return &Response{
		ID:     req.ID,
		Result: true,
	}, nil
}

func (rh *RequestHandler) handleSubmit(session *session.Session, req *Request) (*Response, error) {
	submissionParams, err := parseParams[*SubmissionParams](req.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse submission params: %w", err)
	}

	if rh.submissionManager == nil {
		return nil, errors.New("submission manager not initialized")
	}

	resp, err := rh.submissionManager.HandleSubmit(session, submissionParams)
	if err != nil {
		return nil, err
	}

	resp.ID = req.ID
	return resp, nil
}

func parseParams[T interface{}](params interface{}) (T, error) {
	var zero T

	if params == nil {
		return zero, nil
	}

	jsonData, err := json.Marshal(params)
	if err != nil {
		return zero, fmt.Errorf("failed to marshal params: %w", err)
	}

	resultType := reflect.TypeOf(zero)
	if resultType.Kind() == reflect.Ptr {
		resultType = resultType.Elem()
	}

	result := reflect.New(resultType).Interface()

	if err := json.Unmarshal(jsonData, result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal params: %w", err)
	}

	return result.(T), nil
}
