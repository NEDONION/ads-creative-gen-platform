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
        <NavLink to="/warmup" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-fire"></i>
          <span>{t('navWarmup')}</span>
        </NavLink>
        <NavLink to="/plugin-preview" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
          <i className="fas fa-flask"></i>
          <span>{t('previewPlugin')}</span>
        </NavLink>
      </nav>

      <div className="sidebar-footer">
        <div className="footer-section">
          <div className="footer-title">
            <i className="fas fa-link"></i>
            <span>{t('resourcesTitle')}</span>
          </div>
          <div className="footer-links">
            <a href="https://www.jchu.me" target="_blank" rel="noopener noreferrer" className="footer-link">
              <i className="fas fa-globe"></i>
              <span>{t('personalWebsite')}</span>
              <i className="fas fa-external-link-alt" style={{ fontSize: '10px', marginLeft: 'auto' }}></i>
            </a>
            <a href="https://github.com/NEDONION/experiment-widget-sdk" target="_blank" rel="noopener noreferrer" className="footer-link">
              <i className="fab fa-github"></i>
              <span>{t('githubRepo')}</span>
              <i className="fas fa-external-link-alt" style={{ fontSize: '10px', marginLeft: 'auto' }}></i>
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Sidebar;
