# Example: granting access to a service account.

resource "prefect_service_account" "test" {
  name = "my-service-account"
}

resource "prefect_team" "test" {
  name        = "my-team"
  description = "test-team-description"
}

resource "prefect_team_access" "test" {
  member_type     = "service_account"
  member_id       = prefect_service_account.test.id
  member_actor_id = prefect_service_account.test.actor_id
  team_id         = prefect_team.test.id
}


# Example: granting access to a user.

data "prefect_account_member" "test" {
  email = "marvin@prefect.io"
}

resource "prefect_team" "test" {
  name        = "my-team"
  description = "test-team-description"
}

resource "prefect_team_access" "test" {
  team_id         = prefect_team.test.id
  member_type     = "user"
  member_id       = data.prefect_account_member.test.user_id
  member_actor_id = data.prefect_account_member.test.actor_id
}