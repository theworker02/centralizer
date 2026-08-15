// Package czerr defines typed errors returned by Centralizer public and
// internal APIs. Callers should use errors.Is / errors.As.
package czerr

import (
	"errors"
	"fmt"
)

// Sentinel error categories. Wrap these with extra context rather than
// inventing parallel types for the same failure class.
var (
	ErrTargetNotFound     = errors.New("centralizer: target not found")
	ErrRuntimeUnavailable = errors.New("centralizer: runtime unavailable")
	ErrUnsupportedTarget  = errors.New("centralizer: unsupported target")
	ErrBridgeUnavailable  = errors.New("centralizer: bridge unavailable")
	ErrBridgeFailed       = errors.New("centralizer: bridge failed")
	ErrProtocolMismatch   = errors.New("centralizer: protocol mismatch")
	ErrSchemaMismatch     = errors.New("centralizer: schema mismatch")
	ErrConversion         = errors.New("centralizer: conversion failed")
	ErrTimeout            = errors.New("centralizer: timeout")
	ErrCancelled          = errors.New("centralizer: canceled")
	ErrHandleInvalid      = errors.New("centralizer: invalid handle")
	ErrPolicyDenied       = errors.New("centralizer: policy denied")
	ErrAdapterFailure     = errors.New("centralizer: adapter failure")
	ErrTransportFailure   = errors.New("centralizer: transport failure")
	ErrDiscoveryFailed    = errors.New("centralizer: discovery failed")
	ErrPlannerFailed      = errors.New("centralizer: planner failed")
	ErrManifestInvalid    = errors.New("centralizer: invalid manifest")
	ErrSchemaInvalid      = errors.New("centralizer: invalid schema")
	ErrFrameInvalid       = errors.New("centralizer: invalid protocol frame")
	ErrPayloadTooLarge    = errors.New("centralizer: payload too large")
	ErrQuarantined        = errors.New("centralizer: target quarantined")
	ErrCircuitOpen        = errors.New("centralizer: circuit breaker open")
	ErrNotImplemented     = errors.New("centralizer: not implemented")
	ErrClosed             = errors.New("centralizer: closed")
	ErrInvalidArgument    = errors.New("centralizer: invalid argument")
	ErrSecurity           = errors.New("centralizer: security violation")
	ErrCache              = errors.New("centralizer: cache error")
	ErrStreamClosed       = errors.New("centralizer: stream closed")
	ErrExperimental       = errors.New("centralizer: experimental feature disabled")
)

// Error is a typed Centralizer error that preserves a category sentinel
// and an optional structured detail map for diagnostics.
type Error struct {
	Kind    error
	Message string
	Detail  map[string]string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Cause)
	}
	if e.Message == "" {
		return e.Kind.Error()
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	if e.Cause != nil {
		return fmt.Errorf("%w: %w", e.Kind, e.Cause)
	}
	return e.Kind
}

func (e *Error) Is(target error) bool {
	if e == nil {
		return false
	}
	return errors.Is(e.Kind, target) || (e.Cause != nil && errors.Is(e.Cause, target))
}

// New wraps kind with a human-readable message.
func New(kind error, message string) error {
	return &Error{Kind: kind, Message: message}
}

// Wrap wraps kind with a message and a cause.
func Wrap(kind error, message string, cause error) error {
	return &Error{Kind: kind, Message: message, Cause: cause}
}

// WithDetail attaches diagnostic key/value pairs.
func WithDetail(err error, detail map[string]string) error {
	var typed *Error
	if errors.As(err, &typed) {
		cp := *typed
		if cp.Detail == nil {
			cp.Detail = map[string]string{}
		}
		for k, v := range detail {
			cp.Detail[k] = v
		}
		return &cp
	}
	return &Error{Kind: err, Message: err.Error(), Detail: detail}
}

// Detail returns the diagnostic map if err is a typed Centralizer error.
func Detail(err error) map[string]string {
	var typed *Error
	if errors.As(err, &typed) {
		return typed.Detail
	}
	return nil
}
