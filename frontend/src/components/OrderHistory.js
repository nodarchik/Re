import React, { useState, useEffect } from 'react';
import { getOrders } from '../services/api';

/**
 * OrderHistory component - Display past order calculations
 */
function OrderHistory() {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadOrders();
  }, []);

  const loadOrders = async () => {
    setLoading(true);
    setError('');
    
    try {
      const data = await getOrders(50);
      setOrders(data || []);
    } catch (err) {
      setError('Failed to load order history');
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  return (
    <div className="order-history">
      <h2>Order History</h2>
      
      <button onClick={loadOrders} className="btn-secondary" disabled={loading}>
        {loading ? 'Loading...' : 'Refresh'}
      </button>
      
      {error && <div className="error-message">{error}</div>}
      
      {loading ? (
        <p>Loading orders...</p>
      ) : orders.length === 0 ? (
        <p className="no-data">No orders yet</p>
      ) : (
        <div className="orders-list">
          {orders.map((order) => (
            <div key={order.id} className="order-card">
              <div className="order-header">
                <span className="order-id">Order #{order.id}</span>
                <span className="order-date">{formatDate(order.created_at)}</span>
              </div>
              
              <div className="order-summary">
                <div className="order-stat">
                  <span className="stat-label">Requested:</span>
                  <span className="stat-value">{order.amount}</span>
                </div>
                <div className="order-stat">
                  <span className="stat-label">Total Items:</span>
                  <span className="stat-value">{order.total_items}</span>
                </div>
                <div className="order-stat">
                  <span className="stat-label">Total Packs:</span>
                  <span className="stat-value">{order.total_packs}</span>
                </div>
              </div>
              
              <div className="order-packs">
                <strong>Packs:</strong>
                {Object.entries(order.packs)
                  .sort(([a], [b]) => parseInt(b) - parseInt(a))
                  .map(([packSize, quantity]) => (
                    <span key={packSize} className="pack-badge">
                      {quantity} × {packSize}
                    </span>
                  ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default OrderHistory;

