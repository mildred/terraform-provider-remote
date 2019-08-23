package remote

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mildred/terraform-provider-remote/remote/connection"
)

func dataSourceRemoteFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRemoteFileRead,

		Schema: map[string]*schema.Schema{
			"conn": {
				Type:        schema.TypeString,
				Description: "Connection",
				Required:    true,
				ForceNew:    true,
			},
			"filename": {
				Type:        schema.TypeString,
				Description: "Path to the output file",
				Required:    true,
				ForceNew:    true,
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_base64": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRemoteFileRead(d *schema.ResourceData, _ interface{}) error {
	path := d.Get("filename").(string)
	sess, err := connection.Connect(d.Get("conn").(string))
	if err != nil {
		return err
	}

	content, err := sess.ReadFile(path)
	if err != nil {
		return err
	}

	d.Set("content", string(content))
	d.Set("content_base64", base64.StdEncoding.EncodeToString(content))

	checksum := sha1.Sum([]byte(content))
	d.SetId(hex.EncodeToString(checksum[:]))

	return nil
}
