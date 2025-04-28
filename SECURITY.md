# Security

## Reporting a vulnerability

To report a security issue, file a [Private Security Report](https://github.com/canonical/identity-platform-admin-ui/security/advisories/new)
with a description of the issue, the steps you took to create the issue, affected
versions, and, if known, mitigations for the issue.

The [Ubuntu Security disclosure and embargo policy](https://ubuntu.com/security/disclosure-policy)
contains more information
about what you can expect when you contact us and what we expect from you.

## CVE patching for OCI factory artifacts

When a CVE is reported we are bound to patch the existing OCI artifacts if within the EOL
maintenance window


based on when the artifact was published there are 2 different methods to operate


### before https://github.com/canonical/identity-platform-admin-ui/pull/452 merge

In this case OCI tags include the patch version of the application
To be able to fix the OCI image with the related tag we need to:

* checkout a new branch from the tag and name it `cve/<tag>`

```git
git switch --detach v1.19.0
git switch -c cve/v1.19.0
```

* apply [oci-factory workflow patch](https://github.com/canonical/identity-platform-admin-ui/commit/eb0b7859f2210c9e2ce500e397ae3da688fef4de)
* apply CVE patches (conventional commits won't trigger a release here, so using a chore/feat/fix won't make a difference)
* retag to the head of the branch and push the tag

```git
git tag -f v1.19.0
git push -f --tags origin v1.19.0
```

* let the machinery do its job



### after https://github.com/canonical/identity-platform-admin-ui/pull/452 merge


In this case OCI tags don't include the patch version anymore, we should be able to simply use the current workflows

Two cases are possible now:


#### latest release

If tag is the latest, making `fix` commits to patch the issue and then use the `release-please` flow as usual
That will trigger the usual release PR with a patch version change, OCI tag won't be affected and OCI cli will push
the `<major>.<minor>` with the following

```yaml
    - source: canonical/identity-platform-admin-ui
      commit: c80436a8d26abd33f2d1901ac59393fde69dd987
      directory: ./
      release:
        1.21-22.04:
            end-of-life: "2024-11-26T00:00:00Z"
            risks:
                - candidate
                - edge
```


#### previous release

In the case tag is not on the same minor the same process describe for pre #452 merge applies with some exceptions,


To be able to fix the OCI image with the related tag we need to:

* checkout a new branch from the tag and name it `cve/<tag>`

```git
git switch --detach v1.19.0
git switch -c cve/v1.19.0
```

* apply git patch below (to be changed soon) to avoid pushing to latest stable

```git
diff --git c/.github/workflows/publish.yaml w/.github/workflows/publish.yaml
index 31968d8..f2aa3e2 100644
--- c/.github/workflows/publish.yaml
+++ w/.github/workflows/publish.yaml
@@ -94,7 +94,6 @@ jobs:
           echo IMAGE_VERSION_CANDIDATE=$($YQ '.version | split(".").[0:2] | join(".")' rockcraft.yaml) >> $GITHUB_ENV
       - name: Release
         run: |
-          $OCI_FACTORY upload -y --release track=$IMAGE_VERSION_STABLE-22.04,risks=stable,eol=$EOL_STABLE
           $OCI_FACTORY upload -y --release track=$IMAGE_VERSION_CANDIDATE-22.04,risks=candidate,edge,eol=$EOL_CANDIDATE
         env:
           GITHUB_TOKEN: ${{ secrets.token }}
```


* apply CVE patches (conventional commits won't trigger a release here, so using a chore/feat/fix won't make a difference)
* retag to the head of the branch and push the tag

```git
git tag -f v1.19.0
git push -f --tags origin v1.19.0
```

* let the machinery do its job