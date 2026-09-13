export type Amount = {
	value: string;
	commodity: string
}

export type Position = {
	asset: string
	account: string
	quantity: Amount
	price: Amount
	cost_value: Amount
}

export type Portfolio = {
	positions: Position[];
	total_coust: Amount;
}

export function PortfolioTab({ data }: { data: Portfolio | null }) {
	if (!data) return <p>Loading...</p>

	if (data.positions.length === 0) {
		return <p style={{ color: "#666" }}>No open positions - registra un <code>.buy</code> en <code>data/transactions.json</code>.</p>
	}

	const total = parseFloat(data.total_cost.value) || 0

	return (
		<>
			<p>Total at cost: <strong>{data.total_cost.value} {data.total_cost.commodity}</strong></p>
			<table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.9rem" }}>
				<thead>
					<tr style={{ textAlign: "left", borderBottom: "2px solid #ccc" }}>
						<th>Asset</th>
						<th>Account</th>
						<th style={{ textAlign: "right" }}>Qty</th>
						<th style={{ textAlign: "right" }}>Last price</th>
						<th style={{ textAlign: "right" }}>Cost value</th>
						<th style={{ textAlign: "right" }}>Weight</th>
					</tr>
				</thead>
				<tbody>
					{data.positions.map((p) => {
						const w = total > 0 ? (parseFloat(p.cost_value.value) / total) * 100 : 0

						return (
							<tr key={`${p.account}:${p.asset}`} style={{ borderBottom: "1px solid #eee" }}>
								<td><strong>{p.asset}</strong></td>
								<td>{p.accunt}</td>
								<td style={{ textAlign: "right" }}>{p.quantity.value}</td>
								<td style={{ textAlign: "right" }}>{p.price.value} {p.price.commodity}</td>
								<td style={{ textAlign: "right" }}>{p.cost_value.value} {p.cost_value.commodity}</td>
								<td style={{ textAlign: "right" }}>
									<span style={{ display: "inline-block", width: 60, height: 8, background: "#eee", marginRight: 8, verticalAlign: "middle" }}>
										<span style={{ display: "inline-block", width: `${w}%`, height: 8, background: "#111" }} />
									</span>
									{w.toFixed(1)}%
								</td>
							</tr>
						)
					})}
				</tbody>
			</table>
			<p style={{ color: "#666", fontSize: "0.85rem" }}>Valued at last buy price (cost). Market valuation arrives with prices.</p>
		</>
	)
}
