// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/contentlibrary"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/provider"
)

const (
	contentLibrarySortByLastModified = "last_modified_time"
	contentLibrarySortByCreation     = "creation_time"
	contentLibrarySortByName         = "name"
	contentLibrarySortOrderAsc       = "asc"
	contentLibrarySortOrderDesc      = "desc"
)

func dataSourceVSphereContentLibraryItems() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereContentLibraryItemsRead,
		Schema: map[string]*schema.Schema{
			"library_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the content library to search.",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "Filter items by name (case-insensitive) of the content library item. Mutually exclusive with name_regex.",
				ConflictsWith: []string{"name_regex"},
			},
			"name_regex": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "Filter items by name using a regular expression. Mutually exclusive with name.",
				ValidateFunc:  validation.StringIsValidRegExp,
				ConflictsWith: []string{"name"},
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Filter items by type of the content library item (e.g. ovf, iso, vm-template). If omitted, all items matching the name filter are returned.",
			},
			"sort_by": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      contentLibrarySortByLastModified,
				Description:  "Sort items by the given field. Accepted values: last_modified_time, creation_time, name. Default is last_modified_time.",
				ValidateFunc: validation.StringInSlice([]string{contentLibrarySortByLastModified, contentLibrarySortByCreation, contentLibrarySortByName}, false),
			},
			"sort_order": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      contentLibrarySortOrderDesc,
				Description:  "Sort order for the results. Accepted values: asc, desc. Default is desc (most recent first).",
				ValidateFunc: validation.StringInSlice([]string{contentLibrarySortOrderAsc, contentLibrarySortOrderDesc}, false),
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "Limit the number of items returned after sorting. If omitted, all matching items are returned.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"items": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of content library items matching the search criteria.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The UUID of the content library item.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the content library item.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the content library item.",
						},
						"creation_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The creation time of the content library item (RFC3339).",
						},
						"last_modified_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The last modification time of the content library item (RFC3339).",
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereContentLibraryItemsRead(d *schema.ResourceData, meta interface{}) error {
	rc := meta.(*Client).restClient
	libraryID := d.Get("library_id").(string)
	name := d.Get("name").(string)
	nameRegex := d.Get("name_regex").(string)
	itemType := d.Get("type").(string)
	sortBy := d.Get("sort_by").(string)
	sortOrder := d.Get("sort_order").(string)

	items, err := contentlibrary.ItemsFromCriteria(rc, libraryID, name, nameRegex, itemType)
	if err != nil {
		return provider.Error(libraryID, "dataSourceVSphereContentLibraryItemsRead", err)
	}

	// Sort items client-side
	sort.Slice(items, func(i, j int) bool {
		asc := sortOrder == contentLibrarySortOrderAsc
		if sortBy == contentLibrarySortByName {
			if asc {
				return items[i].Name < items[j].Name
			}
			return items[i].Name > items[j].Name
		}
		ti := sortTime(items[i], sortBy)
		tj := sortTime(items[j], sortBy)
		if asc {
			return ti.Before(tj)
		}
		return ti.After(tj)
	})

	// Apply limit if set
	if v, ok := d.GetOk("limit"); ok {
		limit := v.(int)
		if limit < len(items) {
			items = items[:limit]
		}
	}

	flat := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		entry := map[string]interface{}{
			"id":   item.ID,
			"name": item.Name,
			"type": item.Type,
		}
		if item.CreationTime != nil {
			entry["creation_time"] = item.CreationTime.Format(time.RFC3339)
		}
		if item.LastModifiedTime != nil {
			entry["last_modified_time"] = item.LastModifiedTime.Format(time.RFC3339)
		}
		flat = append(flat, entry)
	}
	if err := d.Set("items", flat); err != nil {
		return err
	}

	// Derive a stable ID from the search parameters
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%s|%s|%s|%s", libraryID, name, nameRegex, itemType)
	d.SetId(fmt.Sprintf("%x", h.Sum(nil)))
	return nil
}

// sortTime returns the relevant time field for sorting, falling back to zero time if nil.
func sortTime(item *library.Item, sortBy string) time.Time {
	if sortBy == contentLibrarySortByCreation {
		if item.CreationTime != nil {
			return *item.CreationTime
		}
		return time.Time{}
	}
	if item.LastModifiedTime != nil {
		return *item.LastModifiedTime
	}
	return time.Time{}
}
