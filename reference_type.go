// Copyright 2018-2024 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package opcua

import (
	"github.com/otfabric/opcua/id"
	"github.com/otfabric/opcua/ua"
)

// ReferenceTypeDisplayName returns a display string for a reference type NodeID.
// For well-known reference types in namespace 0 (e.g. HasComponent, Organizes),
// it returns the standard name; otherwise it returns the NodeID string.
// Use when displaying the reference type column in browse refs or similar UIs.
// Returns the empty string if refTypeID is nil.
func ReferenceTypeDisplayName(refTypeID *ua.NodeID) string {
	if refTypeID == nil {
		return ""
	}
	if refTypeID.Namespace() == 0 {
		if name := id.ReferenceTypeName(refTypeID.IntID()); name != "" {
			return name
		}
	}
	return refTypeID.String()
}
