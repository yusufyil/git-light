
# Git Light

Git Light(git-light), is an alternative version control system that provides basic functionalities like git.

You can add your changes to stage, create new commits from staged files and checkout to specific commits in history at any time.

All basic functionalities mentioned above can be used like using git. However this project is far away from being complete because I have developed this application for POC purposes in my graduation project, after graduation I am aiming to add other critical features like merging and hooks mechanisms.

After all I am not linus torvalds and this vcs is not going to replace git or something like that, so please forgive any unexpected behaviour.


## Installation

In order to install project, you need to install go 1.21 already installed on your computer.

```bash
  go build -o git-light
```

## Features

- git-light init

- git-light add test.txt
- git-light add *

- git-light commit -m"your commit message"

- git-light checkout feature/branch
- git-light checkout 969d6c6ef54ec390afe45d60277ef8e777e82c39
- git-light checkout HEAD~2

- git-light branch branchNameToBeCreated
- git-light branch -d branchNameToBeDeleted
- git-light branch -a

- git-light log


## Lessons Learned

Currently, I am writing an article about this part. So I will update here later.
