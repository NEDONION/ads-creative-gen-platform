import React from 'react';
import { NavLink } from 'react-router-dom';
import { useI18n } from '../i18n';

const Sidebar: React.FC = () => {
  const { t } = useI18n();
  return (
    <div className="sidebar">
      <div className="sidebar-header">
        <h2>
          <i className="fas fa-bullseye"></i> <span>{t('appTitle')}</span>
        </h2>
      </div>
      <nav className="nav-menu">
        <NavLink to="/" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-home"></i>
          <span>{t('navDashboard')}</span>
        </NavLink>
        <NavLink to="/creative" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-magic"></i>
          <span>{t('navCreative')}</span>
        </NavLink>
        <NavLink to="/assets" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-images"></i>
          <span>{t('navAssets')}</span>
        </NavLink>
        <NavLink to="/tasks" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-tasks"></i>
          <span>{t('navTasks')}</span>
        </NavLink>
        <NavLink to="/experiments" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-vial"></i>
          <span>{t('navExperiments')}</span>
        </NavLink>
        <NavLink to="/experiments/new" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-plus-circle"></i>
          <span>{t('navExperimentNew')}</span>
        </NavLink>
        <NavLink to="/traces" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-project-diagram"></i>
          <span>{t('navTraces')}</span>
        </NavLink>
      </nav>
    </div>
  );
};

export default Sidebar;
