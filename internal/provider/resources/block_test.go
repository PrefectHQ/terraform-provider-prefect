package resources_test

// func fixtureAccBlockCreate(name string) string {
// 	return fmt.Sprintf(`
// 		data "prefect_workspace" "evergreen" {
// 			handle = "github-ci-tests"
// 		}
// 		resource "prefect_block" "block" {
// 			name = "%s"
// 			type_slug = "secret"
// 			workspace_id = data.prefect_workspace.evergreen.id
// 			data = jsonencode({
// 				"value": "hello, world"
// 			})
// 		}
// 	`, name)
// }

// //nolint:paralleltest // we use the resource.ParallelTest helper instead
// func TestAccResource_block(t *testing.T) {
// 	resourceName := "prefect_block.block"
// 	randomName := testutils.TestAccPrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

// 	resource.ParallelTest(t, resource.TestCase{
// 		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
// 		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
// 		Steps: []resource.TestStep{
// 			{
// 				// Check creation + existence of the block resource
// 				Config: fixtureAccBlockCreate(randomName),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceName, "name", randomName),
// 					resource.TestCheckResourceAttr(resourceName, "type_slug", "secret"),
// 					resource.TestCheckResourceAttr(resourceName, "data", `{"value":"hello, world"}`),
// 				),
// 			},
// 		},
// 	})
// }
