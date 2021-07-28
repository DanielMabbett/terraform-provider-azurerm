package web

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2021-02-01/web"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/internal/location"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	networkParse "github.com/hashicorp/terraform-provider-azurerm/internal/services/network/parse"
	networkValidate "github.com/hashicorp/terraform-provider-azurerm/internal/services/network/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/web/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/web/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tags"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

const KindASEV3 = "ASEV3"

type ClusterSettingModel struct {
	Name  string `tfschema:"name"`
	Value string `tfschema:"value"`
}

type AppServiceEnvironmentV3Model struct {
	Name                               string                            `tfschema:"name"`
	ResourceGroup                      string                            `tfschema:"resource_group_name"`
	SubnetId                           string                            `tfschema:"subnet_id"`
	AllowNewPrivateEndpointConnections bool                              `tfschema:"allow_new_private_endpoint_connections"`
	ClusterSetting                     []ClusterSettingModel             `tfschema:"cluster_setting"`
	DedicatedHostCount                 int                               `tfschema:"dedicated_host_count"`
	InternalLoadBalancingMode          string                            `tfschema:"internal_load_balancing_mode"`
	Location                           string                            `tfschema:"location"`
	ZoneRedundant                      bool                              `tfschema:"zone_redundant"`
	DnsSuffix                          string                            `tfschema:"dns_suffix"`
	IpSSLAddressCount                  int                               `tfschema:"ip_ssl_address_count"`
	PricingTier                        string                            `tfschema:"pricing_tier"`
	WindowsOutboundIPAddresses         []string                          `tfschema:"windows_outbound_ip_addresses"`
	LinuxOutboundIPAddresses           []string                          `tfschema:"linux_outbound_ip_addresses"`
	InboundNetworkDependencies         []AppServiceV3InboundDependencies `tfschema:"inbound_network_dependencies"`
	Tags                               map[string]interface{}            `tfschema:"tags"`
}

type AppServiceV3InboundDependencies struct {
	Description string   `tfschema:"description"`
	IPAddresses []string `tfschema:"ip_addresses"`
	Ports       []string `tfschema:"ports"`
}

// (@jackofallops) - Important property missing from the SDK / Swagger that will need to be added later: `upgrade_preference` https://docs.microsoft.com/en-us/azure/app-service/environment/using#upgrade-preference

type AppServiceEnvironmentV3Resource struct{}

var _ sdk.Resource = AppServiceEnvironmentV3Resource{}
var _ sdk.ResourceWithUpdate = AppServiceEnvironmentV3Resource{}

func (r AppServiceEnvironmentV3Resource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.AppServiceEnvironmentName,
		},

		"resource_group_name": azure.SchemaResourceGroupName(),

		"subnet_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ValidateFunc: networkValidate.SubnetID,
		},

		"allow_new_private_endpoint_connections": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			Default:  true,
		},

		"cluster_setting": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"value": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
				},
			},
		},

		"dedicated_host_count": {
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.IntBetween(2, 2), // Docs suggest is limited to 2 physical hosts at this time
			ConflictsWith: []string{
				"zone_redundant",
			},
		},

		"internal_load_balancing_mode": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			ForceNew: true,
			Default:  string(web.LoadBalancingModeNone),
			ValidateFunc: validation.StringInSlice([]string{
				string(web.LoadBalancingModeNone),
				string(web.LoadBalancingModeWebPublishing),
			}, false),
		},

		"zone_redundant": {
			Type:     pluginsdk.TypeBool,
			ForceNew: true,
			Optional: true,
			Default:  false,
			ConflictsWith: []string{
				"dedicated_host_count",
			},
		},

		"tags": tags.ForceNewSchema(),
	}
}

func (r AppServiceEnvironmentV3Resource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"dns_suffix": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"pricing_tier": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"ip_ssl_address_count": {
			Type:     pluginsdk.TypeInt,
			Computed: true,
		},

		"location": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"windows_outbound_ip_addresses": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"linux_outbound_ip_addresses": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"inbound_network_dependencies": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"description": {
						Type:     pluginsdk.TypeString,
						Computed: true,
					},

					"ip_addresses": {
						Type:     pluginsdk.TypeList,
						Computed: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},

					"ports": {
						Type:     pluginsdk.TypeList,
						Computed: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},
				},
			},
		},
	}
}

func (r AppServiceEnvironmentV3Resource) ModelObject() interface{} {
	return AppServiceEnvironmentV3Model{}
}

func (r AppServiceEnvironmentV3Resource) ResourceType() string {
	return "azurerm_app_service_environment_v3"
}

func (r AppServiceEnvironmentV3Resource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 6 * time.Hour,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Web.AppServiceEnvironmentsClient
			networksClient := metadata.Client.Network.VnetClient
			subscriptionId := metadata.Client.Account.SubscriptionId

			var model AppServiceEnvironmentV3Model
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding %+v", err)
			}

			subnet, err := networkParse.SubnetID(model.SubnetId)
			if err != nil {
				return err
			}

			vnet, err := networksClient.Get(ctx, subnet.ResourceGroup, subnet.VirtualNetworkName, "")
			if err != nil {
				return fmt.Errorf("retrieving Virtual Network %q (Resource Group %q): %+v", subnet.VirtualNetworkName, subnet.ResourceGroup, err)
			}

			vnetLoc := location.NormalizeNilable(vnet.Location)
			if vnetLoc == "" {
				return fmt.Errorf("determining Location from Virtual Network %q (Resource Group %q): `location` was missing", subnet.VirtualNetworkName, subnet.ResourceGroup)
			}

			id := parse.NewAppServiceEnvironmentID(subscriptionId, model.ResourceGroup, model.Name)
			existing, err := client.Get(ctx, id.ResourceGroup, id.HostingEnvironmentName)
			if err != nil {
				if !utils.ResponseWasNotFound(existing.Response) {
					return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
				}
			}
			if !utils.ResponseWasNotFound(existing.Response) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			envelope := web.AppServiceEnvironmentResource{
				Kind:     utils.String(KindASEV3),
				Location: utils.String(vnetLoc),
				AppServiceEnvironment: &web.AppServiceEnvironment{
					DedicatedHostCount:        utils.Int32(int32(model.DedicatedHostCount)),
					ClusterSettings:           expandClusterSettingsModel(model.ClusterSetting),
					InternalLoadBalancingMode: web.LoadBalancingMode(model.InternalLoadBalancingMode),
					VirtualNetwork: &web.VirtualNetworkProfile{
						ID: utils.String(model.SubnetId),
					},
					// ZoneRedundant: utils.bool(model.ZoneRedundant) // TODO - SDK missing support for this property at time of writing, tracked at https://github.com/Azure/azure-rest-api-specs/issues/15385
				},
				Tags: tags.Expand(model.Tags),
			}

			if _, err = client.CreateOrUpdate(ctx, id.ResourceGroup, id.HostingEnvironmentName, envelope); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			createWait := pluginsdk.StateChangeConf{
				Pending: []string{
					string(web.ProvisioningStateInProgress),
				},
				Target: []string{
					string(web.ProvisioningStateSucceeded),
				},
				MinTimeout:     1 * time.Minute,
				NotFoundChecks: 20,
				Refresh:        appServiceEnvironmentRefresh(ctx, client, id.ResourceGroup, id.HostingEnvironmentName),
			}

			timeout, _ := ctx.Deadline()
			createWait.Timeout = time.Until(timeout)

			if _, err := createWait.WaitForStateContext(ctx); err != nil {
				return fmt.Errorf("waiting for the creation of %s: %+v", id, err)
			}

			if !model.AllowNewPrivateEndpointConnections {
				aseNetworkConfig := web.AseV3NetworkingConfiguration{
					AseV3NetworkingConfigurationProperties: &web.AseV3NetworkingConfigurationProperties{
						AllowNewPrivateEndpointConnections: utils.Bool(model.AllowNewPrivateEndpointConnections),
					},
				}
				if _, err := client.UpdateAseNetworkingConfiguration(ctx, id.ResourceGroup, id.HostingEnvironmentName, aseNetworkConfig); err != nil {
					return fmt.Errorf("setting Allow New Private Endpoint Connections on %s: %+v", id, err)
				}
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (r AppServiceEnvironmentV3Resource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,

		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Web.AppServiceEnvironmentsClient
			id, err := parse.AppServiceEnvironmentID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			existing, err := client.Get(ctx, id.ResourceGroup, id.HostingEnvironmentName)
			if err != nil {
				if utils.ResponseWasNotFound(existing.Response) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			model := AppServiceEnvironmentV3Model{
				Name:          id.HostingEnvironmentName,
				ResourceGroup: id.ResourceGroup,
				Location:      location.NormalizeNilable(existing.Location),
			}

			if props := existing.AppServiceEnvironment; props != nil {
				if props.VirtualNetwork != nil {
					model.SubnetId = utils.NormalizeNilableString(props.VirtualNetwork.ID)
				}
				model.InternalLoadBalancingMode = string(props.InternalLoadBalancingMode)
				model.DedicatedHostCount = int(utils.NormaliseNilableInt32(props.DedicatedHostCount))
				model.PricingTier = utils.NormalizeNilableString(props.MultiSize)
				model.ClusterSetting = flattenClusterSettingsModel(props.ClusterSettings)
				model.DnsSuffix = utils.NormalizeNilableString(props.DNSSuffix)
				model.IpSSLAddressCount = int(utils.NormaliseNilableInt32(existing.IpsslAddressCount))
				// model.ZoneRedundant = *props.ZoneRedundant
			}

			existingNetwork, err := client.GetAseV3NetworkingConfiguration(ctx, id.ResourceGroup, id.HostingEnvironmentName)
			if err != nil {
				return fmt.Errorf("reading network configuration for %s: %+v", id, err)
			}

			if props := existingNetwork.AseV3NetworkingConfigurationProperties; props != nil {
				model.WindowsOutboundIPAddresses = *props.WindowsOutboundIPAddresses
				model.LinuxOutboundIPAddresses = *props.LinuxOutboundIPAddresses
				model.AllowNewPrivateEndpointConnections = *props.AllowNewPrivateEndpointConnections
			}

			// Inbound Details
			inboundNetworkDependencies := &[]AppServiceV3InboundDependencies{}
			inboundNetworkDependencies, err = flattenInboundNetworkDependencies(ctx, client, id)
			if err != nil {
				return err
			}
			model.InboundNetworkDependencies = *inboundNetworkDependencies
			model.Tags = tags.Flatten(existing.Tags)

			return metadata.Encode(&model)
		},
	}
}

func (r AppServiceEnvironmentV3Resource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 6 * time.Hour,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Web.AppServiceEnvironmentsClient

			id, err := parse.AppServiceEnvironmentID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			future, err := client.Delete(ctx, id.ResourceGroup, id.HostingEnvironmentName, utils.Bool(false))
			if err != nil {
				return fmt.Errorf("deleting %s: %+v", id, err)
			}

			if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
				// This future can return a 404 for the polling check if the ASE is successfully deleted but this raises an error in the SDK
				if !response.WasNotFound(future.Response()) {
					return fmt.Errorf("waiting for removal of %s: %+v", id, err)
				}
			}

			return nil
		},
	}
}

func (r AppServiceEnvironmentV3Resource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return validate.AppServiceEnvironmentID
}

func (r AppServiceEnvironmentV3Resource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 6 * time.Hour,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := parse.AppServiceEnvironmentID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			metadata.Logger.Info("Decoding state...")
			var state AppServiceEnvironmentV3Model
			if err := metadata.Decode(&state); err != nil {
				return err
			}

			metadata.Logger.Infof("updating %s", id)
			client := metadata.Client.Web.AppServiceEnvironmentsClient

			patch := web.AppServiceEnvironmentPatchResource{
				AppServiceEnvironment: &web.AppServiceEnvironment{},
			}

			if metadata.ResourceData.HasChange("cluster_setting") {
				patch.AppServiceEnvironment.ClusterSettings = expandClusterSettingsModel(state.ClusterSetting)
			}

			if _, err = client.Update(ctx, id.ResourceGroup, id.HostingEnvironmentName, patch); err != nil {
				return fmt.Errorf("updating %s: %+v", id, err)
			}

			if metadata.ResourceData.HasChange("subnet_id") {
				vnetInfo := web.VirtualNetworkProfile{
					ID: utils.String(state.SubnetId),
				}

				updateFuture, err := client.ChangeVnet(ctx, id.ResourceGroup, id.HostingEnvironmentName, vnetInfo)
				if err != nil {
					return fmt.Errorf("updating Virtual Network location for %s: %+v", id, err)
				}

				if err := updateFuture.WaitForCompletionRef(ctx, client.Client); err != nil {
					return fmt.Errorf("waiting for update of Virtual Network location for %s: %+v", id, err)
				}
			}

			return nil
		},
	}
}

func flattenClusterSettingsModel(input *[]web.NameValuePair) []ClusterSettingModel {
	var output []ClusterSettingModel
	if input == nil || len(*input) == 0 {
		return output
	}

	for _, v := range *input {
		if v.Name == nil {
			continue
		}

		output = append(output, ClusterSettingModel{
			Name:  *v.Name,
			Value: utils.NormalizeNilableString(v.Value),
		})
	}
	return output
}

func expandClusterSettingsModel(input []ClusterSettingModel) *[]web.NameValuePair {
	var clusterSettings []web.NameValuePair
	if input == nil {
		return &clusterSettings
	}

	for _, v := range input {
		clusterSettings = append(clusterSettings, web.NameValuePair{
			Name:  utils.String(v.Name),
			Value: utils.String(v.Value),
		})
	}

	return &clusterSettings
}

func flattenInboundNetworkDependencies(ctx context.Context, client *web.AppServiceEnvironmentsClient, id *parse.AppServiceEnvironmentId) (*[]AppServiceV3InboundDependencies, error) {
	var results []AppServiceV3InboundDependencies
	for inboundNetworking, err := client.GetInboundNetworkDependenciesEndpointsComplete(ctx, id.ResourceGroup, id.HostingEnvironmentName); inboundNetworking.NotDone(); inboundNetworking.NextWithContext(ctx) {
		if err != nil {
			return nil, fmt.Errorf("reading Inbound Network dependencies for %s: %+v", id, err)
		}
		value := inboundNetworking.Value()
		result := AppServiceV3InboundDependencies{
			Description: utils.NormalizeNilableString(value.Description),
		}

		if value.Endpoints != nil {
			result.IPAddresses = *value.Endpoints
		}

		if value.Ports != nil {
			result.Ports = *value.Ports
		}

		results = append(results, result)
	}

	return &results, nil
}
