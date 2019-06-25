package azurerm

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceArmManagedDisk() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmManagedDiskRead,
		Schema: map[string]*schema.Schema{

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"zones": azure.SchemaZonesComputed(),

			"storage_account_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"os_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"disk_size_gb": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"create_option": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func dataSourceArmManagedDiskRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).diskClient
	ctx := meta.(*ArmClient).StopContext

	resGroup := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)

	resp, err := client.Get(ctx, resGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Error: Managed Disk %q (Resource Group %q) was not found", name, resGroup)
		}
		return fmt.Errorf("[ERROR] Error making Read request on Azure Managed Disk %q (Resource Group %q): %s", name, resGroup, err)
	}

	d.SetId(*resp.ID)

	if sku := resp.Sku; sku != nil {
		d.Set("storage_account_type", string(sku.Name))
	}

	if props := resp.DiskProperties; props != nil {
		if diskSize := props.DiskSizeGB; diskSize != nil {
			d.Set("disk_size_gb", *diskSize)
		}
		if osType := props.OsType; osType != "" {
			d.Set("os_type", string(osType))
		}
	}

	if resp.CreationData != nil {
		flattenAzureRmManagedDiskCreationData(d, resp.CreationData)
	}

	d.Set("zones", resp.Zones)

	flattenAndSetTags(d, resp.Tags)

	return nil
}
