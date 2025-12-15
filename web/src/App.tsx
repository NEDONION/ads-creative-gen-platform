import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import DashboardPage from './pages/DashboardPage';
import TasksPage from './pages/TasksPage';
import AssetsPage from './pages/AssetsPage';
import CreativeGeneratorPage from './pages/CreativeGeneratorPage';
import ExperimentsPage from './pages/ExperimentsPage';
import ExperimentCreatePage from './pages/ExperimentCreatePage';
import TracePage from './pages/TracePage';
import { I18nProvider } from './i18n';

const App: React.FC = () => {
  return (
    <I18nProvider>
      <Router>
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/tasks" element={<TasksPage />} />
          <Route path="/assets" element={<AssetsPage />} />
          <Route path="/creative" element={<CreativeGeneratorPage />} />
          <Route path="/experiments" element={<ExperimentsPage />} />
          <Route path="/experiments/new" element={<ExperimentCreatePage />} />
          <Route path="/traces" element={<TracePage />} />
        </Routes>
      </Router>
    </I18nProvider>
  );
};

export default App;
