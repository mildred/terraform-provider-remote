package remote

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mildred/terraform-provider-remote/remote/connection"
)

func resourceRemoteFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceRemoteFileCreate,
		Read:   resourceRemoteFileRead,
		Delete: resourceRemoteFileDelete,

		Schema: map[string]*schema.Schema{
			"conn": {
				Type:        schema.TypeString,
				Description: "Connection",
				Required:    true,
				ForceNew:    true,
			},
			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"sensitive_content", "content_base64"},
			},
			"sensitive_content": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"content", "content_base64"},
			},
			"content_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"sensitive_content", "content"},
			},
			"filename": {
				Type:        schema.TypeString,
				Description: "Path to the output file",
				Required:    true,
				ForceNew:    true,
			},
			"file_permission": {
				Type:         schema.TypeString,
				Description:  "Permissions to set for the output file",
				Optional:     true,
				ForceNew:     true,
				Default:      "0777",
				ValidateFunc: validateMode,
			},
			"directory_permission": {
				Type:         schema.TypeString,
				Description:  "Permissions to set for directories created",
				Optional:     true,
				ForceNew:     true,
				Default:      "0777",
				ValidateFunc: validateMode,
			},
		},
	}
}

func resourceRemoteFileRead(d *schema.ResourceData, _ interface{}) error {
	sess, err := connection.Connect(d.Get("conn").(string))
	if err != nil {
		return err
	}

	// If the output file doesn't exist, mark the resource for creation.
	outputPath := d.Get("filename").(string)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		d.SetId("")
		return nil
	}

	// Verify that the content of the destination file matches the content we
	// expect. Otherwise, the file might have been modified externally and we
	// must reconcile.
	outputContent, err := sess.ReadFile(outputPath)
	if err != nil {
		return err
	}

	outputChecksum := sha1.Sum([]byte(outputContent))
	if hex.EncodeToString(outputChecksum[:]) != d.Id() {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceRemoteFileContent(d *schema.ResourceData, sess *connection.Session) ([]byte, error) {
	if content, sensitiveSpecified := d.GetOk("sensitive_content"); sensitiveSpecified {
		return []byte(content.(string)), nil
	}
	if b64Content, b64Specified := d.GetOk("content_base64"); b64Specified {
		return base64.StdEncoding.DecodeString(b64Content.(string))
	}

	content := d.Get("content")
	return []byte(content.(string)), nil
}

func resourceRemoteFileCreate(d *schema.ResourceData, _ interface{}) error {
	sess, err := connection.Connect(d.Get("conn").(string))
	if err != nil {
		return err
	}

	content, err := resourceRemoteFileContent(d, sess)
	if err != nil {
		return err
	}

	destination := d.Get("filename").(string)

	destinationDir := path.Dir(destination)
	if _, err := os.Stat(destinationDir); err != nil {
		dirPerm := d.Get("directory_permission").(string)
		dirMode, _ := strconv.ParseInt(dirPerm, 8, 64)
		if err := os.MkdirAll(destinationDir, os.FileMode(dirMode)); err != nil {
			return err
		}
	}

	filePerm := d.Get("file_permission").(string)

	fileMode, _ := strconv.ParseInt(filePerm, 8, 64)

	if err := sess.WriteFile(destination, []byte(content), os.FileMode(fileMode)); err != nil {
		return err
	}

	checksum := sha1.Sum([]byte(content))
	d.SetId(hex.EncodeToString(checksum[:]))

	return nil
}

func resourceRemoteFileDelete(d *schema.ResourceData, _ interface{}) error {
	sess, err := connection.Connect(d.Get("conn").(string))
	if err != nil {
		return err
	}

	return sess.RemoveFile(d.Get("filename").(string), true, false)
}
