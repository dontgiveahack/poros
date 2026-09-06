import { useEffect, useMemo, useState } from "react"

type Amount = { value: string; commodity: string }
type BalanceRow = { account: string; commodity: string; amount: Amount }
type Tx = {
	id: string
	date: string
	type: string
	title?: string
	amount?: Amount
	account?: string
	from?: string
	to?: string
	asset?: string
	quantity?: string
	price?: Amount
	category?: string
	tags?: string[]
}

const API = import.meta.env.VITE_API_URL ?? "http://localhost:8080"

function formatTx(r: Tx): string {
	if (r.amount) return `${r.amount.value} ${r.amount.commodity}`
	if (r.quantity && r.asset && r.price) return `${r.quantity} ${r.asset} @ ${r.price.value} ${r.price.commodity}`
	if (r.quantity && r.asset) return `${r.quantity} ${r.asset}`
	return "—"
}

function accountLabel(r: Tx): string {
	if (r.from && r.to) return `${r.from} → ${r.to}`
	return r.account ?? "—"
}

function matchesAccount(r: Tx, acct: string): boolean {
	if (!acct) return true
	return r.account === acct || r.from === acct || r.to === acct
}

function matchesSearch(r: Tx, q: string): boolean {
	if (!q) return true
	const hay = [r.title ?? "", r.category ?? "", r.asset ?? "", ...(r.tags ?? [])]
		.join(" ")
		.toLowerCase()
	return hay.includes(q.toLowerCase())
}

export default function App() {
	const [tab, setTab] = useState<"balances" | "transactions">("balances")
	const [rows, setRows] = useState<BalanceRow[] | null>(null)
	const [txs, setTxs] = useState<Tx[] | null>(null)
	const [error, setError] = useState<string | null>(null)

	// Filters
	const [fAccount, setFAccount] = useState("")
	const [fType, setFType] = useState("")
	const [fCategory, setFCategory] = useState("")
	const [fSearch, setFSearch] = useState("")

	useEffect(() => {
		Promise.all([
			fetch(`${API}/api/v1/balances`).then((r) => {
				if (!r.ok) throw new Error(`balances ${r.status}`)
				return r.json()
			}),
			fetch(`${API}/api/v1/transactions`).then((r) => {
				if (!r.ok) throw new Error(`transactions ${r.status}`)
				return r.json()
			})
		])
			.then(([b, t]) => {
				setRows(b)
				setTxs(t)
			})
			.catch((e) => setError(String(e)))
	}, [])

	// Option lists derived from data
	const accounts = useMemo(() => {
		const s = new Set<string>()
		for (const t of txs ?? []) {
			if (t.account) s.add(t.account)
			if (t.from) s.add(t.from)
			if (t.to) s.add(t.to)
		}

		return [...s].sort()
	}, [txs])

	const types = useMemo(() => {
		const s = new Set<string>()
		for (const t of txs ?? []) s.add(t.type)
		return [...s].sort()
	}, [txs])

	const categories = useMemo(() => {
		const s = new Set<string>()
		for (const t of txs ?? []) if (t.category) s.add(t.category)
		return [...s].sort()
	}, [txs])

	const filtered = useMemo(() => {
		return (txs ?? [])
			.filter((t) => matchesAccount(t, fAccount))
			.filter((t) => !fType || t.type === fType)
			.filter((t) => !fCategory || t.category === fCategory)
			.filter((t) => matchesSearch(t, fSearch))
			.slice()
			.sort((a, b) => b.date.localeCompare(a.date))
	}, [txs, fAccount, fType, fCategory, fSearch])

	const hasFilters = fAccount || fType || fCategory || fSearch
	const clearFilters = () => {
		setFAccount("")
		setFType("")
		setFCategory("")
		setFSearch("")
	}

	const inputStyle: React.CSSProperties = {
		padding: "0.4rem 0.5rem",
		border: "1px solid #ccc",
		borderRadius: 4,
		fontSize: "0.85rem",
	}

	return (
		<main style={{ fontFamily: "system-ui, sans-serif", padding: "2rem", maxWidth: 900, margin: "0 auto" }}>
			<h1>Póros Manager</h1>
			<p style={{ color: "#666" }}>
				API: <code>{API}</code> · <code>podman compose up</code> para levantar DB + API
			</p>

			<nav style={{ display: "flex", gap: "0.5rem", margin: "1rem 0" }}>
				<button
					onClick={() => setTab("balances")}
					style={{
						padding: "0.5rem 1rem",
						border: "1px solid #ccc",
						background: tab === "balances" ? "#111" : "#fff",
						color: tab === "balances" ? "#fff" : "#111",
						cursor: "pointer",
					}}
				>Balances</button>

				<button
					onClick={() => setTab("transactions")}
					style={{
						padding: "0.5rem 1rem",
						border: "1px solid #ccc",
						background: tab === "transactions" ? "#111" : "#fff",
						color: tab === "transactions" ? "#fff" : "#111",
						cursor: "pointer",
					}}
				>Transactions</button>
			</nav>

			{error && <p style={{ color: "crimson" }}>Error: {error}</p>}
			{!rows && !txs && !error && <p>Loading...</p>}

			{tab === "balances" && rows && (
				<table style={{ width: "100%", borderCollapse: "collapse" }}>
				    <thead>
					<tr style={{ textAlign: "left", borderBottom: "2px solid #ccc" }}>
					    <th>Account</th>
					    <th>Commodity</th>
					    <th style={{ textAlign: "right" }}>Amount</th>
					</tr>
				    </thead>
				    <tbody>
					{rows
					    .slice()
					    .sort((a, b) => a.account.localeCompare(b.account) || a.commodity.localeCompare(b.commodity))
					    .map((r) => (
						<tr key={`${r.account}:${r.commodity}`} style={{ borderBottom: "1px solid #eee" }}>
						    <td>{r.account}</td>
						    <td>{r.commodity}</td>
						    <td style={{ textAlign: "right" }}>
							{r.amount.value} {r.amount.commodity}
						    </td>
						</tr>
					    ))}
				    </tbody>
				</table>
			)}

			{tab === "transactions" && txs && (
			<>
				<div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap", marginBottom: "1rem" }}>
					<select value={fAccount} onChange={(e) =>setFAccount(e.target.value)} style={inputStyle} aria-label="Account">
						<option value="">All accounts</option>
						{accounts.map((a) => (
							<option key={a} value={a}>{a}</option>
						))}
					</select>
					<select value={fType} onChange={(e) => setFType(e.target.value)} style={inputStyle} aria-label="Type">
						<option value="">All types</option>
						{types.map((t) => (
							<option key={t} value={t}>{t}</option>
						))}
					</select>
					<select value={fCategory} onChange={(e) => setFCategory(e.target.value)} style={inputStyle} aria-label="Category">
						<option value="">All categories</option>
						{categories.map((c) => (
							<option key={c} value={c}>{c}</option>
						))}
					</select>
					<input
						value={fSearch}
						onChange={(e) => setFSearch(e.target.value)}
						placeholder="Search title, asset, tag..."
						style={{ ...inputStyle, minWidth: 200 }}
					/>
					{hasFilters && (
						<button onClick={clearFilters} style={{ ...inputStyle, cursor: "pointer" }}>
							Clear ✕
						</button>
					)}
				</div>

				<p style={{ color: "#666", fontSize: "0.85rem" }}>
					{filtered.length} of {txs.length} transactions
				</p>

				<table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.9rem" }}>
					<thead>
						<tr style={{ textAlign: "left", borderBottom: "2px solid #ccc" }}>
							<th>Date</th>
							<th>Type</th>
							<th>Title</th>
							<th>Amount</th>
							<th>Account</th>
							<th>Category</th>
						</tr>
					</thead>
					<tbody>
						{filtered.map((r) => (
							<tr key={r.id} style={{ borderBottom: "1px solid #eee" }}>
								<td style={{ whiteSpace: "nowrap" }}>{r.date.slice(0, 10)}</td>
								<td>{r.type}</td>
								<td>{r.title ?? "-"}</td>
								<td style={{ textAlign: "right", whiteSpace: "nowrap" }}>{formatTx(r)}</td>
								<td>{accountLabel(r)}</td>
								<td>{r.category ?? "-"}</td>
							</tr>
						))}
						{filtered.length === 0 && (
							<tr>
								<td colSpan={6} style={{ textAlign: "center", padding: "1rem", color: "#999" }}>
									No matching transactions
								</td>
							</tr>
						)}
					</tbody>
				</table>
			</>
			)}
		</main>
	)
}
