import React from 'react';
import './Header.css';

const Header = ({ userName, userRole }) => {
  return (
    <header className="app-header">
      <div className="header-left">
        <h1 className="app-title">Lael Hospital</h1>
      </div>

      <div className="header-right">
        <div className="user-info">
          <span className="user-name">{userName || 'Guest'}</span>
          <span className="user-role">{userRole || 'Staff'}</span>
        </div>
        <button className="logout-btn">Logout</button>
      </div>
    </header>
  );
};

export default Header;
