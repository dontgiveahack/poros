// Thin fetch client over the Póros REST API (GET only for now).
// Every helper throws on non-2xx so callers handle errors in one place.
import type { BalanceRow, Basis, FireSummary, Portfolio, Tx } from "./types"

export const API = import.meta.env.VITE_API_URL ?? "http://localhost:8080"

async function get<T>(path: string): Promise<T> {
  const r = await fetch(`${API}${path}`)
  if (!r.ok) throw new Error(`${path} ${r.status} ${r.statusText}`)
  return r.json() as Promise<T>
}

export const fetchBalances = () => get<BalanceRow[]>("/api/v1/balances")
export const fetchTransactions = () => get<Tx[]>("/api/v1/transactions")
export const fetchPortfolio = (basis: Basis) => get<Portfolio>(`/api/v1/portfolio?basis=${basis}`)
export const fetchFire = () => get<FireSummary>("/api/v1/fire?simulate=1&runs=5000")
