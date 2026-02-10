resource "prefect_block_type" "test" {
  name = "my-block-type"
  slug = "my-block-type"

  logo_url          = "https://example.com/logo.png"
  documentation_url = "https://example.com/documentation"
  description       = "My custom block type"
  code_example      = "Some code example"
}