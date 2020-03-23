package handler

import (
	"github.com/patdeg/go-appengine/common"
	"github.com/patdeg/go-appengine/track"
	"github.com/patdeg/go-appengine/auth"
	"net/http"
)

type Handler struct {
	w http.ResponseWriter
	r *http.Request
}
