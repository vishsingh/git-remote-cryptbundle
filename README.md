git-remote-cryptbundle
===============================

A (work-in-progress) implementation of the Git "Remote helper" protocol.

https://www.kernel.org/pub/software/scm/git/docs/gitremote-helpers.html

Pushes encrypted Git bundles to a remote that you believe may be readable by an adversary.

## Properties

- No information is ever lost. Git-remote-cryptbundle only pushes new bundles to the remote, and never
  touches old bundles.
  Even if you force-update a branch head, the old branch head is still there in previous bundles.

## Adding a cryptbundle remote

```
git remote add cryptorigin cryptbundle::ssh://username@my.awesome.server.com/path/to/cryptbundle
```
