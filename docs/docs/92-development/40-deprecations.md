# Deprecation Policy

## Pipeline Configuration Changes

Pipeline configuration (YAML syntax) changes follow a strict deprecation process to ensure users have sufficient time to migrate.

### Process Timeline

1. **Minor Version N.x - Add Deprecation Warning**
   - Linter shows a warning (not an error)
   - Old syntax remains functional
   - Documentation is updated to reflect the new syntax
   - Warning message includes guidance on required changes

2. **Major Version (N+1).0 - Warning Becomes Error**
   - Linter issues an error (pipeline fails)
   - Old syntax is no longer supported
   - Breaking change is documented in the migration guide
   - Users **must** update their configurations

3. **Minor Version (N+1).x - Code Cleanup**
   - Deprecated code paths are removed
   - Implementation is simplified/refactored
   - Parser no longer recognizes the old syntax

### Example

Old syntax: `secrets: [token]`
New syntax: `environment: { TOKEN: { from_secret: token } }`

- **v2.5.0:** Deprecation warning added in linter; both syntaxes work
- **v2.6-2.9:** Warning persists; both syntaxes remain functional
- **v3.0.0:** Linter error; old syntax fails (breaking change)
- **v3.1.0:** Deprecated code paths removed; parser simplified

### Implementation Checklist

When deprecating pipeline config syntax:

- [ ] Add linter warning in `/pipeline/frontend/yaml/linter/`
- [ ] Change json schema in `/pipeline/frontend/yaml/linter/schema`
- [ ] Add test cases for deprecated syntax
- [ ] Update documentation with new syntax
