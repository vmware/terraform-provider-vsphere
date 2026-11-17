// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package virtualdevice

import (
	"testing"
)

func TestValidateCdromFields(t *testing.T) {
	cases := []struct {
		name         string
		clientDevice bool
		datastoreID  string
		path         string
		wantErr      string
	}{
		{
			name:         "client device",
			clientDevice: true,
		},
		{
			name:        "datastore iso",
			datastoreID: "datastore-1",
			path:        "iso/ubuntu.iso",
		},
		{
			name:    "empty",
			wantErr: "either client_device or datastore_id and path must be set",
		},
		{
			name:         "client device and datastore",
			clientDevice: true,
			datastoreID:  "datastore-1",
			wantErr:      "cannot have both client_device parameter and ISO file parameters (datastore_id, path) set",
		},
		{
			name:         "client device and path",
			clientDevice: true,
			path:         "iso/ubuntu.iso",
			wantErr:      "cannot have both client_device parameter and ISO file parameters (datastore_id, path) set",
		},
		{
			name:        "datastore without path",
			datastoreID: "datastore-1",
			wantErr:     "either client_device or datastore_id and path must be set",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCdromFields(tc.clientDevice, tc.datastoreID, tc.path)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("expected error %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestValidateCdromBlock(t *testing.T) {
	cases := []struct {
		name    string
		index   int
		raw     interface{}
		wantErr string
	}{
		{
			name:    "nil element",
			index:   0,
			raw:     nil,
			wantErr: "cdrom.0: either client_device or datastore_id and path must be set",
		},
		{
			name:    "empty map",
			index:   0,
			raw:     map[string]interface{}{},
			wantErr: "cdrom.0: either client_device or datastore_id and path must be set",
		},
		{
			name:    "not a map",
			index:   1,
			raw:     "cdrom",
			wantErr: "cdrom.1: either client_device or datastore_id and path must be set",
		},
		{
			name:  "client device",
			index: 0,
			raw:   map[string]interface{}{"client_device": true},
		},
		{
			name:  "datastore iso",
			index: 0,
			raw: map[string]interface{}{
				"datastore_id": "datastore-1",
				"path":         "iso/ubuntu.iso",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCdromBlock(tc.index, tc.raw)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("expected error %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestVerifyVAppTransport(t *testing.T) {
	client := map[string]interface{}{"client_device": true}
	iso := map[string]interface{}{
		"datastore_id": "datastore-1",
		"path":         "iso/ubuntu.iso",
	}
	conflict := map[string]interface{}{
		"client_device": true,
		"datastore_id":  "datastore-1",
	}
	cases := []struct {
		name      string
		blocks    []interface{}
		transport []interface{}
		unknown   map[string]bool
		wantErr   string
	}{
		{
			name:    "nil element",
			blocks:  []interface{}{nil},
			wantErr: "cdrom.0: either client_device or datastore_id and path must be set",
		},
		{
			name:    "empty map",
			blocks:  []interface{}{map[string]interface{}{}},
			wantErr: "cdrom.0: either client_device or datastore_id and path must be set",
		},
		{
			name:      "client device with iso transport",
			blocks:    []interface{}{client},
			transport: []interface{}{"iso"},
		},
		{
			name:      "iso backing with iso-only transport",
			blocks:    []interface{}{iso},
			transport: []interface{}{"iso"},
			wantErr:   "this virtual machine requires a client CDROM device to deliver vApp properties",
		},
		{
			name:    "client device and datastore",
			blocks:  []interface{}{conflict},
			wantErr: "cdrom.0: cannot have both client_device parameter and ISO file parameters (datastore_id, path) set",
		},
		{
			name: "no blocks",
		},
		{
			name:    "second block nil",
			blocks:  []interface{}{client, nil},
			wantErr: "cdrom.1: either client_device or datastore_id and path must be set",
		},
		{
			name:    "unknown datastore id",
			blocks:  []interface{}{map[string]interface{}{"path": "iso/ubuntu.iso"}},
			unknown: map[string]bool{"cdrom.0.datastore_id": true},
		},
		{
			name:      "unknown client device with iso-only transport",
			blocks:    []interface{}{map[string]interface{}{}},
			transport: []interface{}{"iso"},
			unknown:   map[string]bool{"cdrom.0.client_device": true},
		},
		{
			name:      "unknown datastore id with iso-only transport",
			blocks:    []interface{}{map[string]interface{}{"path": "iso/ubuntu.iso"}},
			transport: []interface{}{"iso"},
			unknown:   map[string]bool{"cdrom.0.datastore_id": true},
			wantErr:   "this virtual machine requires a client CDROM device to deliver vApp properties",
		},
		{
			name:    "unknown values with nil block",
			blocks:  []interface{}{nil},
			unknown: map[string]bool{"cdrom.0.client_device": true},
			wantErr: "cdrom.0: either client_device or datastore_id and path must be set",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			known := func(k string) bool { return !tc.unknown[k] }
			err := verifyVAppTransport(tc.blocks, tc.transport, known)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("expected error %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}
