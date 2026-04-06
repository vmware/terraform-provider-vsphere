// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProviderSettings holds resolved provider configuration after applying
// explicit Terraform values, environment variables, and defaults. This is
// shared by the SDK Configure path and the plugin-framework Configure path
// (terraform-plugin-mux requires matching effective configuration).
type ProviderSettings struct {
	User               string
	Password           string
	VSphereServer      string
	VCenterServer      string
	AllowUnverifiedSSL bool
	ClientDebug        bool
	ClientDebugPathRun string
	ClientDebugPath    string
	PersistSession     bool
	VimSessionPath     string
	RestSessionPath    string
	VimKeepAlive       int
	APITimeoutMins     int
}

func defaultVimSessionPath() string {
	return filepath.Join(os.Getenv("HOME"), ".govmomi", "sessions")
}

func defaultRestSessionPath() string {
	return filepath.Join(os.Getenv("HOME"), ".govmomi", "rest_sessions")
}

// ProviderSettingsFromResourceData resolves settings from SDK ResourceData using
// the same precedence as the former schema DefaultFunc values: explicit config,
// then environment variable, then default.
func ProviderSettingsFromResourceData(d *schema.ResourceData) (ProviderSettings, error) {
	var s ProviderSettings
	s.User = resolveStringFromSDK(d, "user", "VSPHERE_USER", "")
	s.Password = resolveStringFromSDK(d, "password", "VSPHERE_PASSWORD", "")
	s.VSphereServer = resolveStringFromSDK(d, "vsphere_server", "VSPHERE_SERVER", "")
	s.VCenterServer = resolveStringFromSDK(d, "vcenter_server", "VSPHERE_VCENTER", "")
	s.AllowUnverifiedSSL = resolveBoolFromSDK(d, "allow_unverified_ssl", "VSPHERE_ALLOW_UNVERIFIED_SSL", false)
	s.ClientDebug = resolveBoolFromSDK(d, "client_debug", "VSPHERE_CLIENT_DEBUG", false)
	s.ClientDebugPathRun = resolveStringFromSDK(d, "client_debug_path_run", "VSPHERE_CLIENT_DEBUG_PATH_RUN", "")
	s.ClientDebugPath = resolveStringFromSDK(d, "client_debug_path", "VSPHERE_CLIENT_DEBUG_PATH", "")
	s.PersistSession = resolveBoolFromSDK(d, "persist_session", "VSPHERE_PERSIST_SESSION", false)
	s.VimSessionPath = resolveStringFromSDK(d, "vim_session_path", "VSPHERE_VIM_SESSION_PATH", defaultVimSessionPath())
	s.RestSessionPath = resolveStringFromSDK(d, "rest_session_path", "VSPHERE_REST_SESSION_PATH", defaultRestSessionPath())
	s.VimKeepAlive = resolveIntFromSDK(d, "vim_keep_alive", "VSPHERE_VIM_KEEP_ALIVE", 10)
	s.APITimeoutMins = resolveIntFromSDK(d, "api_timeout", "VSPHERE_API_TIMEOUT", 5)
	return s, s.validate()
}

func resolveStringFromSDK(d *schema.ResourceData, key, envKey, defaultVal string) string {
	if v, ok := d.GetOk(key); ok {
		if s, ok := v.(string); ok {
			if s != "" {
				return s
			}
		}
	}
	if env := os.Getenv(envKey); env != "" {
		return env
	}
	return defaultVal
}

func resolveBoolFromSDK(d *schema.ResourceData, key, envKey string, defaultVal bool) bool {
	if v, ok := d.GetOkExists(key); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	if env := os.Getenv(envKey); env != "" {
		if b, err := strconv.ParseBool(env); err == nil {
			return b
		}
	}
	return defaultVal
}

func resolveIntFromSDK(d *schema.ResourceData, key, envKey string, defaultVal int) int {
	if v, ok := d.GetOk(key); ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	if env := os.Getenv(envKey); env != "" {
		if i, err := strconv.Atoi(env); err == nil {
			return i
		}
	}
	return defaultVal
}

func (s ProviderSettings) validate() error {
	if s.User == "" {
		return fmt.Errorf("user must be set in the provider block or via VSPHERE_USER")
	}
	if s.Password == "" {
		return fmt.Errorf("password must be set in the provider block or via VSPHERE_PASSWORD")
	}
	server := s.VSphereServer
	if server == "" {
		server = s.VCenterServer
	}
	if server == "" {
		return fmt.Errorf("one of vsphere_server or vcenter_server must be provided (or set VSPHERE_SERVER / VSPHERE_VCENTER)")
	}
	return nil
}

// EffectiveVSphereServer returns the server hostname, preferring vsphere_server over vcenter_server.
func (s ProviderSettings) EffectiveVSphereServer() string {
	if s.VSphereServer != "" {
		return s.VSphereServer
	}
	return s.VCenterServer
}

// ProviderFrameworkConfig carries optional plugin-framework attribute values.
// A nil field means the practitioner did not set the attribute (null).
type ProviderFrameworkConfig struct {
	User               *string
	Password           *string
	VSphereServer      *string
	VCenterServer      *string
	AllowUnverifiedSSL *bool
	ClientDebug        *bool
	ClientDebugPathRun *string
	ClientDebugPath    *string
	PersistSession     *bool
	VimSessionPath     *string
	RestSessionPath    *string
	VimKeepAlive       *int
	APITimeoutMins     *int
}

// ProviderSettingsFromFramework resolves settings using the same rules as
// ProviderSettingsFromResourceData for use with terraform-plugin-framework.
func ProviderSettingsFromFramework(in ProviderFrameworkConfig) (ProviderSettings, error) {
	var s ProviderSettings
	s.User = resolveStringPtr(in.User, "VSPHERE_USER", "")
	s.Password = resolveStringPtr(in.Password, "VSPHERE_PASSWORD", "")
	s.VSphereServer = resolveStringPtr(in.VSphereServer, "VSPHERE_SERVER", "")
	s.VCenterServer = resolveStringPtr(in.VCenterServer, "VSPHERE_VCENTER", "")
	s.AllowUnverifiedSSL = resolveBoolPtr(in.AllowUnverifiedSSL, "VSPHERE_ALLOW_UNVERIFIED_SSL", false)
	s.ClientDebug = resolveBoolPtr(in.ClientDebug, "VSPHERE_CLIENT_DEBUG", false)
	s.ClientDebugPathRun = resolveStringPtr(in.ClientDebugPathRun, "VSPHERE_CLIENT_DEBUG_PATH_RUN", "")
	s.ClientDebugPath = resolveStringPtr(in.ClientDebugPath, "VSPHERE_CLIENT_DEBUG_PATH", "")
	s.PersistSession = resolveBoolPtr(in.PersistSession, "VSPHERE_PERSIST_SESSION", false)
	s.VimSessionPath = resolveStringPtr(in.VimSessionPath, "VSPHERE_VIM_SESSION_PATH", defaultVimSessionPath())
	s.RestSessionPath = resolveStringPtr(in.RestSessionPath, "VSPHERE_REST_SESSION_PATH", defaultRestSessionPath())
	s.VimKeepAlive = resolveIntPtr(in.VimKeepAlive, "VSPHERE_VIM_KEEP_ALIVE", 10)
	s.APITimeoutMins = resolveIntPtr(in.APITimeoutMins, "VSPHERE_API_TIMEOUT", 5)
	return s, s.validate()
}

func resolveStringPtr(val *string, envKey, defaultVal string) string {
	if val != nil && *val != "" {
		return *val
	}
	if env := os.Getenv(envKey); env != "" {
		return env
	}
	if val != nil {
		return *val
	}
	return defaultVal
}

func resolveBoolPtr(val *bool, envKey string, defaultVal bool) bool {
	if val != nil {
		return *val
	}
	if env := os.Getenv(envKey); env != "" {
		if b, err := strconv.ParseBool(env); err == nil {
			return b
		}
	}
	return defaultVal
}

func resolveIntPtr(val *int, envKey string, defaultVal int) int {
	if val != nil {
		return *val
	}
	if env := os.Getenv(envKey); env != "" {
		if i, err := strconv.Atoi(env); err == nil {
			return i
		}
	}
	return defaultVal
}
