## OBS Submit Request GH Action

This action enables the triggering of a Submit Request in OBS, to bump
the version of a package.

### Usage

```yaml
jobs:
  submit-request:
    runs-on: ubuntu-latest
    container:
      image: opensuse/tumbleweed
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Set versioning and channel
      run: |
        local filter
        if [[ "$input_string" == *"production"* ]]; then
            echo "CHANNEL=stable" >> $GITHUB_ENV
            filter = "production"
        else
            echo "CHANNEL=dev" >> $GITHUB_ENV
            filter = "testing"
        fi

        OLD_VERSION="$(gh release list --repo ${REPO_NAME} --limit 1 --jq '.[].tagName | select(contains("production"))' --json tagName,createdAt)"
        
        echo "OLD_VERSION=${OLD_VERSION#v}" >> $GITHUB_ENV
      env:
        NEW_VERSION: ${{ github.ref_name }}
        GH_TOKEN: ${{ github.token }}
        REPO_NAME: ${{ inputs.github-repo }}

    - name: OBS submit request
      uses: actions/obs-submit-request@main
      with:
        new-version: ${{ github.ref_name }}
        old-version: ${{ env.OLD_VERSION }}
        package-name: rancher-selinux
        upstream-project: isv:rancher:${{ env.CHANNEL }}
        target-project: home:rst-bot:${{ env.CHANNEL }}:rancher-selinux
        source-ext: tar.gz
```

### Permissions

