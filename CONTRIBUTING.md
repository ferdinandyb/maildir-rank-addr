# Contribution Guidelines

You can either send me a patch with `git-send-email` to bence@ferdinandy.com
(until I setup a public inbox) or open a pull request on github.

## git-send-email

If you've never done so before, but interested in how it works, you can check out this [interactive tutorial](https://git-send-email.io/).

Before sending a patch, please configure the local clone with

```
git config format.subjectPrefix "PATCH mra"
```
.

## github pull request

Unless you have a super large and complex pull request, I will be rebasing the
PR without a merge commit. This means that

- each commit message must stand on it's own (see below);
- if any changes are requested to the PR, then the commits should be amended instead of pushing new commits.



## Commit messages

Please follow these general rules (adopted somewhat from the aerc contribution guidelines):

- Limit the first line (title) of the commit message to 60 characters. Leave an
- empty space between the title and the body of the commit.  Use a short prefix
- for the commit title for readability with `git log --oneline`. Do not use the
  `fix:` nor `feature:` prefixes. See recent commits for inspiration.
- Only use lower case letters for the commit title except when quoting symbols
  or known acronyms.
- Use the body of the commit message to actually explain what your patch does
  and why it is useful. Even if your patch is a one line fix, the description
  is not limited in length and may span over multiple paragraphs. Use proper
  English syntax, grammar and punctuation. Make sure the body of the commit is
  hard wrapped at 72 characters.
- Address only one issue/topic per commit.  Describe your changes in imperative
- mood, e.g. *"make xyzzy do frotz"*
  instead of *"[This patch] makes xyzzy do frotz"* or *"[I] changed xyzzy to do
  frotz"*, as if you are giving orders to the codebase to change its behaviour.
- If you are fixing an issue, add a `Fixes: #xxx` trailer with the issue id.
- If you are fixing a regression introduced by another commit, add a `Fixes:`
  trailer with the commit id and its title.
- When in doubt, follow the format and layout of the recent existing commits
  (well, at least the later ones :)).

For why and further references, please see [my blog](https://bence.ferdinandy.com/gitcraft/).


## Formatting

Format the code with [gofumpt](https://github.com/mvdan/gofumpt).

## Testing

Please run test with `go test`. If possible, write tests as well as part of
your commits. I call tests "e2e" that start out by reading the email files in
the testdata folder and do asserts only after the final dataset has been put
together (i.e. the output of `calculateRanks`). Unit tests of specific
functions should go into a different file, e.g. like `parseMail_test.go`.
