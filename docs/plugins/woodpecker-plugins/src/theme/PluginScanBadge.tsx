import React, { useEffect, useState } from 'react';

import { WoodpeckerPlugin } from '../types';

type ScanFindings = {
  critical: number;
  high: number;
  medium: number;
  low: number;
  unknown: number;
  secrets: number;
  misconfigurations: number;
  total: number;
};

type ScanPluginResult = {
  docs: string;
  name: string;
  image?: string | null;
  image_digest?: string | null;
  image_digest_ref?: string | null;
  status: 'ok' | 'error' | string;
  scanned_at: string;
  duration_seconds: number;
  error?: string | null;
  details_url?: string | null;
  findings: ScanFindings;
};

type ScanReport = {
  schema_version: number;
  generated_at: string;
  pipeline_url?: string | null;
  scanner: {
    version?: string | null;
    binary_sha256?: string | null;
    db_version?: number | null;
    db_updated_at?: string | null;
  };
  plugins: ScanPluginResult[];
};

type ScanVulnerability = {
  target: string;
  id: string;
  package: string;
  installed_version: string;
  fixed_version?: string | null;
  severity: string;
  title: string;
  url?: string | null;
};

type ScanSecret = {
  target: string;
  rule_id: string;
  title: string;
  severity: string;
  match: string;
};

type ScanMisconfiguration = {
  target: string;
  id: string;
  title: string;
  severity: string;
  message: string;
  resolution: string;
};

type ScanDetails = {
  schema_version: number;
  name: string;
  docs: string;
  image?: string | null;
  image_digest?: string | null;
  scanned_at: string;
  status: string;
  error?: string | null;
  findings: ScanFindings;
  vulnerabilities: ScanVulnerability[];
  secrets: ScanSecret[];
  misconfigurations: ScanMisconfiguration[];
};

type PluginScanData = {
  report: ScanReport;
  result: ScanPluginResult;
};

// Published by the plugin-scan workflow to the website deployment repository.
// It is fetched at runtime so a nightly scan updates the badges without a docs
// rebuild. Preview deployments fall back to no badge if the file is missing.
const SCAN_RESULTS_URL = '/plugin-scan.json';

let scanResultsPromise: Promise<ScanReport | null> | null = null;

function loadScanResults(): Promise<ScanReport | null> {
  if (!scanResultsPromise) {
    scanResultsPromise = fetch(SCAN_RESULTS_URL, { cache: 'no-cache' })
      .then((response) => (response.ok ? (response.json() as Promise<ScanReport>) : null))
      .catch(() => null);
  }
  return scanResultsPromise;
}

function usePluginScan(plugin: WoodpeckerPlugin): PluginScanData | null {
  const [data, setData] = useState<PluginScanData | null>(null);

  useEffect(() => {
    let active = true;

    loadScanResults().then((report) => {
      if (!active || !report) {
        return;
      }
      const result = report.plugins.find((entry) => entry.docs === plugin.docsSource);
      if (result) {
        setData({ report, result });
      }
    });

    return () => {
      active = false;
    };
  }, [plugin.docsSource]);

  return data;
}

function severityRank(severity?: string | null): number {
  switch ((severity ?? 'UNKNOWN').toUpperCase()) {
    case 'CRITICAL':
      return 0;
    case 'HIGH':
      return 1;
    case 'MEDIUM':
      return 2;
    case 'LOW':
      return 3;
    default:
      return 4;
  }
}

function severityClass(severity?: string | null): string {
  switch ((severity ?? 'UNKNOWN').toUpperCase()) {
    case 'CRITICAL':
    case 'HIGH':
      return 'badge--danger';
    case 'MEDIUM':
      return 'badge--warning';
    case 'LOW':
      return 'badge--secondary';
    default:
      return 'badge--info';
  }
}

function scanTitle(data: PluginScanData): string {
  const { report, result } = data;
  const findings = result.findings;
  const lines = [
    `Scanned: ${new Date(result.scanned_at).toLocaleString()}`,
    `Findings: ${findings.critical} critical, ${findings.high} high, ${findings.medium} medium, ${findings.low} low`,
    `Scanner: Trivy ${report.scanner.version ?? 'unknown'}`,
  ];
  if (result.error) {
    lines.push(`Error: ${result.error}`);
  }
  if (report.pipeline_url) {
    lines.push(`Pipeline: ${report.pipeline_url}`);
  }
  return lines.join('\n');
}

export function PluginScanBadge({ plugin }: { plugin: WoodpeckerPlugin }) {
  const data = usePluginScan(plugin);
  if (!data) {
    return null;
  }

  const ok = data.result.status === 'ok';
  return (
    <span className={`badge ${ok ? 'badge--success' : 'badge--danger'} wp-plugin-scan-badge`} title={scanTitle(data)}>
      {ok ? '✓ scanned' : '✗ scan failed'}
    </span>
  );
}

function VulnerabilityTable({ vulnerabilities }: { vulnerabilities: ScanVulnerability[] }) {
  if (vulnerabilities.length === 0) {
    return <p>No vulnerabilities found.</p>;
  }

  const sorted = [...vulnerabilities].sort((a, b) => severityRank(a.severity) - severityRank(b.severity));

  return (
    <div className="wp-plugin-scan-table">
      <table>
        <thead>
          <tr>
            <th>Severity</th>
            <th>ID</th>
            <th>Package</th>
            <th>Installed</th>
            <th>Fixed</th>
            <th>Title</th>
          </tr>
        </thead>
        <tbody>
          {sorted.map((vulnerability, index) => (
            <tr key={`${vulnerability.id}-${vulnerability.package}-${vulnerability.target}-${index}`}>
              <td>
                <span className={`badge ${severityClass(vulnerability.severity)}`}>{vulnerability.severity}</span>
              </td>
              <td>
                {vulnerability.url ? (
                  <a href={vulnerability.url} target="_blank" rel="noopener noreferrer">
                    {vulnerability.id}
                  </a>
                ) : (
                  vulnerability.id
                )}
              </td>
              <td>{vulnerability.package}</td>
              <td>{vulnerability.installed_version}</td>
              <td>{vulnerability.fixed_version ?? '—'}</td>
              <td>{vulnerability.title}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function SecretsTable({ secrets }: { secrets: ScanSecret[] }) {
  if (secrets.length === 0) {
    return null;
  }

  return (
    <>
      <h4>Secrets</h4>
      <div className="wp-plugin-scan-table">
        <table>
          <thead>
            <tr>
              <th>Severity</th>
              <th>Rule</th>
              <th>Title</th>
              <th>Target</th>
            </tr>
          </thead>
          <tbody>
            {secrets.map((secret, index) => (
              <tr key={`${secret.rule_id}-${secret.target}-${index}`}>
                <td>
                  <span className={`badge ${severityClass(secret.severity)}`}>{secret.severity}</span>
                </td>
                <td>{secret.rule_id}</td>
                <td>{secret.title}</td>
                <td>{secret.target}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}

function MisconfigurationsTable({ misconfigurations }: { misconfigurations: ScanMisconfiguration[] }) {
  if (misconfigurations.length === 0) {
    return null;
  }

  return (
    <>
      <h4>Misconfigurations</h4>
      <div className="wp-plugin-scan-table">
        <table>
          <thead>
            <tr>
              <th>Severity</th>
              <th>ID</th>
              <th>Title</th>
              <th>Target</th>
            </tr>
          </thead>
          <tbody>
            {misconfigurations.map((misconfiguration, index) => (
              <tr key={`${misconfiguration.id}-${misconfiguration.target}-${index}`}>
                <td>
                  <span className={`badge ${severityClass(misconfiguration.severity)}`}>
                    {misconfiguration.severity}
                  </span>
                </td>
                <td>{misconfiguration.id}</td>
                <td>{misconfiguration.title}</td>
                <td>{misconfiguration.target}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}

function DetailedFindings({ url, total }: { url: string; total: number }) {
  const [details, setDetails] = useState<ScanDetails | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = () => {
    if (details || loading) {
      return;
    }

    setLoading(true);
    fetch(url, { cache: 'no-cache' })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        return response.json() as Promise<ScanDetails>;
      })
      .then((data) => setDetails(data))
      .catch(() => setError('Could not load the detailed scan results.'))
      .finally(() => setLoading(false));
  };

  return (
    <details
      className="wp-plugin-scan-findings"
      onToggle={(event) => {
        if (event.currentTarget.open) {
          load();
        }
      }}
    >
      <summary>Show detailed Trivy findings ({total})</summary>
      {loading && <p>Loading detailed results…</p>}
      {error && <p className="wp-plugin-scan-error">{error}</p>}
      {details && (
        <>
          <VulnerabilityTable vulnerabilities={details.vulnerabilities ?? []} />
          <SecretsTable secrets={details.secrets ?? []} />
          <MisconfigurationsTable misconfigurations={details.misconfigurations ?? []} />
        </>
      )}
    </details>
  );
}

export function PluginScanDetails({ plugin }: { plugin: WoodpeckerPlugin }) {
  const data = usePluginScan(plugin);
  if (!data) {
    return null;
  }

  const { report, result } = data;
  const ok = result.status === 'ok';
  const findings = result.findings;

  return (
    <div className={`wp-plugin-scan-details ${ok ? 'wp-plugin-scan-ok' : 'wp-plugin-scan-failed'}`}>
      <strong>{ok ? '✓ Security scan completed' : '✗ Security scan failed'}</strong>
      <div>
        Scanned {new Date(result.scanned_at).toLocaleString()} with Trivy {report.scanner.version ?? 'unknown'}
      </div>
      <div>
        {findings.critical} critical · {findings.high} high · {findings.medium} medium · {findings.low} low ·{' '}
        {findings.secrets} secrets · {findings.misconfigurations} misconfigurations
      </div>
      {result.image && (
        <div className="wp-plugin-scan-image">
          <code>{result.image}</code>
          {result.image_digest && <span> @ {result.image_digest}</span>}
        </div>
      )}
      {result.error && <div className="wp-plugin-scan-error">{result.error}</div>}
      {result.details_url && <DetailedFindings url={result.details_url} total={result.findings.total} />}
      {report.pipeline_url && (
        <a href={report.pipeline_url} target="_blank" rel="noopener noreferrer">
          View scan pipeline
        </a>
      )}
    </div>
  );
}
