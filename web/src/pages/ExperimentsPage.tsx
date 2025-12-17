import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { experimentAPI, creativeAPI } from '../services/api';
import Sidebar from '../components/Sidebar';
import type { ExperimentMetrics, Experiment } from '../types';
import Header from '../components/Header';
import { useI18n } from '../i18n';

const ExperimentsPage: React.FC = () => {
  const { t, lang } = useI18n();
  const [message, setMessage] = useState<string | null>(null);
  const [metrics, setMetrics] = useState<ExperimentMetrics | null>(null);
  const [creativeOptions, setCreativeOptions] = useState<{ id: string; label: string; thumb?: string; product_name?: string; cta_text?: string; selling_points?: string[]; title?: string }[]>([]);
  const [experimentList, setExperimentList] = useState<Experiment[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [selectedExp, setSelectedExp] = useState<Experiment | null>(null);
  const [metricsLoading, setMetricsLoading] = useState(false);

  useEffect(() => {
    loadOptions();
    loadExperiments();
  }, []);

  // 当加载列表后，自动选第一个实验（仅初始）
  useEffect(() => {
    if (!selectedExp && experimentList.length > 0) {
      handleSelectExperiment(experimentList[0].experiment_id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [experimentList]);

  const loadOptions = async () => {
    try {
      const [tasksRes, assetsRes] = await Promise.all([
        creativeAPI.listTasks({ page: 1, page_size: 100 }),
        creativeAPI.listAssets({ page: 1, page_size: 100 }),
      ]);

      if (assetsRes.code === 0 && assetsRes.data) {
        const taskProductMap: Record<string, string | undefined> = {};
        (tasksRes.data?.tasks || []).forEach((t) => {
          if (t.id && t.product_name) {
            taskProductMap[t.id] = t.product_name;
          }
        });

        const options = (assetsRes.data.assets || [])
          .map((asset) => {
            const resolvedProduct = asset.product_name || taskProductMap[String(asset.task_id)];
            const label = asset.title || resolvedProduct || t('creativeId');
            return {
              id: asset.id,
              label: `${label} (${asset.id})`,
              thumb: asset.image_url || asset.public_url,
              product_name: resolvedProduct,
              cta_text: asset.cta_text,
              selling_points: asset.selling_points,
              title: asset.title,
            };
          })
          .filter((v) => Boolean(v && v.id));
        setCreativeOptions(options);
      }
    } catch (err) {
      console.error(err);
      setMessage(t('activityError'));
    }
  };

  const loadExperiments = async () => {
    setListLoading(true);
    try {
      const res = await experimentAPI.list({ page: 1, page_size: 50 });
      if (res.code === 0 && res.data) {
        setExperimentList(res.data.experiments || []);
      }
    } catch (err) {
      console.error(err);
      setMessage(t('activityError'));
    } finally {
      setListLoading(false);
    }
  };

  const handleActivate = async (id?: string) => {
    const targetId = id || selectedExp?.experiment_id;
    if (!targetId) return;
    // 已停止的实验不允许重新激活
    const targetExp = experimentList.find((e) => e.experiment_id === targetId);
    if (targetExp?.status === 'archived') {
      setMessage(t('stopped'));
      return;
    }
    try {
      const res = await experimentAPI.updateStatus(targetId, 'active');
      if (res.code === 0) {
        setMessage(t('activate'));
        const now = new Date().toISOString();
        setSelectedExp((prev) => (prev && prev.experiment_id === targetId ? { ...prev, status: 'active', start_at: prev.start_at || now } : prev));
        setExperimentList((prev) =>
          prev.map((exp) =>
            exp.experiment_id === targetId ? { ...exp, status: 'active', start_at: exp.start_at || now } : exp
          )
        );
        loadExperiments();
      } else {
        setMessage(res.message || t('activityError'));
      }
    } catch (err) {
      setMessage(t('activityError') + ': ' + (err as Error).message);
    }
  };

  const handleStop = async (id?: string) => {
    const targetId = id || selectedExp?.experiment_id;
    if (!targetId) return;
    try {
      const res = await experimentAPI.updateStatus(targetId, 'archived');
      if (res.code === 0) {
        setMessage(t('stopped'));
        const now = new Date().toISOString();
        setSelectedExp((prev) => (prev && prev.experiment_id === targetId ? { ...prev, status: 'archived', end_at: now } : prev));
        setExperimentList((prev) =>
          prev.map((exp) =>
            exp.experiment_id === targetId ? { ...exp, status: 'archived', end_at: now } : exp
          )
        );
        loadExperiments();
      } else {
        setMessage(res.message || t('activityError'));
      }
    } catch (err) {
      setMessage(t('activityError') + ': ' + (err as Error).message);
    }
  };

  const loadMetrics = async (id?: string) => {
    setMetricsLoading(true);
    const targetId = id || selectedExp?.experiment_id;
    if (!targetId) {
      setMessage(t('noExperiment'));
      setMetricsLoading(false);
      return;
    }
    try {
      const res = await experimentAPI.metrics(targetId);
      if (res.code === 0 && res.data) {
        setMetrics(res.data);
      } else {
        setMessage(res.message || t('activityError'));
      }
    } catch (err) {
      setMessage(t('activityError') + ': ' + (err as Error).message);
    } finally {
      setMetricsLoading(false);
    }
  };

  const getCreativeInfo = (id: number | string) => creativeOptions.find((opt) => opt.id === String(id));

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setMessage(t('copySuccess'));
    } catch (err) {
      console.error('copy failed', err);
      setMessage(t('copyFail'));
    }
  };

  const formatTime = (s?: string) => {
    if (!s) return '-';
    const d = new Date(s);
    if (Number.isNaN(d.getTime())) return '-';
    return d.toLocaleString();
  };

  const formatDuration = (start?: string, end?: string) => {
    const startDate = start ? new Date(start) : null;
    const endDate = end ? new Date(end) : new Date();
    if (!startDate || Number.isNaN(startDate.getTime()) || !endDate || Number.isNaN(endDate.getTime())) return '-';
    const ms = endDate.getTime() - startDate.getTime();
    const hours = Math.floor(ms / (1000 * 60 * 60));
    const minutes = Math.floor((ms / (1000 * 60)) % 60);
    if (hours >= 24) {
      const days = Math.floor(hours / 24);
      return lang === 'en' ? `${days}d ${hours % 24}h` : `${days}天${hours % 24}小时`;
    }
    if (hours > 0) return lang === 'en' ? `${hours}h ${minutes}m` : `${hours}小时${minutes}分`;
    return lang === 'en' ? `${minutes}m` : `${minutes}分`;
  };

  const handleSelectExperiment = (id: string) => {
    // 切换展开/收起
    if (selectedExp && selectedExp.experiment_id === id) {
      setSelectedExp(null);
      setMetrics(null);
      return;
    }
    setMetrics(null);
    const hit = experimentList.find((e) => e.experiment_id === id) || null;
    setSelectedExp(hit);
    setMessage(`${t('view')} ${id}`);
    loadMetrics(id);
  };

  const metricsSummary = (() => {
    if (!metrics || !metrics.variants || metrics.variants.length === 0) return null;
    const impressions = metrics.variants.reduce((sum, v) => sum + v.impressions, 0);
    const clicks = metrics.variants.reduce((sum, v) => sum + v.clicks, 0);
    const avgCtr = impressions > 0 ? clicks / impressions : 0;
    const best = [...metrics.variants].sort((a, b) => b.ctr - a.ctr)[0];
    return { impressions, clicks, avgCtr, best };
  })();

  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <Header title={t('headerExperiments')} />

        <div className="content">
          <div className="compact-layout">
            {message && (
              <div className="compact-alert compact-alert-info">
                <i className="fas fa-info-circle"></i>
                <span>{message}</span>
              </div>
            )}

            <div className="compact-card gradient-card section-card section-list">
              <div className="compact-card-header">
                <div>
                  <h3 className="compact-card-title">{t('experimentList')}</h3>
                  <div className="compact-card-hint">{t('experimentListHint')}</div>
                </div>
                <div style={{ display: 'flex', gap: 8 }}>
                  <button className="compact-btn compact-btn-text compact-btn-sm" onClick={loadExperiments} disabled={listLoading}>
                    <i className="fas fa-sync-alt"></i>
                    <span>{listLoading ? t('loadingShort') : t('refresh')}</span>
                  </button>
                  <Link className="compact-btn compact-btn-primary compact-btn-sm" to="/experiments/new">
                    {t('createExperiment')}
                  </Link>
                </div>
              </div>
              <div className="compact-card-body">
                {listLoading ? (
                  <div className="compact-loading">
                    <div className="loading"></div>
                  </div>
                ) : experimentList.length === 0 ? (
                  <div style={{ color: '#666', fontSize: 13 }}>{t('noExperiment')}</div>
                ) : (
                  <div className="compact-table-wrapper">
                    <table className="compact-table">
                      <thead>
                        <tr>
                          <th>{t('name')}</th>
                          <th>{t('product')}</th>
                          <th>{t('statusLabel')}</th>
                          <th>{t('createdAtShort')}</th>
                          <th>{t('experimentId')}</th>
                          <th>{t('actionsShort')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {experimentList.map((exp) => {
                          const isSelected = selectedExp?.experiment_id === exp.experiment_id;
                          return (
                          <tr key={exp.experiment_id} className={isSelected ? 'table-row-selected' : ''}>
                            <td>{exp.name}</td>
                            <td>{exp.product_name || '-'}</td>
                            <td>
                              <span className={`status-badge status-${exp.status}`}>
                                {exp.status === 'active' && <i className="fas fa-play-circle" style={{ color: '#52c41a' }}></i>}
                                {exp.status === 'archived' && <i className="fas fa-stop-circle" style={{ color: '#ff7875' }}></i>}
                                {exp.status}
                              </span>
                            </td>
                            <td>{exp.created_at ? new Date(exp.created_at).toLocaleString() : '-'}</td>
                            <td style={{ fontFamily: 'monospace', fontSize: 12 }}>{exp.experiment_id}</td>
                            <td>
                              <button
                                className={`compact-btn compact-btn-xs ${selectedExp?.experiment_id === exp.experiment_id ? 'compact-btn-outline' : 'compact-btn-primary'}`}
                                onClick={() => handleSelectExperiment(exp.experiment_id)}
                              >
                                {selectedExp?.experiment_id === exp.experiment_id ? t('collapse') : t('view')}
                              </button>
                            </td>
                          </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
              {selectedExp && (
                <div className="compact-card section-card section-detail" style={{ marginTop: 16, borderColor: '#e6f0ff', boxShadow: '0 8px 18px rgba(0,0,0,0.05)' }}>
                  <div className="compact-card-header" style={{ alignItems: 'flex-start' }}>
                    <div>
                      <h3 className="compact-card-title">{t('detailTitle')}</h3>
                      <div className="compact-card-hint">{t('detailHint')}</div>
                    </div>
                    <div className="compact-card-actions" style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                      <button
                        className="compact-btn compact-btn-primary compact-btn-xs"
                        onClick={() => handleActivate(selectedExp.experiment_id)}
                        disabled={selectedExp.status === 'active' || selectedExp.status === 'archived'}
                      >
                        {selectedExp.status === 'archived' ? t('stopped') : selectedExp.status === 'active' ? t('activated') : t('activate')}
                      </button>
                      <button
                        className="compact-btn compact-btn-danger compact-btn-xs"
                        style={{ marginLeft: 'auto' }}
                        onClick={() => handleStop(selectedExp.experiment_id)}
                        disabled={selectedExp.status === 'archived'}
                      >
                        {selectedExp.status === 'archived' ? t('stopped') : t('stop')}
                      </button>
                      <button className="compact-btn compact-btn-outline compact-btn-xs" onClick={() => loadMetrics(selectedExp.experiment_id)} disabled={metricsLoading}>
                        {metricsLoading ? t('refreshing') : t('refreshMetrics')}
                      </button>
                      <button className="compact-btn compact-btn-text compact-btn-xs" onClick={() => { setSelectedExp(null); setMetrics(null); }}>
                        {t('collapse')}
                      </button>
                    </div>
                  </div>
                  <div className="compact-card-body">
                    <div className="detail-meta-grid">
                      <div className="meta-block wide">
                        <div className="meta-label">{t('experimentId')}</div>
                          <div className="meta-value code">
                          <code>{selectedExp.experiment_id}</code>
                          <button className="compact-btn compact-btn-outline compact-btn-xs" onClick={() => copyToClipboard(selectedExp.experiment_id)}>
                            {t('copyId')}
                          </button>
                        </div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">{t('statusLabel')}</div>
                        <div className="meta-value"><span className={`status-badge status-${selectedExp.status}`}>{selectedExp.status}</span></div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">{t('duration')}</div>
                        <div className="meta-value">{formatDuration(selectedExp.start_at || selectedExp.created_at, selectedExp.end_at)}</div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">{t('createdAtShort')}</div>
                        <div className="meta-value">{formatTime(selectedExp.created_at)}</div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">{t('startAt')}</div>
                        <div className="meta-value">{formatTime(selectedExp.start_at)}</div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">{t('endAt')}</div>
                        <div className="meta-value">{formatTime(selectedExp.end_at)}</div>
                      </div>
                    </div>

                    <div className="compact-section-title" style={{ marginTop: 12 }}>{t('variantDef')}</div>
                    <div className="compact-form-grid">
                      {(selectedExp.variants || []).map((v, idx) => {
                        const info = getCreativeInfo(v.creative_id) || {
                          thumb: v.image_url,
                          title: v.title,
                          product_name: v.product_name,
                          cta_text: v.cta_text,
                          selling_points: v.selling_points,
                        };
                        const appliedCTA = v.cta_text || info?.cta_text;
                        const appliedSP = v.selling_points && v.selling_points.length > 0 ? v.selling_points : info?.selling_points;
                        return (
                          <div key={idx} className="compact-card" style={{ padding: 10, border: '1px solid #f0f0f0' }}>
                            <div style={{ display: 'flex', gap: 10 }}>
                              {info?.thumb ? (
                                <img src={info.thumb} alt={t('creativeId')} style={{ width: 80, height: 80, objectFit: 'cover', borderRadius: 6, border: '1px solid #eee' }} />
                              ) : (
                                <div style={{ width: 80, height: 80, borderRadius: 6, border: '1px dashed #ccc', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999', fontSize: 12 }}>
                                  -
                                </div>
                              )}
                              <div style={{ flex: 1 }}>
                            <div style={{ fontWeight: 700 }}>{info?.title || info?.product_name || `${t('creativeId')} ${v.creative_id}`}</div>
                                <div style={{ fontSize: 12, color: '#555', marginTop: 4 }}>
                                  {t('creativeId')}: {v.creative_id} | {t('weight')}: {v.weight}
                                </div>
                                {appliedSP && appliedSP.length > 0 && (
                                  <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                                    {t('sellingPointsLabel')}：{appliedSP.slice(0, 2).join(' / ')}
                                  </div>
                                )}
                                {appliedCTA && (
                                  <div style={{ fontSize: 12, color: '#111', marginTop: 4 }}>
                                    CTA：{appliedCTA}
                                  </div>
                                )}
                              </div>
                            </div>
                          </div>
                        );
                      })}
                    </div>

                      {metricsLoading ? (
                        <div className="compact-loading" style={{ padding: 12 }}>
                          <div className="loading"></div>
                        </div>
                      ) : metrics && metrics.variants && metrics.variants.length > 0 ? (
                        <>
                          <div className="compact-section-title" style={{ marginTop: 12 }}>{t('metricsTitle')}</div>
                          {metricsSummary && metricsSummary.impressions > 0 ? (
                            <div className="metrics-grid">
                              <div className="metric-card">
                                <div className="metric-label">{t('totalImpressions')}</div>
                                <div className="metric-value">{metricsSummary.impressions}</div>
                              </div>
                              <div className="metric-card">
                                <div className="metric-label">{t('totalClicks')}</div>
                                <div className="metric-value">{metricsSummary.clicks}</div>
                              </div>
                              <div className="metric-card">
                                <div className="metric-label">{t('avgCtr')}</div>
                                <div className="metric-value">{(metricsSummary.avgCtr * 100).toFixed(2)}%</div>
                              </div>
                              <div className="metric-card highlight">
                                <div className="metric-label">{t('bestCtr')}</div>
                                <div className="metric-value">{(metricsSummary.best.ctr * 100).toFixed(2)}%</div>
                                <div className="metric-sub">Creative {metricsSummary.best.creative_id}</div>
                              </div>
                            </div>
                          ) : (
                            <div style={{ background: '#fffbe6', border: '1px solid #ffe58f', color: '#ad6800', borderRadius: 8, padding: 10, fontSize: 13 }}>
                              {t('noDataNotice')}
                            </div>
                          )}
                          <div className="compact-table-wrapper">
                            <table className="compact-table">
                              <thead>
                                <tr>
                                  <th>Creative ID</th>
                                  <th>{t('totalImpressions')}</th>
                                  <th>{t('totalClicks')}</th>
                                  <th>CTR</th>
                                  <th>{t('status')}</th>
                                </tr>
                              </thead>
                              <tbody>
                                {(metrics.variants || []).map((v) => {
                                  const diff = metricsSummary ? v.ctr - metricsSummary.avgCtr : 0;
                                  const diffText = `${(diff * 100).toFixed(2)}%`;
                                  return (
                                    <tr key={v.creative_id}>
                                      <td>{v.creative_id}</td>
                                      <td>{v.impressions}</td>
                                      <td>{v.clicks}</td>
                                      <td>{(v.ctr * 100).toFixed(2)}%</td>
                                      <td style={{ color: diff >= 0 ? '#52c41a' : '#ff7875' }}>
                                        {diff >= 0 ? t('aboveAvg') : t('belowAvg')} {diffText}
                                      </td>
                                    </tr>
                                  );
                                })}
                              </tbody>
                            </table>
                          </div>
                        </>
                      ) : (
                        <div style={{ background: '#fffbe6', border: '1px solid #ffe58f', color: '#ad6800', borderRadius: 8, padding: 10, fontSize: 13 }}>
                          {t('noMetrics')}
                        </div>
                      )}
                  </div>
                </div>
              )}
            </div>

          </div>
        </div>
      </div>
    </div>
  );
};

export default ExperimentsPage;
