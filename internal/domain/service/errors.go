package service

import "errors"

var (
	ErrPodNotFound              = errors.New("pod not found")
	ErrDeploymentNotFound       = errors.New("deployment not found")
	ErrNoPreviousRevisionsFound = errors.New("no previous revisions")
)
