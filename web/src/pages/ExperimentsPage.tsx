import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { experimentAPI, creativeAPI } from '../services/api';
import Sidebar from '../components/Sidebar';
import type { ExperimentMetrics, Experiment } from '../types';

const ExperimentsPage: React.FC = () => {
  const [message, setMessage] = useState<string | null>(null);
  const [metrics, setMetrics] = useState<ExperimentMetrics | null>(null);
  const [creativeOptions, setCreativeOptions] = useState<{ id: string; label: string; thumb?: string; product_name?: string; cta_text?: string; selling_points?: string[]; title?: string }[]>([]);
  const [experimentList, setExperimentList] = useState<Experiment[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [selectedExp, setSelectedExp] = useState<Experiment | null>(null);

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
            const label = asset.title || resolvedProduct || '创意';
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
      setMessage('加载创意选项失败');
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
      setMessage('加载实验列表失败');
    } finally {
      setListLoading(false);
    }
  };

  const handleActivate = async (id?: string) => {
    const targetId = id || selectedExp?.experiment_id;
    if (!targetId) return;
    try {
      const res = await experimentAPI.updateStatus(targetId, 'active');
      if (res.code === 0) {
        setMessage('已激活实验');
        loadExperiments();
      } else {
        setMessage(res.message || '激活失败');
      }
    } catch (err) {
      setMessage('激活失败: ' + (err as Error).message);
    }
  };

  const handleStop = async (id?: string) => {
    const targetId = id || selectedExp?.experiment_id;
    if (!targetId) return;
    try {
      const res = await experimentAPI.updateStatus(targetId, 'archived');
      if (res.code === 0) {
        setMessage('已停止实验');
        loadExperiments();
      } else {
        setMessage(res.message || '停止失败');
      }
    } catch (err) {
      setMessage('停止失败: ' + (err as Error).message);
    }
  };

  const loadMetrics = async (id?: string) => {
    const targetId = id || selectedExp?.experiment_id;
    if (!targetId) {
      setMessage('请先选择实验');
      return;
    }
    try {
      const res = await experimentAPI.metrics(targetId);
      if (res.code === 0 && res.data) {
        setMetrics(res.data);
      } else {
        setMessage(res.message || '获取结果失败');
      }
    } catch (err) {
      setMessage('获取结果失败: ' + (err as Error).message);
    }
  };

  const getCreativeInfo = (id: number | string) => creativeOptions.find((opt) => opt.id === String(id));

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setMessage('已复制实验ID');
    } catch (err) {
      console.error('copy failed', err);
      setMessage('复制失败');
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
      return `${days}天${hours % 24}小时`;
    }
    if (hours > 0) return `${hours}小时${minutes}分`;
    return `${minutes}分`;
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
    setMessage(`已选择实验 ${id}`);
    loadMetrics(id);
  };

  return (
    <div className="app">
      <Sidebar />

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">实验管理</h1>
          <div className="user-info">
            <div className="avatar">A</div>
            <span>管理员</span>
          </div>
        </div>

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
                  <h3 className="compact-card-title">实验列表</h3>
                  <div className="compact-card-hint">轻量面板：查看 / 激活 / 停止 / 指标</div>
                </div>
                <div style={{ display: 'flex', gap: 8 }}>
                  <button className="compact-btn compact-btn-text compact-btn-sm" onClick={loadExperiments} disabled={listLoading}>
                    <i className="fas fa-sync-alt"></i>
                    <span>{listLoading ? '加载中...' : '刷新'}</span>
                  </button>
                  <Link className="compact-btn compact-btn-primary compact-btn-sm" to="/experiments/new">
                    新建实验
                  </Link>
                </div>
              </div>
              <div className="compact-card-body">
                {listLoading ? (
                  <div className="compact-loading">
                    <div className="loading"></div>
                  </div>
                ) : experimentList.length === 0 ? (
                  <div style={{ color: '#666', fontSize: 13 }}>暂无实验，先创建一个吧</div>
                ) : (
                  <div className="compact-table-wrapper">
                    <table className="compact-table">
                      <thead>
                        <tr>
                          <th>名称</th>
                          <th>商品</th>
                          <th>状态</th>
                          <th>创建时间</th>
                          <th>实验ID</th>
                          <th>操作</th>
                        </tr>
                      </thead>
                      <tbody>
                        {experimentList.map((exp) => (
                          <tr key={exp.experiment_id}>
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
                                {selectedExp?.experiment_id === exp.experiment_id ? '收起' : '查看'}
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
              {selectedExp && (
                <div className="compact-card section-card section-detail" style={{ marginTop: 16, borderColor: '#e6f0ff', boxShadow: '0 8px 18px rgba(0,0,0,0.05)' }}>
                  <div className="compact-card-header" style={{ alignItems: 'flex-start' }}>
                    <div>
                      <h3 className="compact-card-title">实验详情</h3>
                      <div className="compact-card-hint">时长 / 定义 / 变体素材</div>
                    </div>
                    <div className="compact-card-actions" style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                      <button
                        className="compact-btn compact-btn-primary compact-btn-xs"
                        onClick={() => handleActivate(selectedExp.experiment_id)}
                        disabled={selectedExp.status === 'active'}
                      >
                        {selectedExp.status === 'active' ? '已激活' : '激活'}
                      </button>
                      <button
                        className="compact-btn compact-btn-danger compact-btn-xs"
                        style={{ marginLeft: 'auto' }}
                        onClick={() => handleStop(selectedExp.experiment_id)}
                        disabled={selectedExp.status === 'archived'}
                      >
                        {selectedExp.status === 'archived' ? '已停止' : '停止'}
                      </button>
                      <button className="compact-btn compact-btn-outline compact-btn-xs" onClick={() => loadMetrics(selectedExp.experiment_id)}>
                        刷新指标
                      </button>
                      <button className="compact-btn compact-btn-text compact-btn-xs" onClick={() => { setSelectedExp(null); setMetrics(null); }}>
                        收起
                      </button>
                    </div>
                  </div>
                  <div className="compact-card-body">
                    <div className="detail-meta-grid">
                      <div className="meta-block wide">
                        <div className="meta-label">实验ID</div>
                        <div className="meta-value code">
                          <code>{selectedExp.experiment_id}</code>
                          <button className="compact-btn compact-btn-outline compact-btn-xs" onClick={() => copyToClipboard(selectedExp.experiment_id)}>
                            复制
                          </button>
                        </div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">状态</div>
                        <div className="meta-value"><span className={`status-badge status-${selectedExp.status}`}>{selectedExp.status}</span></div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">时长</div>
                        <div className="meta-value">{formatDuration(selectedExp.start_at || selectedExp.created_at, selectedExp.end_at)}</div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">创建</div>
                        <div className="meta-value">{formatTime(selectedExp.created_at)}</div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">开始</div>
                        <div className="meta-value">{formatTime(selectedExp.start_at)}</div>
                      </div>
                      <div className="meta-block">
                        <div className="meta-label">结束</div>
                        <div className="meta-value">{formatTime(selectedExp.end_at)}</div>
                      </div>
                    </div>

                    <div className="compact-section-title" style={{ marginTop: 12 }}>变体定义</div>
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
                                <img src={info.thumb} alt="创意" style={{ width: 80, height: 80, objectFit: 'cover', borderRadius: 6, border: '1px solid #eee' }} />
                              ) : (
                                <div style={{ width: 80, height: 80, borderRadius: 6, border: '1px dashed #ccc', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999', fontSize: 12 }}>
                                  无图
                                </div>
                              )}
                              <div style={{ flex: 1 }}>
                                <div style={{ fontWeight: 700 }}>{info?.title || info?.product_name || `创意 ${v.creative_id}`}</div>
                                <div style={{ fontSize: 12, color: '#555', marginTop: 4 }}>
                                  创意ID: {v.creative_id} | 权重: {v.weight}
                                </div>
                                {appliedSP && appliedSP.length > 0 && (
                                  <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                                    卖点：{appliedSP.slice(0, 2).join(' / ')}
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

                    {metrics && metrics.variants && (
                      <>
                        <div className="compact-section-title" style={{ marginTop: 12 }}>当前指标</div>
                        <div className="compact-table-wrapper">
                          <table className="compact-table">
                            <thead>
                              <tr>
                                <th>Creative ID</th>
                                <th>曝光</th>
                                <th>点击</th>
                                <th>CTR</th>
                              </tr>
                            </thead>
                            <tbody>
                              {(metrics.variants || []).map((v) => (
                                <tr key={v.creative_id}>
                                  <td>{v.creative_id}</td>
                                  <td>{v.impressions}</td>
                                  <td>{v.clicks}</td>
                                  <td>{(v.ctr * 100).toFixed(2)}%</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </>
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
