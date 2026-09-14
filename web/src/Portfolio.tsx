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
  market_value: Amount
}

export type Portfolio = {
  positions: Position[]
  total_cost: Amount
  total_market?: Amount
  basis: string
}

export type Basis = "cost" | "market"

export function PortfolioTab({ data, basis, onBasis }: { data: Portfolio | null; basis: Basis; onBasis: (b: Basis) => void }) {
  if (!data) return <p>Loading...</p>

  if (data.positions.length === 0) {
    return <p style={{ color: "#666" }}>No open positions.</p>
  }

  const useMarket = basis === "market"
  const total = (useMarket && data.total_market ? data.total_market : data.total_cost)
  const totalNum = parseFloat(total.value) || 0
  const valueOf = (p: Position) => (useMarket && p.market_value ? p.market_value : p.cost_value)

  return (
    <>
      <div style={{ display: "flex", gap: "0.5rem", alignItems: "center", marginBottom: "1rem" }}>
        {(["cost", "market"] as Basis[]).map((b) => (
          <button
            key={b}
            onClick={() => onBasis(b)}
            style={{
              padding: "0.4rem 0.8rem",
              border: "1px solid #ccc",
              cursor: "pointer",
              background: basis === b ? "#111" : "#fff",
              color: basis === b ? "#fff" : "#111"
            }}
          >
            {b === "cost" ? "At cost" : "At market"}
          </button>
        ))}
        <span>Total: <strong>{total.value} {total.commodity}</strong></span>
      </div>
      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.9rem" }}>
        <thead>
          <tr style={{ textAlign: "left", borderBottom: "2px solid #ccc" }}>
            <th>Asset</th>
            <th>Account</th>
            <th style={{ textAlign: "right" }}>Qty</th>
            <th style={{ textAlign: "right" }}>{useMarket ? "Market price" : "Last buy"}</th>
            <th style={{ textAlign: "right" }}>Value</th>
            <th style={{ textAlign: "right" }}>Weight</th>
          </tr>
        </thead>
        <tbody>
          {data.positions.map((p) => {
            const v = valueOf(p)
            const w = totalNum > 0 ? (parseFloat(v.value) / totalNum) * 100 : 0
            return (
              <tr key={`${p.account}:${p.asset}`} style={{ borderBottom: "1px solid #eee" }}>
                <td><strong>{p.asset}</strong></td>
                <td>{p.account}</td>
                <td style={{ textAlign: "right" }}>{p.quantity.value}</td>
                <td style={{ textAlign: "right" }}>{useMarket && p.market_value ? `${(parseFloat(p.market_value.value) / parseFloat(p.quantity.value)).toFixed(2)} ${p.market_value.commodity}` : `${p.price.value} ${p.price.commodity}`}</td>
                <td style={{ textAlign: "right" }}>{v.value} {v.commodity}</td>
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
      <p style={{ color: "#666", fontSize: "0.85rem" }}>
        {useMarket ? "Market: latest quote in prices.json (fallback: cost)." : "Cost: last buy price. Switch to market for quotes."}
      </p>
    </>
  )
}
