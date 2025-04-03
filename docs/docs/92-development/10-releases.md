# Releasing a New Version

Following semantic versioning (semver) guidelines, breaking changes will trigger a major version update.

## Release Process

### Pre-Release

1. Update the version section in `CHANGELOG.md`.
2. Check `docs/src/pages/migrations.md`:
   - It has to contain all necessary migration steps and suggested actions for users and instance admins.
   - Check if steps link to the related pull requests or issues.
   - Ensure steps are clear and describe the actions needed to migrate.
      - Good: "Rename your `branch` config option to `when.branch`. (PR#123)"
      - Bad: "Removed `branch` config option in favour of `when.branch`."
      - Provide background if necessary to allow users to understand the change.
2. Add the new version to the versions list in `docs/src/pages/versions.md`.
3. For major releases, create a new blog post in `docs/blog/`, highlighting key changes with a link to the release notes.
4. Ask other maintainers in the chat for any outstanding blockers.
5. Schedule the release with at least 48 hours' notice in the maintainers chat.

### Release

1. Test the latest container images to ensure they work as expected.
2. Update `ci.woodpecker.org` to the latest `next` version and verify it works as expected.
3. Publish the new release on GitHub.

### Post-Release

1. Announce the new release in relevant chats and on social media platforms:
   - [ ] Mastodon (verify if already posted by the release pipeline)
   - [ ] Discord
   - [ ] Matrix
   - [ ] Twitter
