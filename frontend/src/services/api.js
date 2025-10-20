// API service for communicating with the backend

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

/**
 * Calculate pack combination for a given amount
 */
export const calculatePacks = async (amount) => {
  const response = await fetch(`${API_BASE_URL}/api/calculate`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ amount }),
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to calculate packs');
  }
  
  return response.json();
};

/**
 * Get all pack sizes
 */
export const getPackSizes = async () => {
  const response = await fetch(`${API_BASE_URL}/api/packs`);
  
  if (!response.ok) {
    throw new Error('Failed to fetch pack sizes');
  }
  
  return response.json();
};

/**
 * Add a new pack size
 */
export const addPackSize = async (size) => {
  const response = await fetch(`${API_BASE_URL}/api/packs`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ size }),
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to add pack size');
  }
  
  return response.json();
};

/**
 * Delete a pack size
 */
export const deletePackSize = async (size) => {
  const response = await fetch(`${API_BASE_URL}/api/packs/${size}`, {
    method: 'DELETE',
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to delete pack size');
  }
  
  return response.json();
};

/**
 * Get order history
 */
export const getOrders = async (limit = 50) => {
  const response = await fetch(`${API_BASE_URL}/api/orders?limit=${limit}`);
  
  if (!response.ok) {
    throw new Error('Failed to fetch orders');
  }
  
  return response.json();
};

