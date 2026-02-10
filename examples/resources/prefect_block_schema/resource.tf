resource "prefect_block_type" "test" {
  name = "my-block-type"
  slug = "my-block-type"

  logo_url          = "https://example.com/logo.png"
  documentation_url = "https://example.com/documentation"
  description       = "My custom block type"
  code_example      = "Some code example"
}

resource "prefect_block_schema" "test" {
  block_type_id = prefect_block_type.test.id

  capabilities = ["read", "write"]
  version      = "1.0.0"
  fields = jsonencode({
    "title" : "x",
    "type" : "object",
    "properties" : {
      "foo" : {
        "title" : "Foo",
        "type" : "string"
      }
    }
  })
}