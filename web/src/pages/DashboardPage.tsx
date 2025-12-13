import React, { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import { creativeAPI } from '../services/api';

const DashboardPage: React.FC = () => {
  const [totalAssets, setTotalAssets] = useState(0);
  const [totalTasks, setTotalTasks] = useState(0);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      const assetsResponse = await creativeAPI.listAssets({ page: 1, page_size: 1 });
      if (assetsResponse.code === 0 && assetsResponse.data) {
        setTotalAssets(assetsResponse.data.total);
      }

      const tasksResponse = await creativeAPI.listTasks({ page: 1, page_size: 1 });
      if (tasksResponse.code === 0 && tasksResponse.data) {
        setTotalTasks(tasksResponse.data.total);
      }
    } catch (err) {
      console.error('Load stats error:', err);
    }
  };

  return (
    <Layout title="仪表盘">
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-title">总素材数</div>
          <div className="stat-value">{totalAssets}</div>
        </div>
        <div className="stat-card">
          <div className="stat-title">总任务数</div>
          <div className="stat-value">{totalTasks}</div>
        </div>
        <div className="stat-card">
          <div className="stat-title">平均CTR</div>
          <div className="stat-value">-</div>
        </div>
        <div className="stat-card">
          <div className="stat-title">活跃用户</div>
          <div className="stat-value">-</div>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <h3 className="card-title">欢迎使用广告创意生成平台</h3>
        </div>
        <div className="card-body">
          <p>这是一个基于AI的广告创意生成平台,可以帮助您快速生成高质量的广告素材。</p>
        </div>
      </div>
    </Layout>
  );
};

export default DashboardPage;
