package internal

import "net/http"

type BadRequestError struct {
	Message string
}

func (e BadRequestError) Error() string {
	return e.Message
}

func (e BadRequestError) HTTPStatusCode() int {
	return http.StatusBadRequest
}

type InternalServerError struct {
	Message string
}

func (e InternalServerError) Error() string {
	return e.Message
}

func (e InternalServerError) HTTPStatusCode() int {
	return http.StatusInternalServerError
}

type ForbiddenError struct {
	Message string
}

func (e ForbiddenError) Error() string {
	return e.Message
}

func (e ForbiddenError) HTTPStatusCode() int {
	return http.StatusForbidden
}

type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	return e.Message
}

func (e ConflictError) HTTPStatusCode() int {
	return http.StatusConflict
}

type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}

func (e NotFoundError) HTTPStatusCode() int {
	return http.StatusNotFound
}

type UnauthorizedError struct {
	Message string
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

func (e UnauthorizedError) HTTPStatusCode() int {
	return http.StatusUnauthorized
}
