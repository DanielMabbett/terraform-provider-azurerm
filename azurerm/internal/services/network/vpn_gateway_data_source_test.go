package network_test

import (
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
)

type VPNGatewayDataSource struct {
}

func TestAccVPNGatewayDataSource_bgpSettings(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_vpn_gateway", "test")
	r := VPNGatewayDataSource{}

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: r.bgpSettings(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("location").Exists(),
				check.That(data.ResourceName).Key("bgp_settings.asn").HasValue("65515"),
				check.That(data.ResourceName).Key("bgp_settings.peer_weight").HasValue("0"),
				check.That(data.ResourceName).Key("bgp_settings.instance_0_bgp_peering_address.custom_ips.0").HasValue("169.254.21.10"),
				check.That(data.ResourceName).Key("bgp_settings.instance_1_bgp_peering_address.custom_ips.0").HasValue("169.254.21.10"),
			),
		},
		data.ImportStep(),
	})
}

func (r VPNGatewayDataSource) bgpSettings(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestvn-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurerm_virtual_wan" "test" {
  name                = "acctestvwan-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_virtual_hub" "test" {
  name                = "acctestvh-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  address_prefix      = "10.0.1.0/24"
  virtual_wan_id      = azurerm_virtual_wan.test.id
}

resource "azurerm_vpn_gateway" "test" {
  name                = "acctestVPNG-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  virtual_hub_id      = azurerm_virtual_hub.test.id

  bgp_settings {
    asn         = 65515
    peer_weight = 0

    instance_0_bgp_peering_address {
      custom_ips = ["169.254.21.5"]
    }

    instance_1_bgp_peering_address {
      custom_ips = ["169.254.21.10"]
    }
  }
}

data "azurerm_vpn_gateway" "test" {
  name = azurerm_vpn_gateway.test.name
  resource_group_name = azurerm_vpn_gateway.test.resource_group_name
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
