// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package ssohelper

import (
	"context"
	"log"
	"net/url"
	"sync"

	"github.com/vmware/govmomi/ssoadmin"
	"github.com/vmware/govmomi/sts"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

type SsoClient struct {
	vc       *vim25.Client
	userinfo *url.Userinfo

	mu     sync.Mutex
	client *ssoadmin.Client // cached logged-in client; nil until first use
}

func New(vc *vim25.Client, userinfo *url.Userinfo) *SsoClient {
	return &SsoClient{vc: vc, userinfo: userinfo}
}

func (s *SsoClient) Client(ctx context.Context) (*ssoadmin.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.login(ctx)
}

// login returns the cached admin client, or performs a handshake
// and caches the result.
func (s *SsoClient) login(ctx context.Context) (*ssoadmin.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	log.Printf("[INFO] ssohelper: establishing new SSO admin session (handshake)")
	c, err := ssoadmin.NewClient(ctx, s.vc)
	if err != nil {
		return nil, err
	}

	// This mirrors govmomi's govc/sso/client.go flow.
	tokens, err := sts.NewClient(ctx, s.vc)
	if err != nil {
		return nil, err
	}
	signer, err := tokens.Issue(ctx, sts.TokenRequest{
		Certificate: s.vc.Certificate(),
		Userinfo:    s.userinfo,
	})
	if err != nil {
		return nil, err
	}

	header := soap.Header{Security: signer}
	if err := c.Login(c.WithHeader(ctx, header)); err != nil {
		return nil, err
	}

	s.client = c
	return c, nil
}
