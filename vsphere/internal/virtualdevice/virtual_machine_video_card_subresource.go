// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package virtualdevice

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/structure"
)

func VideoCardSubresourceSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"num_displays": {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Number of supported displays",
			ValidateFunc: validation.IntBetween(1, 10),
		},
		"total_video_memory": {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Video RAM size in megabytes",
			ValidateFunc: validation.IntBetween(2, 256),
		},
		"graphics_3d": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"renderer": {
						Type:         schema.TypeString,
						Optional:     true,
						Description:  "Indicates how the virtual device renders 3D graphics",
						ValidateFunc: validation.StringInSlice([]string{"hardware", "software", "automatic"}, false),
					},
					"memory": {
						Type:         schema.TypeInt,
						Optional:     true,
						Description:  "The amount of dedicated graphics memory in megabytes",
						ValidateFunc: validation.IntBetween(2, 4096),
					},
				},
			},
		},
	}
	structure.MergeSchema(s, subresourceSchema())
	return s
}

type VideoCardSubresource struct {
	*Subresource
}

func NewVideoCardSubresource(client *govmomi.Client, rdd resourceDataDiff, d, old map[string]interface{}, idx int) *VideoCardSubresource {
	sr := &VideoCardSubresource{
		Subresource: &Subresource{
			schema:  VideoCardSubresourceSchema(),
			client:  client,
			srtype:  subresourceTypeVideoCard,
			data:    d,
			olddata: old,
			rdd:     rdd,
		},
	}
	sr.Index = idx
	return sr
}

func VideoCardApplyOperation(d *schema.ResourceData, c *govmomi.Client, l object.VirtualDeviceList) (object.VirtualDeviceList, []types.BaseVirtualDeviceConfigSpec, error) {
	o, n := d.GetChange(subresourceTypeVideoCard)
	ods := o.([]interface{})
	nds := n.([]interface{})

	// can have only one virtual video card
	var odsmap map[string]interface{}
	var ndsmap map[string]interface{}
	if len(ods) > 0 {
		odsmap = ods[0].(map[string]interface{})
	}
	if len(nds) > 0 {
		ndsmap = nds[0].(map[string]interface{})
	}

	var specs []types.BaseVirtualDeviceConfigSpec

	if ndsmap != nil {
		if d.IsNewResource() {
			// create
			r := NewVideoCardSubresource(c, d, ndsmap, odsmap, 0)
			spec, err := r.Create(l)
			if err != nil {
				return nil, nil, fmt.Errorf("%s: %s", r.Addr(), err)
			}
			l = applyDeviceChange(l, spec)
			specs = append(specs, spec...)
		} else {
			// update
			r := NewVideoCardSubresource(c, d, ndsmap, odsmap, 0)
			spec, err := r.Update(l)
			if err != nil {
				return nil, nil, fmt.Errorf("%s: %s", r.Addr(), err)
			}
			l = applyDeviceChange(l, spec)
			specs = append(specs, spec...)
		}
	}

	return l, specs, nil
}

func VideoCardRefreshOperation(d *schema.ResourceData, c *govmomi.Client, l object.VirtualDeviceList) error {
	curSet := d.Get(subresourceTypeVideoCard).([]interface{})

	if len(curSet) > 0 {
		r := NewVideoCardSubresource(c, d, make(map[string]interface{}), curSet[0].(map[string]interface{}), 0)
		if err := r.Read(l); err != nil {
			return err
		}
		return d.Set(subresourceTypeVideoCard, []interface{}{r.Data()})
	}

	return nil
}

func ReadVideoCardForDataSource(l object.VirtualDeviceList) ([]map[string]interface{}, error) {
	device, err := findVideoCard(l)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	result["num_displays"] = device.NumDisplays
	result["total_video_memory"] = device.VideoRamSizeInKB / 1024
	if device.Enable3DSupport != nil && *device.Enable3DSupport {
		m := make(map[string]interface{})
		m["renderer"] = device.Use3dRenderer
		m["memory"] = device.GraphicsMemorySizeInKB / 1024
		result["graphics_3d"] = m
	}

	return []map[string]interface{}{result}, nil
}

func (r *VideoCardSubresource) Create(l object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error) {
	var specs []types.BaseVirtualDeviceConfigSpec
	device, err := findVideoCard(l)
	if err != nil {
		return nil, err
	}
	r.mapProperties(device)
	spec, err := object.VirtualDeviceList{device}.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return nil, err
	}
	specs = append(specs, spec...)
	return specs, nil
}

func (r *VideoCardSubresource) Read(l object.VirtualDeviceList) error {
	device, err := findVideoCard(l)
	if err != nil {
		return err
	}

	r.Set("num_displays", device.NumDisplays)
	r.Set("total_video_memory", device.VideoRamSizeInKB/1024)
	if device.Enable3DSupport != nil && *device.Enable3DSupport {
		m := make(map[string]interface{})
		m["renderer"] = device.Use3dRenderer
		m["memory"] = device.GraphicsMemorySizeInKB / 1024
		r.Set("graphics_3d", m)
	}
	r.Set("key", device.Key)
	return nil
}

func (r *VideoCardSubresource) Update(l object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error) {
	device, err := findVideoCard(l)
	if err != nil {
		return nil, err
	}

	r.mapProperties(device)
	r.SetRestart("<video_card> update")
	var specs []types.BaseVirtualDeviceConfigSpec
	spec, err := object.VirtualDeviceList{device}.ConfigSpec(types.VirtualDeviceConfigSpecOperationEdit)
	if err != nil {
		return nil, err
	}
	specs = append(specs, spec...)
	return specs, nil
}

func (r *VideoCardSubresource) mapProperties(videoCard *types.VirtualMachineVideoCard) {
	videoCard.NumDisplays = int32(r.Get("num_displays").(int))
	videoCard.VideoRamSizeInKB = int64(r.Get("total_video_memory").(int) * 1024)

	if graphics3d := r.Get("graphics_3d").([]interface{}); len(graphics3d) > 0 {
		videoCard.Enable3DSupport = structure.BoolPtr(true)

		graphics3dData := graphics3d[0].(map[string]interface{})

		if graphicsMemory := graphics3dData["memory"].(int); graphicsMemory > 0 {
			videoCard.GraphicsMemorySizeInKB = int64(graphicsMemory * 1024)
		}

		if renderer := graphics3dData["renderer"].(string); renderer != "" {
			videoCard.Use3dRenderer = renderer
		}
	}
}

func findVideoCard(l object.VirtualDeviceList) (*types.VirtualMachineVideoCard, error) {
	devices := l.Select(func(device types.BaseVirtualDevice) bool {
		if _, ok := device.(*types.VirtualMachineVideoCard); ok {
			return true
		}
		return false
	})

	if len(devices) == 0 {
		return nil, fmt.Errorf("no video card found")
	}

	return devices[0].(*types.VirtualMachineVideoCard), nil
}
