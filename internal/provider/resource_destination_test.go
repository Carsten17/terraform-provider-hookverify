package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"hookverify": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("HOOKVERIFY_API_KEY") == "" {
		t.Skip("HOOKVERIFY_API_KEY not set, skipping acceptance test")
	}
}

func TestAccDestination_basic(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hookverify_destination.test", "name", "Test Destination"),
					resource.TestCheckResourceAttr("hookverify_destination.test", "url", "https://httpbin.org/post"),
					resource.TestCheckResourceAttr("hookverify_destination.test", "active", "true"),
					resource.TestCheckResourceAttr("hookverify_destination.test", "max_retries", "3"),
					resource.TestCheckResourceAttrSet("hookverify_destination.test", "id"),
				),
			},
		},
	})
}

func TestAccDestination_update(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hookverify_destination.test", "name", "Test Destination"),
				),
			},
			{
				Config: testAccDestinationConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hookverify_destination.test", "name", "Updated Destination"),
					resource.TestCheckResourceAttr("hookverify_destination.test", "url", "https://httpbin.org/anything"),
					resource.TestCheckResourceAttr("hookverify_destination.test", "max_retries", "5"),
				),
			},
		},
	})
}

func TestAccDestination_import(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDestinationConfig_basic(),
			},
			{
				ResourceName:      "hookverify_destination.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDestinationConfig_basic() string {
	return `
resource "hookverify_destination" "test" {
  name = "Test Destination"
  url  = "https://httpbin.org/post"
}
`
}

func testAccDestinationConfig_updated() string {
	return `
resource "hookverify_destination" "test" {
  name        = "Updated Destination"
  url         = "https://httpbin.org/anything"
  max_retries = 5
}
`
}
