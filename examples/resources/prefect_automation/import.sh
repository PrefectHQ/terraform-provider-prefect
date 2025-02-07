# prefect_automation resources can be imported by the automation_id
terraform import prefect_automation.my_automation 00000000-0000-0000-0000-000000000000
#
# or from a different workspace via automation_id,workspace_id
terraform import prefect_automation.my_automation 00000000-0000-0000-0000-000000000000,11111111-1111-1111-1111-111111111111

