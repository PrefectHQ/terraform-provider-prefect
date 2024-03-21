<p align="center"><img src="https://github.com/PrefectHQ/prefect/assets/3407835/c654cbc6-63e8-4ada-a92a-efd2f8f24b85" width=1000></p>

<p align="center">
    <a href="https://prefect.io/slack" alt="Slack">
        <img src="https://img.shields.io/badge/slack-join_community-red.svg?color=0052FF&labelColor=090422&logo=slack" /></a>
    <a href="https://discourse.prefect.io/" alt="Discourse">
        <img src="https://img.shields.io/badge/discourse-browse_forum-red.svg?color=0052FF&labelColor=090422&logo=discourse" /></a>
</p>

<a href="https://terraform.io">
    <img src=".github/tf.png" alt="Terraform logo" title="Terraform" align="left" height="50" />
</a>

# Terraform Provider for Prefect Cloud
[![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/prefecthq/terraform-provider-prefect?label=release)](https://github.com/prefecthq/terraform-provider-prefect/releases) ![Acceptance tests](https://github.com/PrefectHQ/terraform-provider-prefect/actions/workflows/acceptance-tests.yaml/badge.svg) ![Provider Release](https://github.com/PrefectHQ/terraform-provider-prefect/actions/workflows/provider-release.yaml/badge.svg)

- [Documentation](https://registry.terraform.io/providers/PrefectHQ/prefect/latest/docs)
- [Examples](https://github.com/PrefectHQ/terraform-provider-prefect/tree/main/examples)

[Prefect](https://www.prefect.io/) is a powerful tool for creating workflow applications.  The Terraform Prefect provider is a plugin that allows Terraform to manage resources on [Prefect Cloud](https://app.prefect.cloud). This provider is maintained by the [engineering team at Prefect](https://www.prefect.io/blog#engineering).

## Supported objects

We're continuously adding more Prefect Cloud object support, striving for near-parity by the `v1.x.x` release.

Check back with us to see new additions and improvements - and please don't hesitate to peruse our [Contributing section!](#contributing)

| Prefect Cloud object | Datasource support? | Resource support? | Import support? |
|----------------------:|:---------------------:|:-------------------:|:-----------------:|
| Account Member       |       &check;       |                   |                 |
| Account Role         |       &check;       |                   |                 |
| Account              |       &check;       |      &check;      |     &check;     |
| Service Account      |       &check;       |      &check;      |     &check;     |
| Team                 |       &check;       |                   |                 |
| Variable             |       &check;       |      &check;      |     &check;     |
| Work Pool            |       &check;       |      &check;      |     &check;     |
| Workspace Access     |       &check;       |      &check;      |                 |
| Workspace Role       |       &check;       |      &check;      |     &check;     |
| Workspace            |       &check;       |      &check;      |     &check;     |

## Contributing

We appreciate your interest in the Prefect provider! If you find any issues or have ideas for improvement, you can always:

- [Check out our contributing guide](/_about/CONTRIBUTING.md)
- [File an issue](https://github.com/PrefectHQ/terraform-provider-prefect/issues/new?assignees=&labels=bug&projects=&template=bug.md)
- [Shoot us a note](mailto:security@prefect.io) for any potential security issues
- Drop us a line in the [Prefect Slack community workspace](https://communityinviter.com/apps/prefect-community/prefect-community)
