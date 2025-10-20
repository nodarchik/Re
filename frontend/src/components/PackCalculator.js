import React, { useState } from 'react';
import { calculatePacks } from '../services/api';

/**
 * PackCalculator component - Main form for calculating pack combinations
 */
function PackCalculator({ onCalculate }) {
  const [amount, setAmount] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    const amountNum = parseInt(amount);
    if (isNaN(amountNum) || amountNum < 1) {
      setError('Please enter a valid amount (minimum 1)');
      return;
    }
    
    // Validate maximum amount
    const maxAmount = 10000000; // 10 million
    if (amountNum > maxAmount) {
      setError(`Amount too large. Maximum allowed: ${maxAmount.toLocaleString()} items`);
      return;
    }
    
    setLoading(true);
    setError('');
    setResult(null);
    
    try {
      const data = await calculatePacks(amountNum);
      setResult(data);
      if (onCalculate) {
        onCalculate(data);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="calculator-container">
      <h2>Calculate Pack Combination</h2>
      
      <form onSubmit={handleSubmit} className="calculator-form">
        <div className="form-group">
          <label htmlFor="amount">Number of Items to Order:</label>
          <input
            type="number"
            id="amount"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="Enter amount (max 10,000,000)"
            min="1"
            max="10000000"
            required
          />
        </div>
        
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? 'Calculating...' : 'Calculate'}
        </button>
      </form>
      
      {error && (
        <div className="error-message">
          {error}
        </div>
      )}
      
      {result && (
        <div className="result-container">
          <h3>Result</h3>
          <div className="result-summary">
            <div className="result-item">
              <span className="label">Requested:</span>
              <span className="value">{result.amount} items</span>
            </div>
            <div className="result-item">
              <span className="label">Total Items:</span>
              <span className="value">{result.total_items} items</span>
            </div>
            <div className="result-item">
              <span className="label">Total Packs:</span>
              <span className="value">{result.total_packs} packs</span>
            </div>
          </div>
          
          <h4>Pack Breakdown</h4>
          <table className="pack-table">
            <thead>
              <tr>
                <th>Pack Size</th>
                <th>Quantity</th>
                <th>Total Items</th>
              </tr>
            </thead>
            <tbody>
              {Object.entries(result.packs)
                .sort(([a], [b]) => parseInt(b) - parseInt(a))
                .map(([packSize, quantity]) => (
                  <tr key={packSize}>
                    <td>{packSize}</td>
                    <td>{quantity}</td>
                    <td>{parseInt(packSize) * quantity}</td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

export default PackCalculator;

