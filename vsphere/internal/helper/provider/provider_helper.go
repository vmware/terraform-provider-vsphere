// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"time"
)

// DefaultAPITimeout is a default timeout value that is passed to functions
// requiring contexts, and other various waiters.
const DefaultAPITimeout = time.Minute * 5

func Error(id string, function string, err error) error {
	return fmt.Errorf("%s: RESOURCE (%s), ACTION (%s)", err, id, function)
}
