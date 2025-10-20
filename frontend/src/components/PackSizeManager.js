import React, { useState, useEffect } from 'react';
import { getPackSizes, addPackSize, deletePackSize } from '../services/api';

/**
 * PackSizeManager component - Manage pack size configurations
 */
function PackSizeManager() {
  const [packSizes, setPackSizes] = useState([]);
  const [newSize, setNewSize] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Load pack sizes on component mount
  useEffect(() => {
    loadPackSizes();
  }, []);

  const loadPackSizes = async () => {
    try {
      const data = await getPackSizes();
      setPackSizes(data || []);
      setError('');
    } catch (err) {
      setError('Failed to load pack sizes');
    }
  };

  const handleAddPackSize = async (e) => {
    e.preventDefault();
    
    const size = parseInt(newSize);
    if (isNaN(size) || size < 1) {
      setError('Please enter a valid pack size (minimum 1)');
      return;
    }
    
    setLoading(true);
    setError('');
    setSuccess('');
    
    try {
      await addPackSize(size);
      setSuccess(`Pack size ${size} added successfully`);
      setNewSize('');
      await loadPackSizes();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDeletePackSize = async (size) => {
    if (!window.confirm(`Are you sure you want to delete pack size ${size}?`)) {
      return;
    }
    
    setError('');
    setSuccess('');
    
    try {
      await deletePackSize(size);
      setSuccess(`Pack size ${size} deleted successfully`);
      await loadPackSizes();
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="pack-size-manager">
      <h2>Pack Size Configuration</h2>
      
      <form onSubmit={handleAddPackSize} className="add-pack-form">
        <div className="form-group-inline">
          <input
            type="number"
            value={newSize}
            onChange={(e) => setNewSize(e.target.value)}
            placeholder="Enter pack size"
            min="1"
            required
          />
          <button type="submit" disabled={loading} className="btn-secondary">
            {loading ? 'Adding...' : 'Add Pack Size'}
          </button>
        </div>
      </form>
      
      {error && <div className="error-message">{error}</div>}
      {success && <div className="success-message">{success}</div>}
      
      <div className="pack-sizes-list">
        <h3>Available Pack Sizes</h3>
        {packSizes.length === 0 ? (
          <p className="no-data">No pack sizes configured</p>
        ) : (
          <div className="pack-sizes-grid">
            {packSizes
              .sort((a, b) => a.size - b.size)
              .map((pack) => (
                <div key={pack.id} className="pack-size-item">
                  <span className="pack-size-value">{pack.size} items</span>
                  <button
                    onClick={() => handleDeletePackSize(pack.size)}
                    className="btn-delete"
                    title="Delete pack size"
                  >
                    ×
                  </button>
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default PackSizeManager;

