import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import DashboardPage from './pages/DashboardPage';
import TasksPage from './pages/TasksPage';
import AssetsPage from './pages/AssetsPage';
import CreativeGeneratorPage from './pages/CreativeGeneratorPage';
import ExperimentsPage from './pages/ExperimentsPage';

const App: React.FC = () => {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/tasks" element={<TasksPage />} />
        <Route path="/assets" element={<AssetsPage />} />
        <Route path="/creative" element={<CreativeGeneratorPage />} />
        <Route path="/experiments" element={<ExperimentsPage />} />
      </Routes>
    </Router>
  );
};

export default App;
