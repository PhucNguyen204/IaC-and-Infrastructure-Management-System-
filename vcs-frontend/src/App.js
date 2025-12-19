import React, { useState, useEffect } from 'react';
import { Toaster } from 'react-hot-toast';
import './App.css';
import Login from './components/Login';
import StackDashboard from './components/StackDashboard';
import StackDetailPage from './components/StackDetailPage';
import Deploy from './components/Deploy';

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const [currentView, setCurrentView] = useState('dashboard'); // 'dashboard' | 'detail' | 'deploy'
  const [selectedStackId, setSelectedStackId] = useState(null);

  useEffect(() => {
    // Check if user is already logged in
    const token = localStorage.getItem('access_token');
    if (token) {
      setIsAuthenticated(true);
    }
    setLoading(false);
    
  }, []);

  const handleLogin = (token) => {
    localStorage.setItem('access_token', token);
    setIsAuthenticated(true);
  };

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    setIsAuthenticated(false);
    setCurrentView('dashboard');
    setSelectedStackId(null);
  };

  const handleViewStack = (stackId) => {
    setSelectedStackId(stackId);
    setCurrentView('detail');
  };

  const handleBackToDashboard = () => {
    setCurrentView('dashboard');
    setSelectedStackId(null);
  };

  const handleNavigate = (view) => {
    setCurrentView(view);
    if (view !== 'detail') {
      setSelectedStackId(null);
    }
  };

  if (loading) {
    return <div className="loading">Loading...</div>;
  }

  if (!isAuthenticated) {
    return (
      <>
        <Login onLogin={handleLogin} />
        <Toaster position="top-right" />
      </>
    );
  }

  const renderView = () => {
    switch (currentView) {
      case 'deploy':
        return <Deploy onLogout={handleLogout} onNavigate={handleNavigate} />;
      case 'detail':
        return (
          <StackDetailPage 
            stackId={selectedStackId} 
            onBack={handleBackToDashboard}
            onLogout={handleLogout}
          />
        );
      case 'dashboard':
      default:
        return (
          <StackDashboard 
            onLogout={handleLogout} 
            onViewStack={handleViewStack}
            onNavigate={handleNavigate}
          />
        );
    }
  };

  return (
    <>
      <div className="app">
        {renderView()}
      </div>
      <Toaster position="top-right" />
    </>
  );
}

export default App;
