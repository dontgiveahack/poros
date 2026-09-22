import { useMemo } from "react"

export type Amount = { value: string; commodity: string }
export type BalanceRow = { account: string; commodity: string; amount: Amount }

type Node = {
  name: string
  path: string
  children: Map<string, Node>
  totals: Map<string, number>
  leaf: BalanceRow[]
}

function buildTree(rows: BalanceRow[]): Node {
  const root: Node = {
    name: "",
    path: "",
    children: new Map(),
    totals: new Map(),
    leaf: []
  }

  for (const r of rows) {
    const segs = r.account.split("/")
    let node = root

    // Walk/create the path, accumulating the amount at every level
    for (const s of segs) {
      let child = node.children.get(s)
      if (!child) {
        child = {
          name: s,
          path: node.path ? `${node.path}/${s}` : s,
          children: new Map(),
          totals: new Map(),
          leaf: [],
        }

        node.children.set(s, child)
      }

      child.totals.set(r.commodity, (child.totals.get(r.commodity) ?? 0) + parseFloat(r.amount.value))
      node = child
    }

    node.leaf.push(r)
  }

  return root
}

function fmtTotals(totals: Map<string, number>): string {
  return [...totals.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([c, v]) => `${trim(v)} ${c}`)
    .join(" · ")
}

// Trim float noise for display: 1945.6800000001 -> 1945.68
function trim(v: number): string {
  return String(Math.round(v * 100) / 100)
}

function TreeNode({ node, depth, onSelect }: {
  node: Node;
  depth: number;
  onSelect: (acct: string) => void;
}) {
  const kids = [...node.children.values()].sort((a, b) => a.name.localeCompare(b.name))
  return (
    <details open={depth < 1} style={{ marginLeft: depth === 0 ? 0 : "1.2rem" }}>
      <summary style={{ cursor: "pointer", padding: "0.25rem 0" }}>
        <button
          onClick={(e) => {
            e.preventDefault() // don't toggle <details> when navigating
            onSelect(node.path)
          }}
          title={`Filter transactions by ${node.path}`}
          style={{
            background: "none", border: "none", padding: 0,
            font: "inherit", fontWeight: depth === 0 ? 600 : 400,
            color: "#1a56db", cursor: "pointer",
          }}
        >
          {node.name}
        </button>
        <span style={{ color: "#666", marginLeft: "0.5rem" }}>
          {fmtTotals(node.totals)}
        </span>
      </summary>
      {kids.map((k) => (
        <TreeNode key={k.path} node={k} depth={depth + 1} onSelect={onSelect} />
      ))}
    </details>
  )
}

export function AccountTree({ rows, onSelect }: {
  rows: BalanceRow[];
  onSelect: (acct: string) => void;
}) {
  const root = useMemo(() => buildTree(rows), [rows])
  const kids = [...root.children.values()].sort((a, b) => a.name.localeCompare(b.name))
  if (kids.length === 0) return <p style={{ color: "#666" }}>No accounts.</p>

  return (
    <div>
      {kids.map((k) => (
        <TreeNode key={k.path} node={k} depth={0} onSelect={onSelect} />
      ))}
      <p style={{ color: "#666", fontSize: "0.85rem" }}>Click an account to filter its transactions.</p>
    </div>
  )
}
