// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/contentlibrary"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereContentLibrary_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereContentLibraryConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "description", regexp.MustCompile("Library Description"),
					),
					testAccResourceVSphereContentLibraryDescription(regexp.MustCompile("Library Description")),
					testAccResourceVSphereContentLibraryName(regexp.MustCompile("testacc_content_library")),
				),
			},
			{
				ResourceName:      "vsphere_content_library.library",
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testAccResourceVSphereContentLibraryConfig(),
			},
		},
	})
}

func TestAccResourceVSphereContentLibrary_subscribed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testaccresourcevspherecontentlibraryconfigSubscribed(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "description", regexp.MustCompile("Library Description"),
					),
					testAccResourceVSphereContentLibraryDescription(regexp.MustCompile("Library Description")),
					testAccResourceVSphereContentLibraryName(regexp.MustCompile("testacc_subscribed")),
				),
			},
		},
	})
}
func TestAccResourceVSphereContentLibrary_authenticated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testaccresourcevspherecontentlibraryconfigAuthenticated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library.library", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
				),
			},
		},
	})
}

func testAccResourceVSphereContentLibraryDescription(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibrary(s, "library")
		if err != nil {
			return err
		}
		if !expected.MatchString(*library.Description) {
			return fmt.Errorf("content Library description does not match. expected: %s, got %v", expected.String(), library.Description)
		}
		return nil
	}
}

func testAccResourceVSphereContentLibraryName(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibrary(s, "library")
		if err != nil {
			return err
		}
		if !expected.MatchString(library.Name) {
			return fmt.Errorf("content Library name does not match. expected: %s, got %s", expected.String(), library.Name)
		}
		return nil
	}
}

func testaccresourcevspherecontentlibraryconfigBase() string {
	return testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost3(), testhelper.ConfigDataRootDS1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1())
}

func testaccresourcevspherecontentlibraryconfigAuthenticated() string {
	return fmt.Sprintf(`
%s

resource "vsphere_content_library" "library_published" {
  name            = "testacc_published"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
  publication {
    authentication_method = "BASIC"
    username              = "vcsp"
    password              = "Password123!"
    published             = true
  }
}

resource "vsphere_content_library" "library" {
  name            = "testacc_subscribed"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
  subscription {
    authentication_method = "BASIC"
    username              = "vcsp"
    password              = "Password123!"
    subscription_url      = vsphere_content_library.library_published.publication.0.publish_url
  }
}
`, testaccresourcevspherecontentlibraryconfigBase())
}
func testaccresourcevspherecontentlibraryconfigSubscribed() string {
	return fmt.Sprintf(`
%s

resource "vsphere_content_library" "library_published" {
  name            = "testacc_published"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
  publication {
    published = true
  }
}

resource "vsphere_content_library" "library" {
  name            = "testacc_subscribed"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
  subscription {
    subscription_url = vsphere_content_library.library_published.publication.0.publish_url
  }
}
`, testaccresourcevspherecontentlibraryconfigBase())
}

func testAccResourceVSphereContentLibraryConfig() string {
	return fmt.Sprintf(`
%s

resource "vsphere_content_library" "library" {
  name            = "testacc_content_library"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}
`, testaccresourcevspherecontentlibraryconfigBase())
}

func testAccResourceVSphereContentLibraryCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetContentLibrary(s, "library")
		if err != nil {
			missingState, _ := regexp.MatchString("not found in state", err.Error())
			missingVSphere, _ := regexp.MatchString("404 Not Found", err.Error())
			if missingState && !expected || missingVSphere && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected Content Library to be missing")
		}
		return nil
	}
}

const testContentLibrarySubscriptionThumbprint = "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD"

func testContentLibrarySubscriptionData(t *testing.T, subscription map[string]interface{}) *schema.ResourceData {
	raw := map[string]interface{}{
		"name":            "library",
		"storage_backing": []interface{}{"datastore-1"},
	}
	if subscription != nil {
		raw["subscription"] = []interface{}{subscription}
	}
	return schema.TestResourceDataRaw(t, resourceVSphereContentLibrary().Schema, raw)
}

func TestContentLibraryExpandSubscription(t *testing.T) {
	t.Run("no subscription", func(t *testing.T) {
		d := testContentLibrarySubscriptionData(t, nil)
		if actual := contentlibrary.ExpandSubscription(d); actual != nil {
			t.Fatalf("expected nil subscription, got %#v", actual)
		}
	})

	t.Run("ssl thumbprint", func(t *testing.T) {
		d := testContentLibrarySubscriptionData(t, map[string]interface{}{
			"subscription_url": "https://publisher.example.com/lib.json",
			"ssl_thumbprint":   testContentLibrarySubscriptionThumbprint,
		})
		actual := contentlibrary.ExpandSubscription(d)
		if actual == nil {
			t.Fatal("expected subscription, got nil")
		}
		if actual.SslThumbprint != testContentLibrarySubscriptionThumbprint {
			t.Fatalf("expected ssl thumbprint %q, got %q", testContentLibrarySubscriptionThumbprint, actual.SslThumbprint)
		}
		if actual.SubscriptionURL != "https://publisher.example.com/lib.json" {
			t.Fatalf("expected subscription url %q, got %q", "https://publisher.example.com/lib.json", actual.SubscriptionURL)
		}
	})
}

func TestContentLibraryFlattenSubscriptionThumbprint(t *testing.T) {
	cases := []struct {
		name       string
		configured string
		returned   string
		expected   string
	}{
		{
			name:     "not configured",
			returned: testContentLibrarySubscriptionThumbprint,
			expected: testContentLibrarySubscriptionThumbprint,
		},
		{
			name:       "configured value is kept",
			configured: "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd",
			returned:   testContentLibrarySubscriptionThumbprint,
			expected:   "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd",
		},
		{
			name:       "not returned",
			configured: testContentLibrarySubscriptionThumbprint,
			expected:   testContentLibrarySubscriptionThumbprint,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := testContentLibrarySubscriptionData(t, map[string]interface{}{
				"subscription_url": "https://publisher.example.com/lib.json",
				"ssl_thumbprint":   tc.configured,
			})
			err := contentlibrary.FlattenSubscription(d, &library.Subscription{
				AuthenticationMethod: "NONE",
				SubscriptionURL:      "https://publisher.example.com/lib.json",
				SslThumbprint:        tc.returned,
			})
			if err != nil {
				t.Fatalf("error flattening subscription: %s", err)
			}
			if actual := d.Get("subscription.0.ssl_thumbprint").(string); actual != tc.expected {
				t.Fatalf("expected ssl thumbprint %q, got %q", tc.expected, actual)
			}
		})
	}
}
