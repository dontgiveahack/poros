export type Amount = { value: string; commodity: string }

export type FireSummary = {
  year: number
  net_worth: Amount
  annual_income: Amount
  annual_expenses: Amount
  savings_rate: number
  annual_savings: Amount
  fire_number: Amount
  years_to_fire: number
  coast_fire?: Amount
  coast_progress?: number
  lean_fire?: Amount
  fat_fire?: Amount
  simulation?: {
    runs: number
    years: number
    p10: Amount
    p50: Amount
    p90: Amount
    prb_fire: number
  }
}

function Card({ label, value, sub }: {
  label: string;
  value: string;
  sub?: string;
}) {
  return (
    <div style={{ border: "1px solid #ddd", borderRadius: 8, padding: "0.8rem 1rem", minWidth: 150, flex: "1 1 150px" }}>
      <div style={{ fontSize: "0.75rem", color: "#666", textTransform: "uppercase" }}>{label}</div>
      <div style={{ fontSize: "1.3rem", fontWeight: 600 }}>{value}</div>
      {sub && <div style={{ fontSize: "0.8rem", color: "#666" }}>{sub}</div>}
    </div>
  )
}

function Bar({ pct }: { pct: number}) {
  const w = Math.max(0, Math.min(100, pct))
  return (
    <span style={{ display: "inline-block", width: 120, height: 8, background: "#eee", borderRadius: 4, verticalAlign: "middle" }}>
      <span style={{ display: "inline-block", width: `${w}%`, height: 8, borderRadius: 4, backgrund: "#111" }} />
    </span>
  )
}

export function FireTab({ data }: { data: FireSummary | null }) {
  if (!data) return <p>Loading...</p>

  const cur = data.fire_number.commodity
  return (
    <div>
      <div style={{ display: "flex", gap: "0.75rem", flexWrap: "wrap", marginBottom: "1rem" }}>
        <Card
          label="Net worth"
          value={`${data.net_worth.value} ${data.net_worth.commodity}`}
        />
        <Card
          label="Savings rate"
          value={`${(data.savings_rate * 100).toFixed(1)}%`}
          sub={`${data.annual_savings.value} ${data.annual_savings.commodity}/yr`}
        />
        <Card
          label="FIRE number"
          value={`${data.fire_number.value} ${cur}`}
          sub={`${data.annual_expenses.value} ${cur}/yr ÷ 4%`}
        />
        <Card
          label="Years to FIRE"
          value={data.years_to_fire < 0 ? "-" : data.years_to_fire.toFixed(1)}
          sub={data.years_t_fire < 0 ? "never at current pace" : undefined}
        />
      </div>

      {(data.coast_fire || data.lean_fire || data.fat_fire) && (
        <>
          <h3 style={{ marginBottom: "0.25rem" }}>Variants</h3>
          <table style={{ borderCollapse: "collapse", fontSize: "0.9rem", marginBottom: "1rem" }}>
            <tbody>
              {data.coast_fire && (
                <tr style={{ borderBottom: "1px solid #eee" }}>
                  <td style={{ padding: "0.3rem 1rem 0.3rem 0" }}>Coast FIRE</td>
                  <td style={{ textAlign: "right" }}>{data.coast_fire.value} {data.coast_fire.commodity}</td>
                  <td style={{ paddingLeft: "1rem" }}>
                    <Bar pct={(data.coast_progress ?? 0) * 100} />{" "}
                    {((data.coast_progress ?? 0) * 100).toFixed(0)}% funded
                  </td>
                </tr>
              )}
              {data.lean_fire && (
                <tr style={{ borderBottom: "1px solid #eee" }}>
                  <td style={{ padding: "0.3rem 1rem 0.3rem 0" }}>Lean FIRE</td>
                  <td style={{ textAlign: "right" }}>{data.lean_fire.value} {data.lean_fire.commodity}</td>
                </tr>
              )}
              {data.fat_fire && (
                <tr style={{ borderBottom: "1px solid #eee" }}>
                  <td style={{ padding: "0.3rem 1rem 0.3rem 0" }}>Fat FIRE</td>
                  <td style={{ textAlign: "right" }}>{data.fat_fire.value} {data.fat_fire.commodity}</td>
                </tr>
              )}
            </tbody>
          </table>
        </>
      )}

      {data.simulation && (
        <>
          <h3 style={{ marginBottom: "0.25rem"}}>
            Monte Carlo <span style={{
              fontWeight: 400, fontSize: "0.8rem", color: "#666"
            }}>
              ({data.simulation.runs} runs · {data.simulation.years} yrs)
            </span>
          </h3>
          <table style={{ borderCollapse: "collapse", fontSize: "0.9rem" }}>
            <tbody>
              {(["p10", "p50", "p90"] as const).map((k) => (
                <tr key={k} style={{ borderBottom: "1px solid #eee" }}>
                  <td style={{ padding: "0.3rem 1rem 0.3rem 0", textTransform: "uppercase" }}>{k}</td>
                  <td style={{ textAlign: "right" }}>
                    {Number(data.simulation![k].value).toFixed(2)} {data.simulation![k].commodity}
                  </td>
                </tr>
              ))}
              <tr>
                <td style={{ padding: "0.3rem 1rem 0.3rem 0" }}>P(FIRE)</td>
                <td style={{ textAlign: "right" }}>
                  <Bar pct={data.simulation.prob_fire * 100} /> {(data.simulation.prob_fire * 100).toFixed(1)}%
                </td>
              </tr>
            </tbody>
          </table>
        </>
      )}
      {!data.simulation && (
        <p style={{ color: "#666", fontSize: "0.85rem" }}>
          Sin simulación - la API la incluye con <code>?simulate=1</code>
        </p>
      )}
    </div>
  )
}
