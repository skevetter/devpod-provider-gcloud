# GCLOUD Provider for DevPod

[![Join us on Slack!](docs/static/media/slack.svg)](https://slack.loft.sh/) [![Open in DevPod!](https://devpod.sh/assets/open-in-devpod.svg)](https://devpod.sh/open#https://github.com/skevetter/devpod-provider-gcloud)

## Google Cloud Project
This provider runs workspaces on [Google Cloud](https://cloud.google.com/)
virtual machines.
To use it, you need to obtain [Google Cloud](https://cloud.google.com/)
access for your personal Google account,
start the free trial, and set up billing.

You can manage Google Cloud via:
- [Google Cloud Console](https://console.cloud.google.com/)
- [Google Cloud Shell](https://docs.cloud.google.com/shell/docs/how-cloud-shell-works)
- [Google Cloud CLI](https://docs.cloud.google.com/sdk/gcloud) installed on your machine

If you want to use DevPod UI, you need to install
[Google Cloud CLI](https://docs.cloud.google.com/sdk/gcloud)
on your machine:
it is used for authenticating the provider when it is invoked
by DevPod UI - see [Application Default Credentials](#application-default-credentials).

Before you can use this provider, you need to
create a GCP `project` for DevPod virtual machines:
```shell
$ gcloud projects create <your-devpod-project-id>
```

Then, enable `Compute Engine API` for the project:
```shell
$ gcloud services enable compute.googleapis.com \
  --project <your-devpod-project-id>
```

## Authentication

There are two approaches you can use to authenticate this provider to Google Cloud.

### Application Default Credentials

With this approach,
credentials file is deposited into a specific file
where Google Cloud libraries used by DevPod can find it:
- on Linux and macOS: `$HOME/.config/gcloud/application_default_credentials.json`
- on Windows: `%APPDATA%\gcloud\application_default_credentials.json`

For details, see:
- [Application Default Credentials](https://developers.google.com/accounts/docs/application-default-credentials)
- [Authentication: ADC](https://docs.cloud.google.com/docs/authentication/application-default-credentials#personal)

To configure ADC, run:
```shell
$ gcloud auth application-default login <your-personal-Google-account>
```

### Environment Variable

With this approach, environment variable `GOOGLE_APPLICATION_CREDENTIALS`
is set to the path of the JSON `key` file;
Google Cloud libraries used by DevPod use this environment variable
to locate the `key`.

Use one of the following provider options make the `key` available to the provider:
- `KEY_FILE`: _path_ to the JSON `key` file
- `KEY`: _content_ of the JSON `key` file

Since this approach uses JSON `key`s for authentication,
and only Service Accounts have such keys,
you need to use a `service account`.
See [Service Account](#service-account) for details.

With this approach, you lose the ability to use DevPod UI;
you'll need to work with DevPod CLI.

With this approach, you gain the ability to
use more than one account at a time:
if you want to work on projects from multiple
organizations and run workspaces on their respective
Google Cloud infrastructures,
you can configure CLI differently for each of them using `context`s.
See [CLI Workflow](#cli-workflow) for details.

## Service Account

Besides the inability to use more than one account at a time,
there are _security-related_ arguments for using a `service account`
to run things in Google Cloud; for details, see:
- [Principle of Least Privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege)
- [Using IAM Securely](https://cloud.google.com/iam/docs/using-iam-securely)
- [Service Accounts](https://docs.cloud.google.com/docs/authentication#service-accounts)
- [Service Account Overview](https://docs.cloud.google.com/iam/docs/service-account-overview)

To create a `service account`:
```shell
$ gcloud iam service-accounts create <your-devpod-service-account> \
  --project <your-devpod-project-id>
```

Your `service account` gets an email address:
```text
<your-devpod-service-account>@<your-devpod-project-id>.iam.gserviceaccount.com
```

Generate and retrieve JSON `key` for the `service account`:
```shell
$ gcloud iam service-accounts keys create \
  /path/to/service/account/key.json \
  --iam-account=<your-devpod-service-account-email> \
  --project <your-devpod-project-id>
```

Grant to the `service account` the following IAM `roles` on the `project`:

```shell
# Compute Engine billing
$ gcloud projects add-iam-policy-binding <your-devpod-project-id> \
  --member="serviceAccount:<your-devpod-service-account-email>" \
  --role=roles/serviceusage.serviceUsageConsumer
# Compute Engine instance operations
$ gcloud projects add-iam-policy-binding <your-devpod-project-id> \
  --member="serviceAccount:<your-devpod-service-account-email>" \
   --role=roles/compute.instanceAdmin.v1
```

If you need access to Google Cloud services from _within_ the virtual machine
via a separate `service account` that you attach to Google Cloud VMs
using the provider's `SERVICE_ACCOUNT` option,
another `role` is needed: `iam.serviceAccountUser`;
for details, see
[Service Accounts for Instances](https://cloud.google.com/compute/docs/access/create-enable-service-accounts-for-instances).

## Using the Provider

This provider has the following `options`:

| NAME                      | REQUIRED | DESCRIPTION                                                             | DEFAULT                                              |
|---------------------------|----------|-------------------------------------------------------------------------|------------------------------------------------------|
| AGENT_PATH                | false    | The path where to inject the DevPod agent to.                           | /var/lib/toolbox/devpod                              |
| DISK_IMAGE                | false    | The disk image to use.                                                  | projects/cos-cloud/global/images/cos-101-17162-127-5 |
| DISK_SIZE                 | false    | The disk size to use (GB).                                              | 40                                                   |
| INACTIVITY_TIMEOUT        | false    | If defined, will automatically stop the VM after the inactivity period. | 5m                                                   |
| INJECT_DOCKER_CREDENTIALS | false    | If DevPod should inject docker credentials into the remote host.        | true                                                 |
| KEY                       | false    | Google Cloud JSON key.                                                  |                                                      |
| KEY_FILE                  | false    | Path to the Google Cloud JSON key file.                                 |                                                      |
| INJECT_GIT_CREDENTIALS    | false    | If DevPod should inject git credentials into the remote host.           | true                                                 |
| MACHINE_TYPE              | false    | The machine type to use.                                                | c2-standard-4                                        |
| NETWORK                   | false    | The network id to use.                                                  |                                                      |
| PROJECT                   | true     | The project id to use.                                                  |                                                      |
| PUBLIC_IP_ENABLED         | false    | Use a public ip to access the instance                                  | true                                                 |
| SERVICE_ACCOUNT           | false    | A service account to attach                                             |                                                      |
| SUBNETWORK                | false    | The subnetwork id to use.                                               |                                                      |
| TAG                       | false    | A tag to attach to the instance.                                        | devpod                                               |
| ZONE                      | true     | The Google Cloud zone to create the VM in, e.g. europe-west1-d          |                                                      |


You can supply `options` on the command line:
```shell
$ devpod provider add gcloud \
  -o PROJECT=<your-devpod-project-id> \
  -o ZONE=<Google Cloud zone to create the VMs in>
```

You can supply `options` via the `DEVPOD_PROVIDER_GCLOUD_` environment variables:
```shell
$ export DEVPOD_PROVIDER_GCLOUD_PROJECT=<your-devpod-project-id>
$ export DEVPOD_PROVIDER_GCLOUD_ZONE=<Google Cloud zone to create the VMs in>
$ devpod provider add gcloud
```

If a value for an option is supplied both on the command line
and via an environment variable,
value supplied on the command line takes precedence.

If you add the provider with a non-default name
```shell
$ devpod provider add gcloud --name my-gcloud
```
names of the environment variables that supply values for the options
change accordingly: `DEVPOD_PROVIDER_GCLOUD_PROJECT` becomes
`DEVPOD_PROVIDER_MY_GCLOUD_PROJECT`.

With the provider added, to start a workspace:

```sh
# with local sources
$ devpod up .
# with sources from a github repository
$ devpod up github.com/<user>/<repository>
```

You'll need to wait for the machine and workspace setup.

## CLI Workflow

When using CLI, it is convenient to use environment variables to set
values of options rather than supplying them on the command line,
and to use [direnv](https://direnv.net/)
to assign the environment variables;
with `.envrc` files kept in your `dotfiles` repository
managed by a [dotfiles utility](https://dotfiles.github.io/utilities/)
your environment becomes easily reproducible.

Here is an example of an `.envrc` with a Google Cloud configuration:
```env
export DEVPOD_CONTEXT="org1"
export DEVPOD_PROVIDER="gcloud"
export DEVPOD_PROVIDER_GCLOUD_KEY_FILE=/path/to/service/account/key.json
export DEVPOD_PROVIDER_GCLOUD_PROJECT="<your-devpod-project-id>"
export DEVPOD_PROVIDER_GCLOUD_ZONE="us-east4-b"
export DEVPOD_PROVIDER_GCLOUD_MACHINE_TYPE="c2-standard-8"
export DEVPOD_PROVIDER_GCLOUD_DISK_SIZE=20

export DEVPOD_IDE="intellij"
# suppress JetBrains "Trust" dialog when opening a workspace for the first time
export DEVPOD_WORKSPACE_ENV="REMOTE_DEV_TRUST_PROJECTS=1"
```

In any directory where this `.envrc` file is active,
you add the provider (once) with:
```shell
$ devpod provider add gcloud
```

and you are ready to spin up workspaces!

To be able to work with multiple Google Cloud organizations,
you need to configure similar `.envrc` files in different directories;
which organization your workspace will use is
determined by which `.envrc` file is active in the
directory you spin it up in.

DevPod `cli` supports `context`s, which scope providers added to them;
for multiple `.envrc` files to work, each must set
`DEVPOD_CONTEXT` environment variable to a different value.
