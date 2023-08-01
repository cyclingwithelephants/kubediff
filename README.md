# Kubediff

[No, not that kubediff](https://github.com/weaveworks/kubediff) - naming is hard ok? 
If you've got a better idea for a name, I'm all ears.

Due to the usage of helm, kustomize, etc PRs contain diffs only of pre-rendered yamls, which can make it difficult to evaluate changes to your GitOps repo. 
This tool is designed to help with that by building the changed manifests and writing a git-style diff as one or more comments to your PR.

## Usage
### Configuration
kubediff accepts the following environment variables:

| Environment Variable |                             Description                             | Default Value |
|:---:|:-------------------------------------------------------------------:|:---:|
| `ENVS_DIR` |             The directory to the environments/clusters              | N/A |
| `GLOB_LEVELS` | The number of levels to glob in search for kustomization.yaml files | N/A |
| `GITHUB_OWNER` |                          The GitHub owner                           | N/A |
| `GITHUB_REPO` |                        The GitHub repository                        | N/A |
| `GITHUB_PR_NUMBER` |                     The number of the GitHub PR                     | N/A |
| `GITHUB_TOKEN` |                          The GitHub token                           | N/A |
| `DIFF_WITH_COLOUR` |                Boolean flag to show diff with colour                | `"true"` |
| `DIFF_CONTEXT_LINES` |         The (integer) number of context lines for the diff          | `"3"` |
| `PR_BRANCH_DIR` |                   The directory for the PR branch                   | `"pr"` |
| `TARGET_BRANCH_DIR` |                 The directory for the target branch                 | `"target"` |
| `TEMP_PATH` |                    The path for temporary files                     | `"tmp"` |
### Github Actions
[My personal live example](https://github.com/cyclingwithelephants/cloudlab/blob/main/.github/workflows/kubediff.yml)

This is a simple example workflow file that you could use in your repo
```yaml
# .github/workflows/kubediff.yaml
name: Kube diff

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

env:
  GOPATH: ${{ github.workspace }}/go

permissions:
  contents: read
  pull-requests: write

jobs:
  kube-diff:
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    steps:
    - name: Checkout PR branch
      uses: actions/checkout@v3
      with:
        path: ${{ env.PR_DIR }}

    - name: Checkout Target branch
      uses: actions/checkout@v3
      with:
        path: ${{ env.TARGET_DIR }}
        ref: ${{ github.event.pull_request.base.ref }}

    - name: setup kustomize
      shell: bash
      run: curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash

    - uses: actions/setup-go@v4
      with:
        go-version: 1.20.1

    - run: go run github.com/cyclingwithelephants/kubediff/cmd@main
      env:
        GLOB_LEVELS: 3
        ENVS_DIR: manifests/groups
        GITHUB_OWNER: ${{ github.repository_owner }}
        GITHUB_REPO: ${{ github.event.repository.name }}
        GITHUB_PR_NUMBER: ${{ github.event.pull_request.number }}
        GITHUB_TOKEN: ${{ github.token }}
```

## Limitations and Assumptions
- This only deals with kustomize. 
  The good news is that this might not be the large downside it seems.
  
  Kustomize supports: 
  - [helm charts](https://github.com/kubernetes-sigs/kustomize/blob/master/examples/chart.md)
  - bare yaml
  - remote yaml links
  - kustomize!

- Large diffs are a bit painful.
  Github has a max comment size and it's fairly easy to go over that. 
  Rendered diffs are chunked to < `githubMaxCommentSize` and multiple comments are made instead.
  It's a bit ugly, but I don't think I can improve on it much.

- I make assumptions about how you organise your gitops repo: 
  - You have a directory structure with a flat group of environments/clusters all under a single path:
  ```bash
  # e.g.
  ${ENVS_DIR}
    ├── env-c
    ├── env-b
    └── env-c
  ```
  - Those environment might have a tree of yaml, organised however you'd like, and you can recurse down that tree up to `${MAX_GLOBS}` levels to find the root kustomize.yaml that represents each manifest group.
    Note that `cluster` below is one level shallower than `apps` and `addons` - this is allowed since the tool checks that the parent directory doesn't contain a `kustomize.yaml` file.
  ```bash
  env-a # example environment
  ├── addons
  │   ├── argocd
  │   └── ingress-nginx
  ├── apps
  │   ├── app-a
  │   └── app-b
  └── cluster
      ├── kustomization.yaml # for example if you use CAPI
      └── patches
  ```
  - Changes made outside the `${ENVS_DIR}` directory won't be picked up. For example, you might have a `base` directory that the environmental directories inherit from.
    I assume that changes to e.g. a `base` directory imply changes to an environmental directory too, therefore picking up the changes in the environmental directory might be sufficient.
  
  - This tool exec's a kustomize binary in your ${PATH}.
