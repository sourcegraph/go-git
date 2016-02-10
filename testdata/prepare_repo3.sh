#!/bin/bash

# Description: Creates a repo with packed references with a tag "master-123"
#
# packed-refs contents:
# ```
# 8b61789a76de9edaa49b2529d3aaa302ba238c0b refs/tags/master
# ```

set -ex

export GIT_DIR=repo3
export GIT_AUTHOR_NAME="Test Author"
export GIT_AUTHOR_EMAIL="author@example.com"
export GIT_AUTHOR_DATE="Thu, 07 Apr 2005 22:13:13 +0200"
export GIT_COMMITTER_NAME="Test Committer"
export GIT_COMMITTER_EMAIL="committer@example.com"
export GIT_COMMITTER_DATE="Thu, 07 Apr 2005 22:13:14 +0200"

rm -rf $GIT_DIR

git init --bare

echo -n "test" | git hash-object -w --stdin # 30d74d258442c7c65512eafab474568dd706c430
echo -n "test unpacked" | git hash-object -w --stdin # d76bde4f5d1ed609dc82d8cd7d216d893830f1c9
git update-index --add --cacheinfo 100644 30d74d258442c7c65512eafab474568dd706c430 test.txt
git write-tree # 095a057d4a651ec412d06b59e32e9b02871592d5
git commit-tree -m "test commit" 095a057d4a651ec412d06b59e32e9b02871592d5 # 8b61789a76de9edaa49b2529d3aaa302ba238c0b

git update-ref refs/heads/master 8b61789a76de9edaa49b2529d3aaa302ba238c0b

set +x
for i in `seq 1 1000`; do
  obj=`echo -n $i | git hash-object -w --stdin`
  git update-index --add --cacheinfo 100644 $obj $i.txt
done
set -x

git repack
git prune-packed

git tag master
git pack-refs
