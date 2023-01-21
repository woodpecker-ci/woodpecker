// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

type VisibleType string

const (
	VisibleTypePublic  VisibleType = "public"
	VisibleTypeLimited VisibleType = "limited"
	VisibleTypePrivate VisibleType = "private"
)
