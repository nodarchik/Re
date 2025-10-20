import React, { useState } from 'react';
import './App.css';
import PackCalculator from './components/PackCalculator';
import PackSizeManager from './components/PackSizeManager';
import OrderHistory from './components/OrderHistory';

function App() {
  const [activeTab, setActiveTab] = useState('calculator');

  return (
    <div className="App">
      <header className="app-header">
        <h1>Pack Size Calculator</h1>
        <p className="subtitle">Calculate optimal pack combinations for your orders</p>
      </header>

      <nav className="tab-navigation">
        <button
          className={`tab-button ${activeTab === 'calculator' ? 'active' : ''}`}
          onClick={() => setActiveTab('calculator')}
        >
          Calculator
        </button>
        <button
          className={`tab-button ${activeTab === 'packs' ? 'active' : ''}`}
          onClick={() => setActiveTab('packs')}
        >
          Pack Sizes
        </button>
        <button
          className={`tab-button ${activeTab === 'history' ? 'active' : ''}`}
          onClick={() => setActiveTab('history')}
        >
          Order History
        </button>
      </nav>

      <main className="app-main">
        {activeTab === 'calculator' && <PackCalculator />}
        {activeTab === 'packs' && <PackSizeManager />}
        {activeTab === 'history' && <OrderHistory />}
      </main>

      <footer className="app-footer">
        <p>Pack Calculator API - Dynamic programming solution for optimal pack combinations</p>
      </footer>
    </div>
  );
}

export default App;
