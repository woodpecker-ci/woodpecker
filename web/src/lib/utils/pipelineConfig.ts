// Matches a `commands:` block and captures its indented list items.
// Group 1: the base indentation of the `commands:` key.
// Group 2: all subsequent lines that belong to the list (same indent + deeper).
const commandsBlockRegex = /^(\s+)commands:\s*\n((?:\1\s+-\s.+\n?)*)/gm;

// Matches a single YAML list item: optional whitespace, dash, space, content.
const commandLineRegex = /^\s*-\s(.+)$/gm;

// Characters that are special in a RegExp and must be escaped.
const specialCharsRegex = /[.*+?^${}()|[\]\\]/g;

// After escaping, a variable \${VAR} looks like \$\{WORD\} — replace with .* wildcard.
const matrixVariableRegex = /\\\$\\\{\w+\\\}/g;

/**
 * Parse one decoded pipeline config YAML and return a list of RegExp matchers,
 * one per shell command found inside `commands:` blocks.
 *
 * Matrix variables (`${VAR}`) are turned into `.*` wildcards so that a command
 * like `echo ${VERSION}` still matches its runtime log line.
 *
 * Patterns that degenerate to pure wildcards (e.g. a bare `${VAR}` entry) are
 * dropped — they would match every log line and produce false positives.
 */
export function extractCommandMatchers(decodedYaml: string): RegExp[] {
  const patterns: RegExp[] = [];

  for (const block of decodedYaml.matchAll(commandsBlockRegex)) {
    const blockContent = block[2];
    for (const match of blockContent.matchAll(commandLineRegex)) {
      const rawCommand = match[1].trim();

      const patternString = rawCommand
        .replace(specialCharsRegex, '\\$&') // escape regex special chars
        .replace(matrixVariableRegex, '.*'); // replace escaped \${VAR} with .*

      // Skip patterns that collapsed entirely to wildcards — they match everything.
      if (/^(\.\*)*$/.test(patternString)) {
        continue;
      }

      patterns.push(new RegExp(`^${patternString}$`));
    }
  }

  return patterns;
}
