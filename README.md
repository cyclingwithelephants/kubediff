

## limitations
- rendered diffs don't show the colours properly, instead we've got a bunch of escape sequences like:
  ```
  �[36m@@ -1,17494 +1 @@
  �[0m�[31m-apiVersion: v1
  �[0m�[31m-kind: Namespace
  �[0m�[31m-metadata:
  �[0m�[31m-  name: argocd
  �[0m�[31m----
  ```
- the diff direction might be the wrong way around (Additions show as deletions and vice versa)
