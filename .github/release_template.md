<!-- markdownlint-disable MD041 -->

### Prerequisites

- [ ] MAJOR: Check `docs/src/pages/migrations.md`
  - [ ] Check whether it contains all the necessary migration steps and recommended actions for users and administrators
  - [ ] Check whether the steps refer to the associated pull requests or issues
  - [ ] Ensure that the steps are clear and describe the actions required for the migration
    - Good: "Rename your `branch` configuration option to `when.branch` (PR#123)"
    - Bad: "Remove the `branch` configuration option in favor of `when.branch`"
    - If necessary, provide background information so that users can understand the change
- [ ] MAJOR: Create a blog entry in `docs/blog/` that highlights the most important changes and includes a link to the release notes.
- [ ] Prepare docs PR for new version and delete old versions (keep only the last three minor versions for the current major version)
  - [ ] Copy `docs/docs` to `docs/versioned_docs/version-<version>` and delete old versions
  - [ ] Create `docs/versioned_sidebars/version-<version>-sidebars.json` and delete old ones
  - [ ] Add new version to `docs/versions.json` and delete old versions
  - [ ] Add new version to the version list in `docs/src/pages/versions.md`
- [ ] Announce the release in the maintainer chat and ask for pending blockers

### Release

- [ ] Test the latest container images to make sure they work as expected
- [ ] Update `https://ci.woodpecker.org` to the latest version of `next` and verify that it works as expected
- [ ] Merge the release PR to start the release pipeline

### Post-release

- [ ] Merge the docs PR
- [ ] Announce release in relevant chats and on social media platforms
  - [ ] Mastodon (check if already posted from the release pipeline)
  - [ ] Matrix
  - [ ] Twitter
