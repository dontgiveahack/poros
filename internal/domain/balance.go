// Package domain balance calculation
package domain

import (
	"fmt"
	"math/big"
)

// Balances maps account -> commodity -> amount.
// E.g. balances["bank/checking"]["EUR"] == -554.32 EUR
type Balances map[string]map[Commodity]Amount

// CalculateBalances applies every transaction to the account balances.
// Supports income, expense, transfer, buy, sell and dividend.
// Unknown types are ignored (future types can be added without breaking).
func CalculateBalances(txs []Transaction) (Balances, error) {
	balances := make(Balances)
	for _, tx := range txs {
		switch tx.Type {
		case TxExpense:
			if err := add(balances, tx.Account, tx.Amount.Neg()); err != nil {
				return nil, fmt.Errorf("tx %s: %w", tx.ID, err)
			}

		case TxIncome:
			if err := add(balances, tx.Account, tx.Amount); err != nil {
				return nil, fmt.Errorf("ts %s: %w", tx.ID, err)
			}

		case TxTransfer:
			if err := add(balances, tx.From, tx.Amount.Neg()); err != nil {
				return nil, fmt.Errorf("tx %s: %w", tx.ID, err)
			}

			if err := add(balances, tx.To, tx.Amount); err != nil {
				return nil, fmt.Errorf("tx %s: %w", tx.ID, err)
			}

		case TxBuy:
			qty, err := tx.QuantityRat()
			if err != nil {
				return nil, err
			}

			totalRat := new(big.Rat).Mul(qty, tx.Price.Rat())
			total := NewAmount(totalRat, tx.Price.Commodity())
			if err := add(balances, tx.Account, total.Neg()); err != nil {
				return nil, fmt.Errorf("tx %s: %w", tx.ID, err)
			}

			qtyAmt := NewAmount(qty, Commodity(tx.Asset))
			if err := add(balances, tx.Account, qtyAmt); err != nil {
				return nil, fmt.Errorf("tx %s: %w", tx.ID, err)
			}

		case TxSell:
			qty, err := tx.QuantityRat()
			if err != nil {
				return nil, err
			}

			totalRat := new(big.Rat).Mul(qty, tx.Price.Rat())
			total := NewAmount(totalRat, tx.Price.Commodity())
			if err := add(balances, tx.Account, total); err != nil {
				return nil, fmt.Errorf("tx %s: %w", tx.ID, err)
			}

			qtyAmt := NewAmount(qty, Commodity(tx.Asset))
			if err := add(balances, tx.Account, qtyAmt.Neg()); err != nil {
				return nil, fmt.Errorf("tx %s: %w", tx.ID, err)
			}

		case TxDividend, TxInterest:
			if err := add(balances, tx.Account, tx.Amount); err != nil {
				return nil, fmt.Errorf("tx %s: %w", tx.ID, err)
			}

			if tx.Tax != nil {
				if err := add(balances, tx.Account, tx.Tax.Neg()); err != nil {
					return nil, fmt.Errorf("tx %s tax: %w", tx.ID, err)
				}
			}

		case TxFee:
			if err := add(balances, tx.Account, tx.Amount.Neg()); err != nil {
				return nil, fmt.Errorf("tx %s: %w", tx.ID, err)
			}
		}
	}

	return balances, nil
}

func add(balances Balances, account string, amount Amount) error {
	if account == "" {
		return fmt.Errorf("empty account")
	}

	m, ok := balances[account]
	if !ok {
		m = make(map[Commodity]Amount)
		balances[account] = m
	}

	cur, exists := m[amount.Commodity()]
	if !exists {
		m[amount.Commodity()] = amount
		return nil
	}

	sum, err := cur.Add(amount)
	if err != nil {
		return err
	}

	m[amount.Commodity()] = sum
	return nil
}
