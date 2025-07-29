package spacelift

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	. "github.com/spacelift-io/terraform-provider-spacelift/spacelift/internal/testhelpers"
)

func TestWorkerPoolResource(t *testing.T) {
	const resourceName = "spacelift_worker_pool.test"

	t.Run("without a CSR", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name        = "My first worker pool %s"
					description = "%s"
				}
			`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", IsNotEmpty()),
					Attribute("description", Equals("old description")),
					Attribute("name", Equals(fmt.Sprintf("My first worker pool %s", randomID))),
					Attribute("private_key", IsNotEmpty()),
					Attribute("config", IsNotEmpty()),
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
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name = "My second worker pool %s"
				csr  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
			}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("My second worker pool %s", randomID))),
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

	t.Run("with labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name = "My third worker pool %s"
				csr  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
				labels = ["label-custom", "label-custom-2"]
			}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("My third worker pool %s", randomID))),
					SetEquals("labels", "label-custom", "label-custom-2"),
					AttributeNotPresent("private_key"),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_worker_pool" "test" {
					name = "Worker pool remove labels test %s"
					csr  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
					labels = ["one", "two"]
			}`, randomID),
				Check: Resource(
					"spacelift_worker_pool.test",
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_worker_pool" "test" {
					name = "Worker pool remove labels test %s"
					csr  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
					labels = []
				}`, randomID),
				Check: Resource(
					"spacelift_worker_pool.test",
					SetEquals("labels"),
				),
			},
		})
	})

	t.Run("with drift detection run limit", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := func(limit int) string {
			return fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name                      = "Worker pool with drift limit %s"
					description               = "Test drift detection run limit"
					drift_detection_run_limit = %d
				}
			`, randomID, limit)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config(5),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("Worker pool with drift limit %s", randomID))),
					Attribute("drift_detection_run_limit", Equals("5")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"csr", "private_key"},
			},
			{
				Config: config(10),
				Check: Resource(
					resourceName,
					Attribute("drift_detection_run_limit", Equals("10")),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "spacelift_worker_pool" "test" {
						name = "Worker pool with drift limit %s"
						description = "Test drift detection run limit"
					}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("drift_detection_run_limit", Equals("-1")),
				),
			},
		})
	})

	t.Run("with drift detection default behavior", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_worker_pool" "test" {
						name        = "Worker pool default drift %s"
						description = "Test default drift detection behavior"
					}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("Worker pool default drift %s", randomID))),
					Attribute("drift_detection_run_limit", Equals("-1")),
				),
			},
		})
	})

	t.Run("with drift detection set to zero", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_worker_pool" "test" {
						name                      = "Worker pool zero drift %s"
						description               = "Test zero drift detection limit"
						drift_detection_run_limit = 0
					}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("Worker pool zero drift %s", randomID))),
					Attribute("drift_detection_run_limit", Equals("0")),
				),
			},
		})
	})

	t.Run("with drift detection negative values", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "spacelift_worker_pool" "test" {
						name                      = "Worker pool negative drift %s"
						drift_detection_run_limit = -5
					}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("Worker pool negative drift %s", randomID))),
					Attribute("drift_detection_run_limit", Equals("-5")),
				),
			},
		})
	})

}

func TestWorkerPoolResourceSpace(t *testing.T) {
	const resourceName = "spacelift_worker_pool.test"

	t.Run("without a CSR", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := func(description string) string {
			return fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name        = "My first worker pool %s"
					description = "%s"
					space_id    = "root"
				}
			`, randomID, description)
		}

		testSteps(t, []resource.TestStep{
			{
				Config: config("old description"),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", IsNotEmpty()),
					Attribute("description", Equals("old description")),
					Attribute("name", Equals(fmt.Sprintf("My first worker pool %s", randomID))),
					Attribute("private_key", IsNotEmpty()),
					Attribute("space_id", Equals("root")),
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
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name = "My second worker pool %s"
				csr  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
			}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("My second worker pool %s", randomID))),
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

	t.Run("with labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
			resource "spacelift_worker_pool" "test" {
				name = "My third worker pool %s"
				csr  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
				labels = ["label-custom", "label-custom-2"]
			}
			`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", IsNotEmpty()),
					Attribute("name", Equals(fmt.Sprintf("My third worker pool %s", randomID))),
					SetEquals("labels", "label-custom", "label-custom-2"),
					AttributeNotPresent("private_key"),
				),
			},
		})
	})

	t.Run("can remove all labels", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "spacelift_worker_pool" "test" {
				name = "Worker pool remove labels test %s"
				csr  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
				labels = ["one", "two"]
			}`, randomID),
				Check: Resource(
					"spacelift_worker_pool.test",
					SetEquals("labels", "one", "two"),
				),
			},
			{
				Config: fmt.Sprintf(`resource "spacelift_worker_pool" "test" {
					name = "Worker pool remove labels test %s"
					csr  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
					labels = []
				}`, randomID),
				Check: Resource(
					"spacelift_worker_pool.test",
					SetEquals("labels"),
				),
			},
		})
	})

	t.Run("CSR changes reset worker pool", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		var originalId string
		var originalConfig string
		csr := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name = "My test workerpool %s"
				}
				`, randomID),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", IsNotEmpty()),
					Attribute("csr", NotEquals(csr)),

					// This function saves the id of the resource to a higher scoped variable
					// so we can use it later
					func(attributes map[string]string) error {
						originalId = attributes["id"]
						originalConfig = attributes["config"]
						return nil
					},
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name = "My test workerpool %s"
					csr  = "%s"
				}
				`, randomID, csr),
				Check: Resource(
					resourceName,
					// We validate the id has not changed, which tells us the resource was not recreated
					Attribute("id", func(v string) error { return Equals(originalId)(v) }),
					// We validate the config has not changed
					Attribute("config", func(v string) error { return NotEquals(originalConfig)(v) }),
					// We also validate the CSR is the new one
					Attribute("csr", Equals(csr)),
				),
			},
		})
	})

	t.Run("Name and description change does not impact csr or config", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		var originalId string
		var originalConfig string
		csr := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURSBSRVFVRVNULS0tLS0KTUlJRWtqQ0NBbm9DQVFBd1RURUxNQWtHQTFVRUJoTUNVRXd4RGpBTUJnTlZCQW9NQldKaFkyOXVNUTR3REFZRApWUVFEREFWaVlXTnZiakVlTUJ3R0NTcUdTSWIzRFFFSkFSWVBZbUZqYjI1QVltRmpiMjR1YjNKbk1JSUNJakFOCkJna3Foa2lHOXcwQkFRRUZBQU9DQWc4QU1JSUNDZ0tDQWdFQXdhand1UmlreTFkODF0TVpJZytJSXBHUHQxclUKV2t4UGhDOENKNzNUWmx3ZTdVcC9McFFiNnpYU0I2eStCWkludnptd1ZBNzNuM0dnVEdFeS9VbDF2VUthaXZmaQpna3lnd05vV0ExYzRTaUNnbjdYTnl1T2c2MktSWGxNb05TeCsrZmVINXZzVGRRVVd2TjZIZkJEQ2dGZ1VQa1JuClp1MDUwOWxBQ2ZrZ00ycnl0b3N3enplbUVUbWRrNlhsYXBnWE9Ebll5bGgvbnRrVFJqZU91VThOUUF1eGRmSUEKY2JFQ0lJZ1Vuak44WWJhWTlGL1RyRjBHUGlQRVZuTEh3Yi95REM2d0NiOXFITUFHRXJhZ0d0cHVzbis0eTBsRwo5S0IvTzZ1R2haRk5HK3FDYUM3MFFKZWI3TzRSdlI3VlA4aWxPOU8rQnE2OU4vY1B4cUFXRTY3WUplUzVxa1hNClFRVVBxVGVXMGs2NC9KZ2c0Nm5ZTmhueGJ5Rkp1MzZ5ME1xbndDN1FYVjZicjFDNldsM3gzTzlNZng4UGVaWGIKdjFqejhod2RWSGFIc0ZLTkgwemdrTk5ISkJ1ZTAyZWwxRkNnbCtMSGNTdWJKdHJnaXpLWkVFSGlFeWhUVURUOQpqeTlSWGpPSUUrNTQ3TkFNMHZvVlY1aTg1eDN0LzdFeFI5R0lraFpwejNQSlV3WUplbHE1M3JPakRvRXZhWTF1CmFUSm9VclYwUUUwK0hTN3ZyaWxXb0VXWlFjOUFiNFFmNnZicmpncCsvVzFEVU5WcGFtVjhQU2dTS3M4RUkwNW4Kek5hc3Q3cnA3b1A2WXBiR2VrbGVQRllWVUVqNTZOKzBxNnh3MFdtS1loNmtYOGRxTTVoNWlkVTFsdUlSU01xMgpkUmJZWStwRFQyeHAwaWtDQXdFQUFhQUFNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUNjZUM2VXRSbnZ3MkRmCkpsQ285NFdIZDRUTTdQYVBtYmdkeVlMSGpacTZKNGdMcGxrcFlnSno2TnA4OThhTExtRDluTlNEV1c0QkpieDkKdXNaaTA3eFl1cjZybjY0cFRUeUhOK1U4WHZsYVdCQjVoMmV3NytZeXVDNWh4RkU1Mjg0OEJ2WG9LNFdmSzRIegpsZ25vWW9qWERNWEpSRTBqR0drVk8rckt4ZW41ak9ZQW4rbkxQT25HNzRSR25kZ2xTYVFhbFFidjFZb095L1dSCll3QzNqM2JodzUrTG9BNVUvaXhZSytma09rZmZOR0VpaU91K0tZV1J6cTVUd2hOKzFHV1l1M3B4WGJ3ZHM3emgKcjlrdVRvdUhpbDg0OTdaZTJGY2t1VTF3OWVaSmY4WlBHWlNLamhmSUhMYWtwNE9UQUlDb1hDS0hhNEhtUVVzRApVdFBjS0E4ZUdkNmh3U0gyS0FndWU2VVdsMDhFZ2xnRlhkOC90Qy9wYzhNR3QxU2RtTzgzUlVEenJLREt3TCszClhNc0xYOWlic1VTZzk3ZzF5R1RxWE1JeUhXK0tiT3lOZS9JYVBYblJJKy9zdkJaTEY0OGQ4UTdKY2xQcHZ6SysKSnlhMXVLWkI4MFRlZnlpaW5oa21GcmcvWmNzdEI2MEI5VFVHaHNib3JmNW5hdnNCcWIxUkN6c2J5VUFvOVphUgpTUXQyNDlMOUc1bmlIcUNTUENxWXVqRktuMWxIVjVicGxwaDFzWHozOVU5RXVTanNxRlNlMlorM0duUVNSSHlNCkx1YTNPT2pmRXh6UUl3Zm5DUy8wMjVIZENjMDZXY3hNK3JUUlA1UW13eGRJNFBtTTNEU2dCRXE0L2RjeEZwTUYKWnp4VkNreU5PWUJPRklTTXRUWDNiQXI3K3JST2VBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tCg=="
		testSteps(t, []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name = "My workerpool to update %s"
					csr  = "%s"
				}
				`, randomID, csr),
				Check: Resource(
					resourceName,
					Attribute("id", IsNotEmpty()),
					Attribute("csr", Equals(csr)),
					Attribute("name", Equals(fmt.Sprintf("My workerpool to update %s", randomID))),
					func(attributes map[string]string) error {
						originalId = attributes["id"]
						originalConfig = attributes["config"]
						return nil
					},
				),
			},
			{
				Config: fmt.Sprintf(`
				resource "spacelift_worker_pool" "test" {
					name        = "My workerpool to update %s - updated"
					description = "My workerpool to update %s - updated"
					csr  = "%s"
				}
				`, randomID, randomID, csr),
				Check: Resource(
					resourceName,
					Attribute("id", func(v string) error { return Equals(originalId)(v) }),
					Attribute("config", func(v string) error { return Equals(originalConfig)(v) }),
					Attribute("csr", Equals(csr)),
					Attribute("name", Equals(fmt.Sprintf("My workerpool to update %s - updated", randomID))),
					Attribute("description", Equals(fmt.Sprintf("My workerpool to update %s - updated", randomID))),
				),
			},
		})
	})
}
