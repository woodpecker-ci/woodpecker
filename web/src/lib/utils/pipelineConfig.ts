import { load } from 'js-yaml';

// Characters that are special in a RegExp and must be escaped.
const specialCharsRegex = /[.*+?^${}()|[\]\\]/g;

// After escaping, a matrix variable ${VAR} looks like \$\{WORD\} — replace with .* wildcard.
const matrixVariableRegex = /\\\$\\\{\w+\\\}/g;

// Shape of the parts of a Woodpecker pipeline config we care about.
interface StepConfig {
  name?: string;
  commands?: string | string[];
}

interface PipelineConfigYaml {
  steps?: Record<string, StepConfig> | StepConfig[];
}

/**
 * Turn a raw config string into an anchored regex source, replacing matrix
 * variables (`${VAR}`) with `.*` wildcards. Matrix variables are interpolated
 * before the pipeline runs, so the stored config keeps `${VAR}` while the
 * runtime value (log line / step name) is concrete — the wildcard bridges that.
 */
function toPatternSource(raw: string): string {
  return raw
    .replace(specialCharsRegex, '\\$&') // escape regex special chars
    .replace(matrixVariableRegex, '.*'); // replace escaped \${VAR} with .*
}

/**
 * Build a RegExp matching a single shell command.
 *
 * Returns null for patterns that collapse to pure wildcards (e.g. a bare `${VAR}`
 * command) — those would match every log line and produce false positives.
 */
function commandToMatcher(rawCommand: string): RegExp | null {
  const patternString = toPatternSource(rawCommand.trim());

  if (/^(?:\.\*)*$/.test(patternString)) {
    return null;
  }

  return new RegExp(`^${patternString}$`);
}

/**
 * Build a RegExp matching a step name. Unlike commands, a pure-wildcard name
 * (a step named entirely by a matrix variable) is kept on purpose: it matches
 * every runtime step, which is the desired "can't disambiguate, so apply to
 * all" fallback.
 */
function stepNameToMatcher(rawName: string): RegExp {
  return new RegExp(`^${toPatternSource(rawName.trim())}$`);
}

/**
 * Normalize the `steps` node (map form or list form) into a flat list of steps,
 * using the map key as the name in map form.
 */
function normalizeSteps(steps: Record<string, StepConfig> | StepConfig[]): StepConfig[] {
  if (Array.isArray(steps)) {
    return steps;
  }
  return Object.entries(steps).map(([name, config]) => ({ ...(config ?? {}), name }));
}

function toCommandList(commands: string | string[] | undefined): string[] {
  if (commands === undefined || commands === null) {
    return [];
  }
  return Array.isArray(commands) ? commands : [commands];
}

/**
 * Collect the `commands` of every config step whose (wildcard-aware) name
 * matches the given runtime step name. Returns an empty array if the YAML is
 * invalid or no matching step has commands.
 */
function getStepCommands(decodedYaml: string, stepName: string): string[] {
  let doc: PipelineConfigYaml | undefined;
  try {
    doc = load(decodedYaml) as PipelineConfigYaml;
  } catch {
    return [];
  }

  const steps = doc?.steps;
  if (steps === undefined || steps === null) {
    return [];
  }

  const commands: string[] = [];
  for (const step of normalizeSteps(steps)) {
    if (step.name === undefined || step.name === null) {
      continue;
    }
    if (stepNameToMatcher(step.name).test(stepName)) {
      commands.push(...toCommandList(step.commands));
    }
  }
  return commands;
}

/**
 * Return RegExp matchers for the commands of the step(s) matching `stepName`
 * within a pipeline config, used to detect which log lines start a new
 * (collapsible) command group.
 *
 * Both step names and commands support matrix-variable wildcards. Scoping to
 * the matching step(s) avoids false matches from unrelated steps; YAML
 * anchors/aliases are resolved by the parser.
 */
export function extractCommandMatchers(decodedYaml: string, stepName: string): RegExp[] {
  return getStepCommands(decodedYaml, stepName)
    .map((command) => commandToMatcher(String(command).trim()))
    .filter((matcher): matcher is RegExp => matcher !== null);
}

/**
 * Extract the command string from a shell trace line (`set -x` output).
 *
 * Bash and POSIX shells prefix each traced command with `+ `. Some backends
 * (e.g. the Windows local backend) additionally wrap the command in single or
 * double quotes: `+ 'net use'`. This function strips both the `+ ` prefix and
 * any surrounding matching quote pair so the result can be compared against the
 * plain command strings from the pipeline config.
 *
 * Returns null if the line is not a trace line (does not start with `+ `).
 */
export function extractCmdFromTrace(line: string): string | null {
  if (!line.startsWith('+ ')) {
    return null;
  }

  const raw = line.slice(2).trim();

  // Strip surrounding matching single or double quotes added by some backends.
  const quoteMatch = /^(['"])([\s\S]*)\1$/.exec(raw);
  return quoteMatch !== null ? quoteMatch[2] : raw;
}
