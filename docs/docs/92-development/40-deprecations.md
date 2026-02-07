# Deprecation Policy

## Pipeline Configuration Changes

Pipeline configuration (YAML syntax) changes follow a strict deprecation process to give users time to migrate.

### Process Timeline

1. **Minor Version N.x - Add Deprecation Warning**
   - Linter shows warning (not error)
   - Old syntax still works
   - Documentation updated to show new syntax
   - Warning message explains what to change

2. **Major Version (N+1).0 - Warning Becomes Error**
   - Linter shows error (pipeline fails)
   - Old syntax no longer works
   - Breaking change documented in migration guide
   - Users MUST update their configs

3. **Minor Version (N+1).x - Code Cleanup**
   - Remove deprecated code paths
   - Simplify/refactor implementation
   - Parser no longer recognizes old syntax

### Example

Old syntax: `secrets: [token]`
New syntax: `environment: { TOKEN: { from_secret: token } }`

- v2.5.0: Add deprecation warning in linter, both work
- v2.6-2.9: Warning continues, both still work
- v3.0.0: Error in linter, old syntax fails (BREAKING)
- v3.1.0: Remove old code paths, simplify parser

### Implementation Checklist

When deprecating pipeline config syntax:

- [ ] Add linter warning in `/pipeline/frontend/yaml/linter/`
- [ ] Change json schema in `/pipeline/frontend/yaml/linter/schema`
- [ ] Add test cases for deprecated syntax
- [ ] Update documentation with new syntax
