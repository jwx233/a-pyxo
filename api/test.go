package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"


	"github.com/jwx233s/a-service/pkg/response"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	response.Success(w,"Hello World")
}
