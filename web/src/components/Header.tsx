import React from 'react';

interface HeaderProps {
  title: string;
}

const Header: React.FC<HeaderProps> = ({ title }) => {
  return (
    <div className="header">
      <h1 className="page-title">{title}</h1>
      <div className="user-info">
        <div className="avatar">A</div>
        <span>管理员</span>
      </div>
    </div>
  );
};

export default Header;
