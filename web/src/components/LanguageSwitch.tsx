import React from 'react';
import { useI18n } from '../i18n';

const LanguageSwitch: React.FC = () => {
  const { lang, setLang } = useI18n();
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
      <span role="img" aria-label="language" style={{ fontSize: 14 }}>
        🌐
      </span>
      <select
        className="compact-input"
        style={{ width: 120, padding: '4px 8px', height: 30, fontSize: 12 }}
        value={lang}
        onChange={(e) => setLang(e.target.value as 'zh' | 'en')}
      >
        <option value="zh">中文</option>
        <option value="en">English</option>
      </select>
    </div>
  );
};

export default LanguageSwitch;
