# Manual pipelines

Users with push access can start a manual pipeline from a repository's pipeline page. Choose the source that Woodpecker should use:

- **Branch** runs the current head commit of the selected branch.
- **Tag** runs the commit referenced by the selected tag.
- **Commit** runs a full or short commit SHA.

Additional pipeline variables can be supplied when triggering the run.

## Workflow event

Manual pipeline runs always use the `manual` event, including when their source is a tag. Select workflows with `when.event: manual`; choosing a tag does not select workflows that only use `when.event: tag`.
