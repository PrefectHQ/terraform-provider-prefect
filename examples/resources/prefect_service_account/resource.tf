resource "prefect_service_account" "test" {
  name = "test"
  role = "USER"
  api_keys = [
    {
      name       = "key1"
      expiration = "2015-10-20T13:00:00+00:00"
    },
    {
      name = "key2"
    },
  ]
}
