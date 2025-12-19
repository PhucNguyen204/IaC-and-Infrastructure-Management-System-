import React from 'react';
import { 
  LayoutDashboard, 
  Package, 
  Rocket, 
  BarChart3, 
  Bell, 
  Settings,
  LogOut,
  User
} from 'lucide-react';
import './Layout.css';

const Layout = ({ children, onLogout, activeTab = 'stacks', onNavigate }) => {
  const handleNavClick = (e, tab) => {
    e.preventDefault();
    if (onNavigate) {
      onNavigate(tab);
    }
  };

  return (
    <div className="layout">
      <header className="layout-header">
        <div className="header-left">
          <Package className="header-logo" />
          <h1>VCS Infrastructure Platform</h1>
        </div>
        <div className="header-right">
          <button className="icon-btn">
            <Bell size={20} />
            <span className="badge">3</span>
          </button>
          <div className="user-menu">
            <User size={20} />
            <span>Admin</span>
          </div>
          <button className="icon-btn" onClick={onLogout}>
            <LogOut size={20} />
          </button>
        </div>
      </header>

      <div className="layout-body">
        <aside className="sidebar">
          <nav className="sidebar-nav">
            <a 
              href="#overview" 
              className={`nav-item ${activeTab === 'overview' ? 'active' : ''}`}
              onClick={(e) => handleNavClick(e, 'overview')}
            >
              <LayoutDashboard size={20} />
              <span>Overview</span>
            </a>
            <a 
              href="#stacks" 
              className={`nav-item ${activeTab === 'stacks' ? 'active' : ''}`}
              onClick={(e) => handleNavClick(e, 'dashboard')}
            >
              <Package size={20} />
              <span>Stacks</span>
            </a>
            <a 
              href="#deploy" 
              className={`nav-item ${activeTab === 'deploy' ? 'active' : ''}`}
              onClick={(e) => handleNavClick(e, 'deploy')}
            >
              <Rocket size={20} />
              <span>Deploy</span>
            </a>
            <a 
              href="#metrics" 
              className={`nav-item ${activeTab === 'metrics' ? 'active' : ''}`}
              onClick={(e) => handleNavClick(e, 'metrics')}
            >
              <BarChart3 size={20} />
              <span>Metrics</span>
            </a>
            <a 
              href="#alerts" 
              className={`nav-item ${activeTab === 'alerts' ? 'active' : ''}`}
              onClick={(e) => handleNavClick(e, 'alerts')}
            >
              <Bell size={20} />
              <span>Alerts</span>
            </a>
            <a 
              href="#settings" 
              className={`nav-item ${activeTab === 'settings' ? 'active' : ''}`}
              onClick={(e) => handleNavClick(e, 'settings')}
            >
              <Settings size={20} />
              <span>Settings</span>
            </a>
          </nav>
        </aside>

        <main className="main-content">
          {children}
        </main>
      </div>
    </div>
  );
};

export default Layout;
