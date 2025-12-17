import React from 'react';
import LanguageSwitch from './LanguageSwitch';
import { useI18n } from '../i18n';

interface HeaderProps {
  title: string;
  showLanguageSwitch?: boolean;
}

const Header: React.FC<HeaderProps> = ({ title, showLanguageSwitch = true }) => {
  const { t } = useI18n();

  return (
    <div className="header">
      <h1 className="page-title">{title}</h1>
      <div className="user-info">
        {showLanguageSwitch && <LanguageSwitch />}
        <div className="avatar">A</div>
        <span>{t('admin')}</span>
      </div>
    </div>
  );
};

export default Header;
