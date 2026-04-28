// Minimal HTTP client for the Entropy Shear API.
//
// This SDK only wraps fetch — it never evaluates rules locally and never
// caches decisions. Every call hits the configured base URL.

export type Verdict = "Yes" | "No" | "Hold";

export type Operator =
  | "=="
  | "!="
  | ">"
  | "<"
  | ">="
  | "<="
  | "in"
  | "contains";

export interface Condition {
  field: string;
  operator: Operator;
  value: unknown;
}

export interface Rule {
  id: string;
  priority: number;
  condition: Condition;
  effect: Verdict;
  route?: string;
  reason: string;
}

export interface Policy {
  id: string;
  version: string;
  rules: Rule[];
  default_effect: Verdict;
  default_reason: string;
}

export type Facts = Record<string, unknown>;

export interface TraceItem {
  rule_id: string;
  evaluated: boolean;
  matched: boolean;
  detail: string;
}

export interface ShearResult {
  verdict: Verdict;
  applied_rule_id: string | null;
  route?: string;
  reason: string;
  trace: TraceItem[];
  signature: string;
  shear_id: string;
}

export interface LedgerRecord {
  shear_id: string;
  timestamp: string;
  policy_id: string;
  policy_version: string;
  input_hash: string;
  verdict: Verdict;
  applied_rule_id: string | null;
  trace_hash: string;
  previous_shear_hash: string;
  current_shear_hash: string;
}

export interface VerifyResult {
  ok: boolean;
  total: number;
  broken_at: number | null;
  latest_shear_id?: string;
  latest_hash?: string;
  detail?: string;
}

export interface HealthResult {
  ok: boolean;
  service: string;
  version: string;
}

export interface ClientOptions {
  baseUrl: string;
  /** Override fetch (e.g. for testing). Defaults to globalThis.fetch (Node 18+). */
  fetch?: typeof fetch;
  /** Optional AbortSignal applied to every request. */
  signal?: AbortSignal;
  /** Request timeout in milliseconds. Default: 10_000. */
  timeoutMs?: number;
}

export class EntropyShearError extends Error {
  readonly status: number;
  readonly code?: string;
  readonly detail?: string;
  constructor(status: number, body: { error?: string; detail?: string } | string) {
    const code = typeof body === "string" ? undefined : body.error;
    const detail = typeof body === "string" ? body : body.detail;
    super(`entropy-shear ${status}${code ? " " + code : ""}${detail ? ": " + detail : ""}`);
    this.name = "EntropyShearError";
    this.status = status;
    this.code = code;
    this.detail = detail;
  }
}

export class EntropyShearClient {
  private readonly baseUrl: string;
  private readonly fetchImpl: typeof fetch;
  private readonly signal?: AbortSignal;
  private readonly timeoutMs: number;

  constructor(opts: ClientOptions) {
    if (!opts?.baseUrl) {
      throw new Error("EntropyShearClient: baseUrl is required");
    }
    this.baseUrl = opts.baseUrl.replace(/\/+$/, "");
    this.fetchImpl = opts.fetch ?? globalThis.fetch;
    if (typeof this.fetchImpl !== "function") {
      throw new Error(
        "EntropyShearClient: fetch is not available. Pass opts.fetch or run on Node 18+."
      );
    }
    this.signal = opts.signal;
    this.timeoutMs = opts.timeoutMs ?? 10_000;
  }

  health(): Promise<HealthResult> {
    return this.request<HealthResult>("GET", "/health");
  }

  shear(req: { policy: Policy; facts: Facts }): Promise<ShearResult> {
    return this.request<ShearResult>("POST", "/shear", req);
  }

  getLedgerRecord(shearId: string): Promise<LedgerRecord> {
    if (!shearId) throw new Error("shearId required");
    return this.request<LedgerRecord>("GET", `/ledger/${encodeURIComponent(shearId)}`);
  }

  verifyLedger(): Promise<VerifyResult> {
    return this.request<VerifyResult>("GET", "/ledger/verify");
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const url = this.baseUrl + path;
    const ctl = new AbortController();
    const timer = setTimeout(() => ctl.abort(), this.timeoutMs);
    if (this.signal) {
      this.signal.addEventListener("abort", () => ctl.abort(), { once: true });
    }
    try {
      const res = await this.fetchImpl(url, {
        method,
        headers: body !== undefined
          ? { "Content-Type": "application/json", "Accept": "application/json" }
          : { "Accept": "application/json" },
        body: body !== undefined ? JSON.stringify(body) : undefined,
        signal: ctl.signal,
      });
      const text = await res.text();
      const parsed = text ? safeParse(text) : undefined;
      if (!res.ok) {
        throw new EntropyShearError(res.status, parsed ?? text);
      }
      return parsed as T;
    } finally {
      clearTimeout(timer);
    }
  }
}

function safeParse(s: string): unknown {
  try {
    return JSON.parse(s);
  } catch {
    return s;
  }
}
