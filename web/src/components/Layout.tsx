import React from 'react';
import Sidebar from './Sidebar';
import Header from './Header';

interface LayoutProps {
  children: React.ReactNode;
  title: string;
}

const Layout: React.FC<LayoutProps> = ({ children, title }) => {
  return (
    <div className="app">
      <Sidebar />
      <div className="main-content">
        <Header title={title} />
        <div className="content">{children}</div>
      </div>
    </div>
  );
};

export default Layout;
