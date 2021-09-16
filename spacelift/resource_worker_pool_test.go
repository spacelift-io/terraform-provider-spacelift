package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWorkerPoolResource(t *testing.T) {
	const resourceName = "spacelift_worker_pool.test"

	t.Run("without a CSR", func(t *testing.T) {
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name        = "My first worker pool"
					description = "%s"
				}
			`, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", IsNotEmpty()),
					Attribute("description", Equals("old description")),
					Attribute("name", Equals("My first worker pool")),
					Attribute("private_key", IsNotEmpty()),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"csr", "private_key"},
			},
			{
				Config: config("new description"),
				Check: Resource(
					resourceName,
					Attribute("description", Equals("new description")),
				),
			},
		})
	})

	t.Run("with a CSR", func(t *testing.T) {
		testSteps(t, []resource.TestStep{
			{
				Config: `
			resource "spacelift_worker_pool" "test" {
				name = "My second worker pool"
				csr  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
			}
			`,
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", IsNotEmpty()),
					Attribute("name", Equals("My second worker pool")),
					AttributeNotPresent("private_key"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"csr", "private_key"},
			},
		})
	})
}
