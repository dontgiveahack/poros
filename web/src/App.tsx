import { useEffect, useState } from "react"

type Amount = { value: string; commodity: string }
type BalanceRow = { account: string; commodity: string; amount: Amount }

const API = import.meta.env.VITE_API_URL ?? "http://localhost:8080"

export default function App() {
	const [rows, setRows] = useState<BalanceRow[] | null>(null)
	const [error, setError] = useState<string | null>(null)

	useEffect(() => {
		fetch(`${API}/api/v1/balances`)
			.then((r) => {
				if (!r.ok) throw new Error(`${r.status} ${r.statusText}`)
				return r.json()
			})
			.then(setRows)
			.catch((e) => setError(String(e)))
	}, [])

	return (
		<main style={{ fontFamily: "system-ui, sans-serif", padding: "2rem", maxWidth: 640 }}>
			<h1>Póros Manager</h1>
			<p>Balances from <code>{API}/api/v1/balances</code></p>

			{error && <p style={{ color: "crimsom" }}>Error: {error}</p>}
			{rows === null && !error && <p>Loading...</p>}

			{rows && (
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
									<td style={{ textAlign: "right" }}>{r.amount.value} {r.amount.commodity}</td>
								</tr>
							))
						}
					</tbody>
				</table>
			)}

			<p style={{ marginTop: "2rem", color: "#666", fontSize: "0.85rem" }}>
				Run <code>poros serve --data ./data</code> in another terminal.
			</p>
		</main>
	)
}
