## OBS Checkout

This action branches an upstream package in OBS, and checks it out.

### Usage

```yaml
jobs:
  package-checkout:
    runs-on: ubuntu-latest
    container:
      image: opensuse/tumbleweed
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Checkout OBS package
      uses: actions/obs-checkout@main
      with:
        package-name: rancher-selinux
        upstream-project: isv:rancher:stable
        target-project: home:rst-bot:stable:rancher-selinux
        obs-config: ${{ secrets.OBS_CONFIG }}
```

The example above ensures that `home:rst-bot:stable:rancher-selinux` exists
and contains the latest version of `rancher-selinux` from the upstream project `isv:rancher:stable`.

### Permissions

The `obs-config` must have permissions to update/create the target project.
An example of an OBS config:

```ini
[general]
apiurl=https://api.opensuse.org

[https://api.opensuse.org]
user=<user>
pass=<passwd>
trusted_prj=openSUSE:Tumbleweed devel:languages:go
```

Refer to `oscrc(5)` man page for the full list of OBS config options.
