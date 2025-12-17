import React, { useEffect, useMemo, useState } from 'react';
import Sidebar from '../components/Sidebar';
import { warmupAPI } from '../services/api';
import type { WarmupStats } from '../types';
import { useI18n } from '../i18n';
import Header from '../components/Header';

const formatDateTime = (value?: string, locale: string = 'zh-CN') => {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString(locale, {
    hour12: false,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
};

const formatDuration = (ms?: number) => {
  if (ms === undefined || ms === null) return '-';
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
};

const WarmupPage: React.FC = () => {
  const { t, lang } = useI18n();
  const [stats, setStats] = useState<WarmupStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const locale = useMemo(() => (lang === 'zh' ? 'zh-CN' : 'en-US'), [lang]);

  const loadStatus = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await warmupAPI.status();
      if (res.code === 0 && res.data) {
        setStats(res.data);
      } else {
        setError(res.message || `API returned error code: ${res.code}`);
      }
    } catch (err) {
      const message = (err as Error)?.message || (err as any)?.toString() || 'Failed to load warmup status';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  const runNow = async () => {
    setRunning(true);
    setError(null);
    try {
      const res = await warmupAPI.run();
      if (res.code === 0 && res.data) {
        setStats(res.data);
      } else {
        setError(res.message || `API returned error code: ${res.code}`);
      }
    } catch (err) {
      const message = (err as Error)?.message || (err as any)?.toString() || 'Failed to run warmup';
      setError(message);
    } finally {
      setRunning(false);
    }
  };

  useEffect(() => {
    loadStatus();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const records = useMemo(() => (stats?.recent || []).slice(0, 10), [stats]);
  const successRate = stats && stats.runs > 0 ? ((stats.successes / stats.runs) * 100).toFixed(1) : '0';
  const limitHint = t('warmupRecentHint', '只显示最近 10 条记录');
  const lastUpdated = stats?.last_run ? formatDateTime(stats.last_run, locale) : t('notFilled');

  return (
    <div className="app">
      <Sidebar />
      <div className="main-content">
        <Header title={t('headerWarmup')} />

        <div className="content">
          <div className="dashboard-layout" style={{ maxWidth: 1200, gap: 16 }}>
            {error && (
              <div className="compact-alert compact-alert-error" style={{ marginBottom: 12 }}>
                <i className="fas fa-exclamation-circle" />
                <span>{error}</span>
              </div>
            )}

              <div className="compact-toolbar" style={{ alignItems: 'center' }}>
                <div className="compact-toolbar-left" style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
                  <span className="compact-stats-text">
                    {t('warmupRuns')}：{stats?.runs ?? '-'}
                  </span>
                  <span className="compact-stats-text">
                    {t('warmupSuccess')}：{stats?.successes ?? '-'}
                  </span>
                  <span className="compact-stats-text">
                    {t('warmupFail')}：{stats?.failures ?? '-'}
                  </span>
                  <span className="compact-stats-text" style={{ color: '#1677ff', fontWeight: 600 }}>
                    {t('warmupSuccessRate', 'Success Rate')}：{successRate}%
                  </span>
                  <span className="compact-card-hint" style={{ fontSize: 12 }}>
                    {t('refresh')}：{lastUpdated}
                  </span>
                </div>
                <div className="compact-toolbar-right" style={{ display: 'flex', gap: 8 }}>
                  <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={loadStatus} disabled={loading || running}>
                    <i className="fas fa-sync" />
                    <span>{loading ? t('loading') : t('refresh')}</span>
                  </button>
                  <button className="compact-btn compact-btn-primary compact-btn-sm" onClick={runNow} disabled={loading || running}>
                    <i className="fas fa-bolt" />
                    <span>{running ? t('running', 'Running...') : t('warmupRunNow')}</span>
                  </button>
              </div>
            </div>

            <div className="compact-card" style={{ border: '1px solid #f0f0f0', boxShadow: '0 6px 18px rgba(0,0,0,0.04)' }}>
              <div className="compact-card-body" style={{ padding: 12 }}>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16 }}>
                  <div className="compact-meta-block">
                    <div className="compact-card-hint" style={{ textTransform: 'uppercase' }}>
                      {t('warmupLastRun')}
                    </div>
                    <div style={{ fontSize: 14, fontWeight: 600, color: '#1f2937' }}>
                      {formatDateTime(stats?.last_run, locale)}
                    </div>
                  </div>
                  <div className="compact-meta-block">
                    <div className="compact-card-hint" style={{ textTransform: 'uppercase' }}>
                      {t('warmupLastSuccess')}
                    </div>
                    <div style={{ fontSize: 14, fontWeight: 600, color: '#52c41a' }}>
                      {formatDateTime(stats?.last_success, locale)}
                    </div>
                  </div>
                  <div className="compact-meta-block">
                    <div className="compact-card-hint" style={{ textTransform: 'uppercase' }}>
                      {t('warmupLastError')}
                    </div>
                    <div style={{ fontSize: 12, color: stats?.last_error ? '#ff4d4f' : '#8c8c8c' }}>
                      {stats?.last_error || '-'}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div className="card" style={{ overflow: 'hidden' }}>
              <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <h3 className="card-title" style={{ marginBottom: 4 }}>
                    {t('warmupRecent')}
                  </h3>
                  <div className="compact-card-hint" style={{ fontSize: 12 }}>{limitHint}</div>
                </div>
                <span style={{ fontSize: 12, color: '#8c8c8c' }}>{t('warmupRuns')}：{stats?.runs ?? '-'}</span>
              </div>
              <div className="card-body" style={{ padding: 0 }}>
                {loading && !stats ? (
                  <div style={{ padding: 40, textAlign: 'center', color: '#8c8c8c' }}>
                    <div className="loading" style={{ margin: '0 auto 12px' }} />
                    {t('loading')}
                  </div>
                ) : records.length === 0 ? (
                  <div style={{ padding: 40, textAlign: 'center', color: '#8c8c8c' }}>
                    <i className="fas fa-inbox" style={{ fontSize: 48, marginBottom: 12, opacity: 0.3 }} />
                    <div>{t('warmupEmpty')}</div>
                  </div>
                ) : (
                  <div className="compact-table-wrapper">
                    <table className="compact-table">
                      <thead>
                        <tr>
                          <th style={{ width: '200px' }}>{t('warmupStartTime', 'Start Time')}</th>
                          <th style={{ width: '100px' }}>{t('warmupDuration')}</th>
                          <th style={{ width: '100px' }}>{t('status')}</th>
                          <th>{t('warmupActions')}</th>
                          <th style={{ width: '220px' }}>{t('warmupErrors')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {records.map((rec, idx) => (
                          <tr key={`${rec.started_at}-${idx}`}>
                            <td style={{ fontSize: 12, fontFamily: 'monospace' }}>{formatDateTime(rec.started_at, locale)}</td>
                            <td>
                              <span
                                style={{
                                  fontSize: 12,
                                  fontWeight: 600,
                                  color: rec.duration < 100 ? '#52c41a' : rec.duration < 500 ? '#faad14' : '#ff4d4f',
                                }}
                              >
                                {formatDuration(rec.duration)}
                              </span>
                            </td>
                            <td>
                              <span className={`compact-status ${rec.success ? 'compact-status-completed' : 'compact-status-failed'}`}>
                                <i className={`fas ${rec.success ? 'fa-check' : 'fa-times'}`} />
                                {rec.success ? t('warmupSuccess') : t('warmupFail')}
                              </span>
                            </td>
                            <td>
                              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                                {(rec.actions_run || []).map((action, actionIdx) => (
                                  <span
                                    key={`${action}-${actionIdx}`}
                                    style={{
                                      fontSize: 10,
                                      padding: '2px 6px',
                                      borderRadius: 4,
                                      background: '#f0f5ff',
                                      color: '#1d39c4',
                                      fontFamily: 'monospace',
                                    }}
                                  >
                                    {action}
                                  </span>
                                ))}
                              </div>
                            </td>
                            <td>
                              {rec.errors && rec.errors.length > 0 ? (
                                <div style={{ fontSize: 11, color: '#ff4d4f', lineHeight: 1.4 }}>
                                  {rec.errors.map((errMsg, errIdx) => (
                                    <div key={`${errMsg}-${errIdx}`}>• {errMsg}</div>
                                  ))}
                                </div>
                              ) : (
                                <span style={{ color: '#8c8c8c', fontSize: 12 }}>-</span>
                              )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default WarmupPage;
