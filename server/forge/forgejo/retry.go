// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package forgejo

import (
	"fmt"
	"time"
)

func Retry[T any](fun func() (T, bool, string, error), tries int) (T, int, string, error) {
	reasons := make([]string, 0, tries)
	fmtReasons := func() string {
		return fmt.Sprintf("Retry reasons %v", reasons)
	}
	for i := 0; i < tries; i++ {
		result, retry, reason, err := fun()
		if err != nil {
			return result, i, fmtReasons(), err
		}
		if !retry {
			return result, i, fmtReasons(), nil
		}
		reasons = append(reasons, reason)
		<-time.After(1 * time.Second)
	}
	var result T
	return result, tries, fmtReasons(), fmt.Errorf(fmtReasons())
}
