// Copyright 2018 Gin Core Team.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"net/http"

	"google.golang.org/protobuf/proto"
)

// ProtoBuf contains the given interface object.
type ProtoBuf struct {
	Status  int
	Headers map[string]string
	Data    any
}

var protobufContentType = []string{"application/x-protobuf"}

// Render (ProtoBuf) marshals the given interface object and writes data with custom ContentType.
func (r ProtoBuf) Render(w http.ResponseWriter) error {
	writeHeaderCode(w, r.Status)
	writeHeaders(w, r.Headers)
	r.WriteContentType(w)

	bytes, err := proto.Marshal(r.Data.(proto.Message))
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	return err
}

// WriteContentType (ProtoBuf) writes ProtoBuf ContentType.
func (r ProtoBuf) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, protobufContentType)
}
