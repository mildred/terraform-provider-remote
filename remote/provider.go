package remote

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"remote_file": resourceRemoteFile(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"remote_file":           dataSourceRemoteFile(),
			"remote_ssh_connection": dataSourceSSHConnection(),
		},
	}
}
