import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { creativeAPI } from '../services/api';
import Sidebar from '../components/Sidebar';

// 统计卡片组件
interface StatCardProps {
  icon: string;
  value: string | number;
  label: string;
  trend?: {
    value: string;
    isPositive: boolean;
  };
  loading?: boolean;
}

const StatCard: React.FC<StatCardProps> = ({ icon, value, label, trend, loading }) => {
  return (
    <div className="stat-card group">
      <div className="stat-icon">{icon}</div>
      <div className="stat-content">
        <div className="stat-value">{loading ? '...' : value}</div>
        <div className="stat-label">{label}</div>
      </div>
      {trend && !loading && (
        <div className={`stat-trend ${trend.isPositive ? 'positive' : 'negative'}`}>
          {trend.isPositive ? '↑' : '↓'} {trend.value}
        </div>
      )}
    </div>
  );
};

// 信息卡片组件
interface InfoCardProps {
  title: string;
  icon?: string;
  children: React.ReactNode;
  action?: {
    label: string;
    onClick: () => void;
  };
}

const InfoCard: React.FC<InfoCardProps> = ({ title, icon, children, action }) => {
  return (
    <div className="info-card">
      <div className="info-card-header">
        {icon && <span className="info-icon">{icon}</span>}
        <h3 className="info-title">{title}</h3>
      </div>
      <div className="info-card-body">{children}</div>
      {action && (
        <button className="info-action" onClick={action.onClick}>
          {action.label} →
        </button>
      )}
    </div>
  );
};

// 活动项组件
interface ActivityItemProps {
  type: 'task' | 'asset' | 'user';
  message: string;
  time: string;
}

const ActivityItem: React.FC<ActivityItemProps> = ({ type, message, time }) => {
  const iconMap = {
    task: '✓',
    asset: '🎨',
    user: '👤',
  };

  return (
    <div className="activity-item">
      <div className="activity-icon">{iconMap[type]}</div>
      <div className="activity-content">
        <span className="activity-message">{message}</span>
        <span className="activity-time">{time}</span>
      </div>
    </div>
  );
};

// 主 Dashboard 组件
const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const [totalAssets, setTotalAssets] = useState(0);
  const [totalTasks, setTotalTasks] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    setLoading(true);
    try {
      const [assetsRes, tasksRes] = await Promise.all([
        creativeAPI.listAssets({ page: 1, page_size: 1 }),
        creativeAPI.listTasks({ page: 1, page_size: 1 }),
      ]);

      if (assetsRes.code === 0 && assetsRes.data) {
        setTotalAssets(assetsRes.data.total);
      }
      if (tasksRes.code === 0 && tasksRes.data) {
        setTotalTasks(tasksRes.data.total);
      }
    } catch (err) {
      console.error('Load stats error:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">仪表盘</h1>
          <div className="user-info">
            <div className="avatar">A</div>
            <span>管理员</span>
          </div>
        </div>

        <div className="content">
          <div className="dashboard-layout">
            {/* Stats Grid */}
            <div className="stats-grid">
              <StatCard
                icon="📊"
                value={totalAssets}
                label="总素材数"
                trend={{ value: '+12%', isPositive: true }}
                loading={loading}
              />
              <StatCard
                icon="📋"
                value={totalTasks}
                label="总任务数"
                trend={{ value: '+8%', isPositive: true }}
                loading={loading}
              />
              <StatCard icon="📈" value="2.4%" label="平均CTR" loading={loading} />
              <StatCard icon="👤" value="89" label="活跃用户" loading={loading} />
            </div>

            {/* Info Cards Row */}
            <div className="info-grid">
              <InfoCard
                title="欢迎使用"
                icon="👋"
                action={{
                  label: '立即开始',
                  onClick: () => navigate('/creative'),
                }}
              >
                <p>开始创建您的第一个广告创意，体验 AI 驱动的高效创作流程。</p>
              </InfoCard>

              <InfoCard title="关于平台" icon="💡">
                <p>AI 驱动的广告创意生成平台</p>
                <ul className="feature-list">
                  <li>✓ 多尺寸智能生成</li>
                  <li>✓ CTR 预测与排序</li>
                  <li>✓ 云端存储管理</li>
                </ul>
              </InfoCard>
            </div>

            {/* Recent Activity */}
            <div className="activity-card">
              <div className="activity-header">
                <h3 className="activity-title">最近活动</h3>
                <a href="/tasks" className="activity-link">
                  查看全部 →
                </a>
              </div>
              <div className="activity-list">
                <ActivityItem type="task" message="任务 #c1b5... 已完成" time="2分钟前" />
                <ActivityItem type="asset" message="生成了 3 个素材" time="5分钟前" />
                <ActivityItem type="task" message="任务 #d65a... 已完成" time="10分钟前" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;
