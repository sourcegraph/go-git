git
=====

Pure Go read access to a git repository.

**State**: Actively maintained, used in production site, but without warranty, of course.<br>
**Maturity level**: 4/5 (works well in all tested repositories, expect API change, few corner cases not implemented yet)<br>
**License**: Free software (MIT License)<br>
**Installation**: Just run `go get github.com/gogits/git`<br>
**API documentation**: http://gowalker.org/github.com/gogits/git<br>
**Dependencies**: None<br>
**Contribution**: We like to get any kind of feedback (success stories, bug reports, merge requests, ...)

Example
-------

Sample application to list the latest directory (recursively):

```Go
package main

import (
    git "github.com/gogits/git"
    "log"
    "os"
    "path"
    "path/filepath"
)

func walk(dirname string, te *gogit.TreeEntry) int {
    log.Println(path.Join(dirname, te.Name))
    return 0
}

func main() {
    wd, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }
    repository, err := git.OpenRepository(filepath.Join(wd, "src/github.com/gogits/git/_testdata/testrepo.git"))
    if err != nil {
        log.Fatal(err)
    }
    ref, err := repository.LookupReference("HEAD")
    if err != nil {
        log.Fatal(err)
    }
    ci, err := repository.LookupCommit(ref.Oid)
    if err != nil {
        log.Fatal(err)
    }
    ci.tree.Walk(walk)
}
```
