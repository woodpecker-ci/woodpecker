### Prerequisites

- [ ] MAJOR: Check `docs/src/pages/migrations.md`
  - [ ] It has to contain all necessary migration steps and suggested actions for users and instance admins
  - [ ] Check if steps link to the related pull requests or issues
  - [ ] Ensure steps are clear and describe the actions needed to migrate
    - Good: "Rename your `branch` config option to `when.branch` (PR#123)"
    - Bad: "Removed `branch` config option in favour of `when.branch`"
    - Provide background if necessary to allow users to understand the change
- [ ] MAJOR: create a blog post in `docs/blog/`, highlighting key changes with a link to the release notes
- [ ] Prepare docs PR for new version and delete old versions (only keep last three minor versions for current major version)
  - [ ] Copy `docs/docs` to `docs/versioned_docs/version-<version>` and delete old old ones
  - [ ] Create `docs/versioned_sidebars/version-<version>-sidebars.json` and delete old old ones
  - [ ] Add new version to `docs/versions.json` and delete old old ones
  - [ ] Add new version to the versions list in `docs/src/pages/versions.md`
- [ ] Announce the release in the maintainers chat and ask for any outstanding blockers

### Release

- [ ] Test the latest container images to ensure they work as expected
- [ ] Update `https://ci.woodpecker.org` to the latest `next` version and verify it works as expected
- [ ] Merge the Release PR to trigger the release pipeline

### Post-Release

- [ ] Merge docs PR
- [ ] Announce release in relevant chats and on social media platforms
  - [ ] Mastodon (verify if already posted by the release pipeline)
  - [ ] Discord
  - [ ] Matrix
  - [ ] Twitter
