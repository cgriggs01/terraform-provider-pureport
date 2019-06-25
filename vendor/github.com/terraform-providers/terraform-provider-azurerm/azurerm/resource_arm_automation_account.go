package azurerm

import (
	"fmt"
	"log"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/services/automation/mgmt/2015-10-31/automation"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmAutomationAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmAutomationAccountCreateUpdate,
		Read:   resourceArmAutomationAccountRead,
		Update: resourceArmAutomationAccountCreateUpdate,
		Delete: resourceArmAutomationAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					//todo this will not allow single character names, even thou they are valid
					regexp.MustCompile(`^[0-9a-zA-Z]([-0-9a-zA-Z]{0,48}[0-9a-zA-Z])?$`),
					`The account name must not be empty, and must not exceed 50 characters in length.  The account name must start with a letter or number.  The account name can contain letters, numbers, and dashes. The final character must be a letter or a number.`,
				),
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"sku": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          string(automation.Basic),
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(automation.Basic),
								string(automation.Free),
							}, true),
						},
					},
				},
			},

			"tags": tagsSchema(),

			"dsc_server_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dsc_primary_access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dsc_secondary_access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceArmAutomationAccountCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automation.AccountClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for Automation Account create/update.")

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)

	if requireResourcesToBeImported && d.IsNewResource() {
		existing, err := client.Get(ctx, resGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Automation Account %q (Resource Group %q): %s", name, resGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_automation_account", *existing.ID)
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	tags := d.Get("tags").(map[string]interface{})
	sku := expandAutomationAccountSku(d)

	parameters := automation.AccountCreateOrUpdateParameters{
		AccountCreateOrUpdateProperties: &automation.AccountCreateOrUpdateProperties{
			Sku: sku,
		},
		Location: utils.String(location),
		Tags:     expandTags(tags),
	}

	if _, err := client.CreateOrUpdate(ctx, resGroup, name, parameters); err != nil {
		return fmt.Errorf("Error creating/updating Automation Account %q (Resource Group %q) %+v", name, resGroup, err)
	}

	read, err := client.Get(ctx, resGroup, name)
	if err != nil {
		return fmt.Errorf("Error retrieving Automation Account %q (Resource Group %q) %+v", name, resGroup, err)
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Automation Account %q (Resource Group %q) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmAutomationAccountRead(d, meta)
}

func resourceArmAutomationAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automation.AccountClient
	registrationClient := meta.(*ArmClient).automation.AgentRegistrationInfoClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["automationAccounts"]

	resp, err := client.Get(ctx, resGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Automation Account %q was not found in Resource Group %q - removing from state!", name, resGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Automation Account %q (Resource Group %q): %+v", name, resGroup, err)
	}

	keysResp, err := registrationClient.Get(ctx, resGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Agent Registration Info for Automation Account %q was not found in Resource Group %q - removing from state!", name, resGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request for Agent Registration Info for Automation Account %q (Resource Group %q): %+v", name, resGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if err := d.Set("sku", flattenAutomationAccountSku(resp.Sku)); err != nil {
		return fmt.Errorf("Error setting `sku`: %+v", err)
	}

	d.Set("dsc_server_endpoint", keysResp.Endpoint)
	if keys := keysResp.Keys; keys != nil {
		d.Set("dsc_primary_access_key", keys.Primary)
		d.Set("dsc_secondary_access_key", keys.Secondary)
	}

	if tags := resp.Tags; tags != nil {
		flattenAndSetTags(d, tags)
	}

	return nil
}

func resourceArmAutomationAccountDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automation.AccountClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["automationAccounts"]

	resp, err := client.Delete(ctx, resGroup, name)

	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return fmt.Errorf("Error issuing AzureRM delete request for Automation Account '%s': %+v", name, err)
	}

	return nil
}

func flattenAutomationAccountSku(sku *automation.Sku) []interface{} {
	if sku == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{}
	result["name"] = string(sku.Name)
	return []interface{}{result}
}

func expandAutomationAccountSku(d *schema.ResourceData) *automation.Sku {
	inputs := d.Get("sku").([]interface{})
	input := inputs[0].(map[string]interface{})
	name := automation.SkuNameEnum(input["name"].(string))

	sku := automation.Sku{
		Name: name,
	}

	return &sku
}
