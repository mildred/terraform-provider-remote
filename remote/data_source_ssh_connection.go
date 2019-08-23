package remote

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mildred/terraform-provider-remote/remote/connection"
)

func dataSourceSSHConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSSHConnectionRead,

		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Description: "Host to connect to",
				Required:    true,
				ForceNew:    true,
			},
			"user": {
				Type:        schema.TypeString,
				Description: "Username to use",
				Default:     "",
				Optional:    true,
				ForceNew:    true,
			},
			"sudo": {
				Type:        schema.TypeBool,
				Description: "If using sudo",
				Default:     false,
				Optional:    true,
				ForceNew:    true,
			},
			"port": {
				Type:        schema.TypeInt,
				Description: "Post number",
				Default:     22,
				Optional:    true,
				ForceNew:    true,
			},
			"conn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSSHConnectionRead(d *schema.ResourceData, _ interface{}) error {
	conn := &connection.Connection{
		SSH: &connection.SSHConnection{
			Host: d.Get("host").(string),
			User: d.Get("user").(string),
			Sudo: d.Get("sudo").(bool),
			Port: d.Get("port").(int),
		},
	}

	encoded, id := conn.Encode()

	d.Set("conn", encoded)
	d.SetId(id)

	return nil
}
