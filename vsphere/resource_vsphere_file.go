// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/soap"
)

// file represents a datastore file operation, including paths, datacenters, and configuration options.
type file struct {
	sourceDatacenter  string
	datacenter        string
	sourceDatastore   string
	datastore         string
	sourceFile        string
	destinationFile   string
	createDirectories bool
	copyFile          bool
}

// resourceVSphereFile defines a resource for managing files or virtual disks on a datastore.
func resourceVSphereFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereFileCreate,
		Read:   resourceVSphereFileRead,
		Update: resourceVSphereFileUpdate,
		Delete: resourceVSphereFileDelete,

		Schema: map[string]*schema.Schema{
			"datacenter": {
				Type:        schema.TypeString,
				Description: "The name of a datacenter to which the file will be uploaded.",
				Optional:    true,
			},
			"source_datacenter": {
				Type:        schema.TypeString,
				Description: "The name of a datacenter from which the file will be copied.",
				Optional:    true,
				ForceNew:    true,
			},
			"datastore": {
				Type:        schema.TypeString,
				Description: "The name of the datastore to which to upload the file.",
				Required:    true,
			},
			"source_datastore": {
				Type:        schema.TypeString,
				Description: "The name of the datastore from which file will be copied.",
				Optional:    true,
				ForceNew:    true,
			},
			"source_file": {
				Type:        schema.TypeString,
				Description: "The path to the file being uploaded from or copied.",
				Required:    true,
				ForceNew:    true,
			},
			"destination_file": {
				Type:        schema.TypeString,
				Description: "The path to where the file should be uploaded or copied to on the destination datastore.",
				Required:    true,
			},
			"create_directories": {
				Type:        schema.TypeBool,
				Description: "Specifies whether to create the parent directories of the destination file if they do not exist.",
				Optional:    true,
			},
		},
	}
}

func resourceVSphereFileCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] creating file: %#v", d)
	client := meta.(*Client).vimClient

	f := file{}

	if v, ok := d.GetOk("source_datacenter"); ok {
		f.sourceDatacenter = v.(string)
		f.copyFile = true
	}

	if v, ok := d.GetOk("datacenter"); ok {
		f.datacenter = v.(string)
	}

	if v, ok := d.GetOk("source_datastore"); ok {
		f.sourceDatastore = v.(string)
		f.copyFile = true
	}

	if v, ok := d.GetOk("datastore"); ok {
		f.datastore = v.(string)
	} else {
		return fmt.Errorf("datastore argument is required")
	}

	if v, ok := d.GetOk("source_file"); ok {
		f.sourceFile = v.(string)
	} else {
		return fmt.Errorf("source_file argument is required")
	}

	if v, ok := d.GetOk("destination_file"); ok {
		f.destinationFile = v.(string)
	} else {
		return fmt.Errorf("destination_file argument is required")
	}

	if v, ok := d.GetOk("create_directories"); ok {
		f.createDirectories = v.(bool)
	}

	err := createFile(client, &f)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("[%v] %v/%v", f.datastore, f.datacenter, f.destinationFile))
	log.Printf("[INFO] Created file: %s", f.destinationFile)

	return resourceVSphereFileRead(d, meta)
}

// resourceVSphereFileRead retrieves the state of the datastore file and updates the resource data.
func resourceVSphereFileRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] reading file: %#v", d)
	f := file{}

	if v, ok := d.GetOk("source_datacenter"); ok {
		f.sourceDatacenter = v.(string)
	}

	if v, ok := d.GetOk("datacenter"); ok {
		f.datacenter = v.(string)
	}

	if v, ok := d.GetOk("source_datastore"); ok {
		f.sourceDatastore = v.(string)
	}

	if v, ok := d.GetOk("datastore"); ok {
		f.datastore = v.(string)
	} else {
		return fmt.Errorf("datastore argument is required")
	}

	if v, ok := d.GetOk("source_file"); ok {
		f.sourceFile = v.(string)
	} else {
		return fmt.Errorf("source_file argument is required")
	}

	if v, ok := d.GetOk("destination_file"); ok {
		f.destinationFile = v.(string)
	} else {
		return fmt.Errorf("destination_file argument is required")
	}

	client := meta.(*Client).vimClient
	finder := find.NewFinder(client.Client, true)

	dc, err := finder.Datacenter(context.TODO(), f.datacenter)
	if err != nil {
		return fmt.Errorf("error finding datacenter %s", err)
	}
	finder = finder.SetDatacenter(dc)

	ds, err := getDatastore(finder, f.datastore)
	if err != nil {
		return fmt.Errorf("error finding datastore %s", err)
	}

	_, err = ds.Stat(context.TODO(), f.destinationFile)
	if err != nil {
		log.Printf("[DEBUG] resourceVSphereFileRead - stat failed on: %v", f.destinationFile)
		d.SetId("")

		var notFoundError *object.DatastoreNoSuchFileError
		ok := errors.As(err, &notFoundError)
		if !ok {
			return err
		}
	}

	return nil
}

// resourceVSphereFileUpdate handles updating a file resource when attributes are modified.
func resourceVSphereFileUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] updating file: %#v", d)

	if d.HasChange("destination_file") || d.HasChange("datacenter") || d.HasChange("datastore") {
		// File needs to be moved, get old and new destination changes
		var oldDataceneter, newDatacenter, oldDatastore, newDatastore, oldDestinationFile, newDestinationFile string
		if d.HasChange("datacenter") {
			tmpOldDataceneter, tmpNewDatacenter := d.GetChange("datacenter")
			oldDataceneter = tmpOldDataceneter.(string)
			newDatacenter = tmpNewDatacenter.(string)
		} else if v, ok := d.GetOk("datacenter"); ok {
			oldDataceneter = v.(string)
			newDatacenter = oldDataceneter
		}
		if d.HasChange("datastore") {
			tmpOldDatastore, tmpNewDatastore := d.GetChange("datastore")
			oldDatastore = tmpOldDatastore.(string)
			newDatastore = tmpNewDatastore.(string)
		} else {
			oldDatastore = d.Get("datastore").(string)
			newDatastore = oldDatastore
		}
		if d.HasChange("destination_file") {
			tmpOldDestinationFile, tmpNewDestinationFile := d.GetChange("destination_file")
			oldDestinationFile = tmpOldDestinationFile.(string)
			newDestinationFile = tmpNewDestinationFile.(string)
		} else {
			oldDestinationFile = d.Get("destination_file").(string)
			newDestinationFile = oldDestinationFile
		}

		// Get old and new datacenter and datastore.
		client := meta.(*Client).vimClient
		dcOld, err := getDatacenter(client, oldDataceneter)
		if err != nil {
			return err
		}
		dcNew, err := getDatacenter(client, newDatacenter)
		if err != nil {
			return err
		}
		finder := find.NewFinder(client.Client, true)
		finder = finder.SetDatacenter(dcOld)
		dsOld, err := getDatastore(finder, oldDatastore)
		if err != nil {
			return fmt.Errorf("error finding old datastore %s", err)
		}
		finder = finder.SetDatacenter(dcNew)
		dsNew, err := getDatastore(finder, newDatastore)
		if err != nil {
			return fmt.Errorf("error finding new datastore %s", err)
		}

		// Move file between old/new datacenter, datastore, and path.
		fm := object.NewFileManager(client.Client)
		task, err := fm.MoveDatastoreFile(context.TODO(), dsOld.Path(oldDestinationFile), dcOld, dsNew.Path(newDestinationFile), dcNew, true)
		if err != nil {
			return err
		}
		_, err = task.WaitForResultEx(context.TODO(), nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// resourceVSphereFileDelete deletes a file or virtual disk from a datastore using the resource data.
func resourceVSphereFileDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] deleting file: %#v", d)
	f := file{}

	if v, ok := d.GetOk("datacenter"); ok {
		f.datacenter = v.(string)
	}

	if v, ok := d.GetOk("datastore"); ok {
		f.datastore = v.(string)
	} else {
		return fmt.Errorf("datastore argument is required")
	}

	if v, ok := d.GetOk("source_file"); ok {
		f.sourceFile = v.(string)
	} else {
		return fmt.Errorf("source_file argument is required")
	}

	if v, ok := d.GetOk("destination_file"); ok {
		f.destinationFile = v.(string)
	} else {
		return fmt.Errorf("destination_file argument is required")
	}

	client := meta.(*Client).vimClient

	err := deleteFile(client, &f)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// createDirectory ensures the parent directories of the destination file are created in the datastore, if required.
func createDirectory(datastoreFileManager *object.DatastoreFileManager, f *file) error {
	directoryPathIndex := strings.LastIndex(f.destinationFile, "/")
	if directoryPathIndex < 0 {
		// there are no parent directories in the file's name - nothing to create
		return nil
	}

	targetPath := f.destinationFile[0:directoryPathIndex]
	err := datastoreFileManager.FileManager.MakeDirectory(context.TODO(),
		datastoreFileManager.Datastore.Path(targetPath), datastoreFileManager.Datacenter, true)
	if err != nil {
		return err
	}
	return nil
}

// createFile uploads or copies a file to a datastore, with optional directory creation and VMDK handling.
func createFile(client *govmomi.Client, f *file) error {
	finder := find.NewFinder(client.Client, true)

	dstDatacenter, err := finder.Datacenter(context.TODO(), f.datacenter)
	if err != nil {
		return fmt.Errorf("error finding destination datacenter %s", err)
	}
	finder = finder.SetDatacenter(dstDatacenter)

	dstDatastore, err := getDatastore(finder, f.datastore)
	if err != nil {
		return fmt.Errorf("error finding destination datastore %s", err)
	}
	dstDfm := dstDatastore.NewFileManager(dstDatacenter, false)

	if f.createDirectories {
		err = createDirectory(dstDfm, f)
		if err != nil {
			return fmt.Errorf("error creating directory %s", err)
		}
	}

	switch {
	case f.copyFile:
		srcDatacenter, err := finder.Datacenter(context.TODO(), f.sourceDatacenter)
		if err != nil {
			return fmt.Errorf("error finding source datacenter %s", err)
		}
		srcDatastore, err := getDatastore(finder, f.sourceDatastore)
		if err != nil {
			return fmt.Errorf("error finding source datastore %s", err)
		}
		srcDfm := srcDatastore.NewFileManager(srcDatacenter, false)
		srcDfm.DatacenterTarget = dstDatacenter
		dstFilePath := dstDfm.Path(f.destinationFile)
		err = srcDfm.Copy(context.TODO(), f.sourceFile, dstFilePath.String())
		if err != nil {
			return fmt.Errorf("error copying file %s", err)
		}
	case path.Ext(f.sourceFile) == ".vmdk":
		_, fileName := path.Split(f.destinationFile)
		// Temporary directory path to upload VMDK file.
		tempDstFile := fmt.Sprintf("tfm-temp-%d/%s", time.Now().Nanosecond(), fileName)

		err = fileUpload(client, dstDatacenter, dstDatastore, f.sourceFile, tempDstFile)
		if err != nil {
			return fmt.Errorf("error uploading vmdk file %s", err)
		}
		err = dstDfm.Move(context.TODO(), tempDstFile, f.destinationFile)
		if err != nil {
			return fmt.Errorf("error moving vmdk file %s", err)
		}

	default:
		err = fileUpload(client, dstDatacenter, dstDatastore, f.sourceFile, f.destinationFile)
		if err != nil {
			return fmt.Errorf("error uploading file %s", err)
		}
	}

	return nil
}

// deleteFile deletes a file or virtual disk from a datastore.
func deleteFile(client *govmomi.Client, f *file) error {
	dc, err := getDatacenter(client, f.datacenter)
	if err != nil {
		return fmt.Errorf("failed to get datacenter %q: %w", f.datacenter, err)
	}

	finder := find.NewFinder(client.Client, true)
	finder.SetDatacenter(dc)

	ds, err := getDatastore(finder, f.datastore)
	if err != nil {
		return fmt.Errorf("failed to get datastore %q: %w", f.datastore, err)
	}

	var task *object.Task
	var deleteErr error
	datastorePath := ds.Path(f.destinationFile)
	ctx := context.TODO()

	if path.Ext(f.destinationFile) == ".vmdk" {
		vdm := object.NewVirtualDiskManager(client.Client)
		task, deleteErr = vdm.DeleteVirtualDisk(ctx, datastorePath, dc)
	} else {
		fm := object.NewFileManager(client.Client)
		task, deleteErr = fm.DeleteDatastoreFile(ctx, datastorePath, dc)
	}

	if deleteErr != nil {
		return fmt.Errorf("failed to initiate delete for %q: %w", datastorePath, deleteErr)
	}

	_, err = task.WaitForResultEx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error waiting for delete task on %q: %w", datastorePath, err)
	}

	return nil
}

// fileUpload uploads a local file to a datastore.
func fileUpload(client *govmomi.Client, dc *object.Datacenter, ds *object.Datastore, source, destination string) error {
	// Define a slice for the special characters.
	specialChars := []string{"+"}

	// Decode the source path.
	var err error
	source, err = url.PathUnescape(source)
	if err != nil {
		return fmt.Errorf("error unescaping source path %s", err)
	}

	// Clean the source and destination paths.
	source = filepath.Clean(source)
	destination = filepath.Clean(destination)

	// Save the original destination for later use.
	originalDestination := destination

	// Check for special characters in the destination path.
	for _, char := range specialChars {
		if strings.Contains(destination, char) {
			// If it does, replace the special character with its URL-encoded equivalent.
			destination = strings.ReplaceAll(destination, char, url.QueryEscape(char))
		}
	}

	dsurl := ds.NewURL(destination)

	p := soap.DefaultUpload
	err = client.UploadFile(context.TODO(), source, dsurl, &p)
	if err != nil {
		return fmt.Errorf("error uploading file %s", err)
	}

	// Check for special characters in the original destination path.
	for _, char := range specialChars {
		if strings.Contains(originalDestination, char) {
			// If it does, rename the file to the original destination path.
			fm := object.NewFileManager(client.Client)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)

			task, err := fm.MoveDatastoreFile(ctx, ds.Path(destination), dc, ds.Path(originalDestination), dc, false)
			if err != nil {
				cancel()
				return fmt.Errorf("error renaming uploaded file %s", err)
			}

			_, err = task.WaitForResult(ctx, nil)
			cancel()

			if err != nil {
				return fmt.Errorf("error waiting for rename task %s", err)
			}

			break
		}
	}

	return nil
}

// getDatastore returns a reference to a specified datastore or the default datastore if none is provided.
func getDatastore(f *find.Finder, ds string) (*object.Datastore, error) {
	if ds != "" {
		dso, err := f.Datastore(context.TODO(), ds)
		return dso, err
	}
	dso, err := f.DefaultDatastore(context.TODO())
	return dso, err
}
