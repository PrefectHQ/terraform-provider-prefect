# prefect_automation resources can be imported by the Automation ID
terraform import prefect_automation.my_automation 00000000-0000-0000-0000-000000000000

# Pass an optional, comma-separated value following the identifier
# if you need to import a resource in a different workspace
# from the one that your provider is configured with
# NOTE: you must specify the workspace_id attribute in the addressed resource
#
# <block_id>,<workspace_id>
terraform import prefect_automation.my_automation 00000000-0000-0000-0000-000000000000,11111111-1111-1111-1111-111111111111

