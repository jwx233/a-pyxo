package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jwx233s/a-service/pkg/db"
	"github.com/jwx233s/a-service/pkg/response"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	response.Error(w,"Hello World",200)
}
