// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

// func GetPlayerBody returns Players info sent from http.Request
func GetPlayerBody(r *http.Request, p *Players) error {
	// get "Content-Length" from http header
	cl := r.Header.Get("Content-Length")
	if len(cl) == 0 {
		return errors.New("no parameters")
	}
	length, err := strconv.Atoi(cl)
	if err != nil {
		return err
	}

	// get request body
	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		return err
	}

	// parse request body to json
	err = json.Unmarshal(body, &p)
	if err != nil {
		return err
	}
	return nil
}

// func LogSuccessResponse logs success log to stderr and http.ResponseWriter
func LogSuccessResponse(w http.ResponseWriter, format string, v ...interface{}) {
	log.Printf(format, v...)
	fmt.Fprintf(w, format, v...)
}

// func LogErrorResponse logs error log to stderr and http.ResponseWriter
func LogErrorResponse(err error, w http.ResponseWriter) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
