package mysql

var _ = []struct {
	name string
	stmt string
}{
	{
		name: "alter-table-drop-repo-fallback",
		stmt: alterTableDropRepoFallback,
	},
	{
		name: "drop-allow-push-tags-deploys-columns",
		stmt: dropAllowPushTagsDeploysColumns,
	},
}

//
// 026_drop_repo_fallback_column.sql
//

var alterTableDropRepoFallback = `
ALTER TABLE repos DROP COLUMN repo_fallback
`

//
// 027_drop_allow_push_tags_deployments_columns.sql
//

var dropAllowPushTagsDeploysColumns = `
ALTER TABLE repos DROP COLUMN repo_allow_push, DROP COLUMN repo_allow_deploys, DROP COLUMN repo_allow_tags
`
