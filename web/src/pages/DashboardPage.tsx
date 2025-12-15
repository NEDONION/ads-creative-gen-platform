import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { creativeAPI, experimentAPI } from '../services/api';
import Sidebar from '../components/Sidebar';
import LanguageSwitch from '../components/LanguageSwitch';
import { useI18n } from '../i18n';

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
  type: 'task' | 'asset' | 'user' | 'experiment';
  message: string;
  time: string;
}

const ActivityItem: React.FC<ActivityItemProps> = ({ type, message, time }) => {
  const iconMap: Record<ActivityItemProps['type'], string> = {
    task: '✓',
    asset: '🎨',
    user: '👤',
    experiment: '🧪',
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

type ActivityType = 'task' | 'asset' | 'experiment';

interface ActivityEntry {
  id: string;
  type: ActivityType;
  message: string;
  time: string;
  timestamp: number;
}

const formatRelativeTime = (dateString?: string) => {
  if (!dateString) return '-';
  const date = new Date(dateString);
  const ts = date.getTime();
  if (!Number.isFinite(ts)) return dateString;

  const diffMs = Date.now() - ts;
  const minutes = Math.floor(diffMs / 60000);
  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes} 分钟前`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} 小时前`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days} 天前`;

  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
};

const shortId = (id?: string) => {
  if (!id) return '-';
  if (id.length <= 8) return id;
  return `${id.substring(0, 8)}...`;
};

const parseTimestamp = (value?: string) => {
  if (!value) return 0;
  const ts = new Date(value).getTime();
  return Number.isFinite(ts) ? ts : 0;
};

const getTaskStatusText = (status?: string) => {
  const map: Record<string, string> = {
    pending: '待处理',
    queued: '排队中',
    processing: '处理中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
  };
  return map[status || ''] || status || '';
};

const getExperimentStatusText = (status?: string) => {
  const map: Record<string, string> = {
    active: '进行中',
    archived: '已归档',
    draft: '草稿',
    completed: '已完成',
  };
  return map[status || ''] || status || '已更新';
};

// 主 Dashboard 组件
const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useI18n();
  const [totalAssets, setTotalAssets] = useState(0);
  const [totalTasks, setTotalTasks] = useState(0);
  const [totalExperiments, setTotalExperiments] = useState(0);
  const [loading, setLoading] = useState(true);
  const [activities, setActivities] = useState<ActivityEntry[]>([]);
  const [activityLoading, setActivityLoading] = useState(true);
  const [activityError, setActivityError] = useState<string | null>(null);

  useEffect(() => {
    loadStats();
    loadActivities();
  }, []);

  const loadStats = async () => {
    setLoading(true);
    try {
      const [assetsRes, tasksRes, experimentsRes] = await Promise.all([
        creativeAPI.listAssets({ page: 1, page_size: 1 }),
        creativeAPI.listTasks({ page: 1, page_size: 1 }),
        experimentAPI.list({ page: 1, page_size: 1 }),
      ]);

      if (assetsRes.code === 0 && assetsRes.data) {
        setTotalAssets(assetsRes.data.total);
      }
      if (tasksRes.code === 0 && tasksRes.data) {
        setTotalTasks(tasksRes.data.total);
      }
      if (experimentsRes.code === 0 && experimentsRes.data) {
        setTotalExperiments(experimentsRes.data.total);
      }
    } catch (err) {
      console.error('Load stats error:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadActivities = async () => {
    setActivityLoading(true);
    setActivityError(null);
    try {
      const [tasksRes, assetsRes, experimentsRes] = await Promise.all([
        creativeAPI.listTasks({ page: 1, page_size: 20 }),
        creativeAPI.listAssets({ page: 1, page_size: 20 }),
        experimentAPI.list({ page: 1, page_size: 20 }),
      ]);

      const taskActivities: ActivityEntry[] = (tasksRes.data?.tasks || []).map((task) => {
        const timestamp = parseTimestamp(task.completed_at || task.created_at);
        return {
          id: task.id,
          type: 'task',
          message: `任务 ${shortId(task.id)} ${getTaskStatusText(task.status)}`,
          time: formatRelativeTime(task.completed_at || task.created_at),
          timestamp,
        };
      });

      const assetActivities: ActivityEntry[] = (assetsRes.data?.assets || []).map((asset) => {
        const timestamp = parseTimestamp(asset.created_at || asset.updated_at);
        const label = asset.title || asset.product_name || '创意';
        return {
          id: asset.id,
          type: 'asset',
          message: `${label} 素材已生成`,
          time: formatRelativeTime(asset.created_at || asset.updated_at),
          timestamp,
        };
      });

      const experimentActivities: ActivityEntry[] = (experimentsRes.data?.experiments || []).map((exp) => {
        const timestamp = parseTimestamp(exp.start_at || exp.created_at);
        const displayName = exp.name || shortId(exp.experiment_id);
        return {
          id: exp.experiment_id,
          type: 'experiment',
          message: `实验 ${displayName} ${getExperimentStatusText(exp.status)}`,
          time: formatRelativeTime(exp.start_at || exp.created_at),
          timestamp,
        };
      });

      const merged = [...taskActivities, ...assetActivities, ...experimentActivities]
        .filter((item) => item.timestamp > 0)
        .sort((a, b) => b.timestamp - a.timestamp)
        .slice(0, 5);

      setActivities(merged);
    } catch (err) {
      console.error('Load activities error:', err);
      setActivityError('加载最近活动失败');
    } finally {
      setActivityLoading(false);
    }
  };

  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">{t('headerDashboard')}</h1>
          <div className="user-info">
            <LanguageSwitch />
            <div className="avatar">A</div>
            <span>{t('admin')}</span>
          </div>
        </div>

        <div className="content">
          <div className="dashboard-layout">
            {/* Stats Grid */}
            <div className="stats-grid">
              <StatCard
                icon="📊"
                value={totalAssets}
                label={t('statAssets')}
                trend={{ value: '+12%', isPositive: true }}
                loading={loading}
              />
              <StatCard
                icon="📋"
                value={totalTasks}
                label={t('statTasks')}
                trend={{ value: '+8%', isPositive: true }}
                loading={loading}
              />
              <StatCard
                icon="🧪"
                value={totalExperiments}
                label={t('statExperiments')}
                loading={loading}
              />
            </div>

            {/* Info Cards Row */}
            <div className="info-grid">
              <InfoCard
                title={t('welcomeTitle')}
                icon="👋"
                action={{
                  label: t('welcomeAction'),
                  onClick: () => navigate('/creative'),
                }}
              >
                <p>{t('welcomeP1')}</p>
                <p>{t('welcomeP2')}</p>
                <p style={{ color: '#8c8c8c', fontSize: 12 }}>{t('welcomeHint')}</p>
              </InfoCard>

              <InfoCard title={t('aboutTitle')} icon="💡">
                <p>{t('aboutParagraph')}</p>
                <ul className="feature-list">
                  <li>✓ {t('feature1')}</li>
                  <li>✓ {t('feature2')}</li>
                  <li>✓ {t('feature3')}</li>
                  <li>✓ {t('feature4')}</li>
                </ul>
                <p style={{ color: '#8c8c8c', fontSize: 12 }}>{t('migrateTip')}</p>
              </InfoCard>
            </div>

            {/* Recent Activity */}
            <div className="activity-card">
              <div className="activity-header">
                <h3 className="activity-title">{t('recentActivity')}</h3>
                <a href="/tasks" className="activity-link">
                  {t('viewAll')} →
                </a>
              </div>
              <div className="activity-list">
                {activityLoading ? (
                  <div className="compact-loading">
                    <div className="loading"></div>
                    <div className="compact-loading-text">{t('loading')}</div>
                  </div>
                ) : activityError ? (
                  <div className="compact-alert compact-alert-error">
                    <i className="fas fa-exclamation-circle"></i>
                    <span>{t('activityError')}</span>
                    <button className="compact-btn compact-btn-text compact-btn-xs" style={{ marginLeft: 8 }} onClick={loadActivities}>
                      {t('retry')}
                    </button>
                  </div>
                ) : activities.length === 0 ? (
                  <div className="compact-empty">
                    <i className="fas fa-stream"></i>
                    <div className="compact-empty-text">{t('noActivity')}</div>
                    <div className="compact-empty-hint">{t('noActivityHint')}</div>
                  </div>
                ) : (
                  <>
                    {activities.map((activity) => (
                      <ActivityItem
                        key={`${activity.type}-${activity.id}-${activity.timestamp}`}
                        type={activity.type}
                        message={activity.message}
                        time={activity.time}
                      />
                    ))}
                    <div style={{ fontSize: 12, color: '#8c8c8c', padding: '6px 2px' }}>{t('activityMore')}</div>
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;
