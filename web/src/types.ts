// Shared Póros API types.
// Single source of truth for the frontend until the OpenAPI-generated client
// replaces this file

export type Amount = { value: string; commodity: string }

export type BalanceRow = { account: string; commodity: string; amount: Amount }

export type Tx = {
  id:        string
  date:      string
  type:      string
  title?:    string
  amount?:   Amount
  account?:  string
  from?:     string
  to?:       string
  asset?:    string
  quantity?: string
  price?:    Amount
  category?: string
  tags?:     string[]
}

export type Position = {
  asset:        string
  account:      string
  quantity:     Amount
  price:        Amount
  cost_value:   Amount
  market_value?: Amount
}

export type Portfolio = {
  positions:     Position[]
  total_cost:    Amount
  total_market?: Amount
  basis?:        string
}

export type Basis = "cost" | "market"

export type Simulation = {
  runs:      number
  years:     number
  p10:       Amount
  p50:       Amount
  p90:       Amount
  prob_fire: number
}

export type FireSummary = {
  year:            number
  net_worth:       Amount
  annual_income:   Amount
  annual_expenses: Amount
  savings_rate:    number
  annual_savings:  Amount
  fire_number:     Amount
  years_to_fire:   number
  coast_fire?:     Amount
  coast_progress?: number
  lean_fire?:      Amount
  fat_fire?:       Amount
  simulation?:     Simulation
}
