import React from 'react';
import { NavLink } from 'react-router-dom';

const Sidebar: React.FC = () => {
  return (
    <div className="sidebar">
      <div className="sidebar-header">
        <h2>
          <i className="fas fa-bullseye"></i> <span>创意平台</span>
        </h2>
      </div>
      <nav className="nav-menu">
        <NavLink to="/" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-home"></i>
          <span>仪表盘</span>
        </NavLink>
        <NavLink to="/creative" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-magic"></i>
          <span>创意生成</span>
        </NavLink>
        <NavLink to="/assets" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-images"></i>
          <span>素材管理</span>
        </NavLink>
        <NavLink to="/tasks" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-tasks"></i>
          <span>任务管理</span>
        </NavLink>
        <NavLink to="/experiments" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-vial"></i>
          <span>实验</span>
        </NavLink>
      </nav>
    </div>
  );
};

export default Sidebar;
