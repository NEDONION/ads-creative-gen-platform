import React, { useEffect, useState } from 'react';
import { experimentAPI, creativeAPI } from '../services/api';
import type { ExperimentVariantInput, ExperimentMetrics, TaskListItem, Experiment } from '../types';

const ExperimentsPage: React.FC = () => {
  const [name, setName] = useState('');
  const [productName, setProductName] = useState('');
  const [variants, setVariants] = useState<ExperimentVariantInput[]>([{ creative_id: 0, weight: 0.5 }, { creative_id: 0, weight: 0.5 }]);
  const [creating, setCreating] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [expId, setExpId] = useState<string>('');
  const [metrics, setMetrics] = useState<ExperimentMetrics | null>(null);
  const [tasks, setTasks] = useState<TaskListItem[]>([]);
  const [productOptions, setProductOptions] = useState<string[]>([]);
  const [creativeOptions, setCreativeOptions] = useState<{ id: number; label: string; thumb?: string; product_name?: string; cta_text?: string; selling_points?: string[]; title?: string }[]>([]);
  const [experimentList, setExperimentList] = useState<Experiment[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [selectedExp, setSelectedExp] = useState<Experiment | null>(null);

  useEffect(() => {
    loadTasks();
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

  const loadTasks = async () => {
    try {
      const res = await creativeAPI.listTasks({ page: 1, page_size: 50 });
      if (res.code === 0 && res.data) {
        setTasks(res.data.tasks || []);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const loadOptions = async () => {
    try {
      const [tasksRes, assetsRes] = await Promise.all([
        creativeAPI.listTasks({ page: 1, page_size: 100 }),
        creativeAPI.listAssets({ page: 1, page_size: 100 }),
      ]);

      if (tasksRes.code === 0 && tasksRes.data) {
        const names = (tasksRes.data.tasks || [])
          .map((t) => t.product_name)
          .filter((n): n is string => Boolean(n));
        const uniqueNames = Array.from(new Set(names));
        setProductOptions(uniqueNames);
        if (!productName && uniqueNames.length > 0) {
          setProductName(uniqueNames[0]);
        }
      }

      if (assetsRes.code === 0 && assetsRes.data) {
        const taskProductMap: Record<string, string | undefined> = {};
        (tasksRes.data?.tasks || []).forEach((t) => {
          if (t.id && t.product_name) {
            taskProductMap[t.id] = t.product_name;
          }
        });

          const options = (assetsRes.data.assets || [])
          .map((asset) => {
            const idNum = parseInt(asset.id, 10);
            if (Number.isNaN(idNum)) return null;
            const resolvedProduct = asset.product_name || taskProductMap[String(asset.task_id)];
            const label = asset.title || resolvedProduct || '创意';
            return {
              id: idNum,
              label: `${label} (${asset.id})`,
              thumb: asset.image_url || asset.public_url,
              product_name: resolvedProduct,
              cta_text: asset.cta_text,
              selling_points: asset.selling_points,
              title: asset.title,
            };
          })
          .filter((v): v is { id: number; label: string; thumb?: string; product_name?: string; cta_text?: string; selling_points?: string[]; title?: string } => Boolean(v));
        setCreativeOptions(options);
      }
    } catch (err) {
      console.error(err);
      setMessage('加载商品/创意选项失败');
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

  const handleCreate = async () => {
    if (!name.trim()) {
      setMessage('请填写实验名称');
      return;
    }
    if (variants.some((v) => v.creative_id === 0 || v.weight <= 0)) {
      setMessage('请填写有效的创意ID和权重');
      return;
    }
    setCreating(true);
    setMessage(null);
    try {
      const res = await experimentAPI.create({ name: name.trim(), product_name: productName.trim(), variants });
      if (res.code === 0 && res.data) {
        setExpId(res.data.experiment_id);
        setMessage(`创建成功，ID: ${res.data.experiment_id}`);
        loadExperiments();
      } else {
        setMessage(res.message || '创建失败');
      }
    } catch (err) {
      setMessage('创建失败: ' + (err as Error).message);
    } finally {
      setCreating(false);
    }
  };

  const handleActivate = async (id?: string) => {
    const targetId = id || expId;
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
    const targetId = id || expId;
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
    const targetId = id || expId;
    if (!targetId) {
      setMessage('请先创建或填入实验ID');
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

  const updateVariant = (idx: number, field: keyof ExperimentVariantInput, value: number) => {
    setVariants((prev) => {
      const copy = [...prev];
      copy[idx] = { ...copy[idx], [field]: value };
      return copy;
    });
  };

  const addVariant = () => {
    setVariants((prev) => [...prev, { creative_id: 0, weight: 0.1 }]);
  };

  const removeVariant = (idx: number) => {
    setVariants((prev) => prev.filter((_, i) => i !== idx));
  };

  const getCreativeLabel = (id: number) => {
    const opt = creativeOptions.find((o) => o.id === id);
    if (opt) return opt.label;
    const t = tasks.find((t) => t.id && t.id.startsWith(id.toString().substring(0, 1))); // fallback
    return t ? `${t.title} (${t.id.slice(0, 6)})` : id;
  };

  const normalizeName = (s?: string) => (s || '').trim();
  const filteredCreativeOptions = productName
    ? creativeOptions.filter((opt) => normalizeName(opt.product_name) === normalizeName(productName))
    : creativeOptions;

  const getCreativeInfo = (id: number) => creativeOptions.find((opt) => opt.id === id);

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
    setExpId(id);
    setMetrics(null);
    const hit = experimentList.find((e) => e.experiment_id === id) || null;
    setSelectedExp(hit);
    setMessage(`已选择实验 ${id}`);
    loadMetrics(id);
  };

  useEffect(() => {
    if (!productName) return;
    const allowedIds = new Set(filteredCreativeOptions.map((o) => o.id));
    setVariants((prev) =>
      prev.map((v) => (allowedIds.has(v.creative_id) ? v : { ...v, creative_id: 0 }))
    );
  }, [productName, creativeOptions]);

  return (
    <div className="app">
      <div className="sidebar">
        <div className="sidebar-header">
          <h2>
            <i className="fas fa-bullseye"></i> <span>创意平台</span>
          </h2>
        </div>
        <nav className="nav-menu">
          <a href="/" className="nav-item">
            <i className="fas fa-home"></i>
            <span>仪表盘</span>
          </a>
          <a href="/creative" className="nav-item">
            <i className="fas fa-magic"></i>
            <span>创意生成</span>
          </a>
          <a href="/assets" className="nav-item">
            <i className="fas fa-images"></i>
            <span>创意管理</span>
          </a>
          <a href="/tasks" className="nav-item">
            <i className="fas fa-tasks"></i>
            <span>任务管理</span>
          </a>
          <a href="/experiments" className="nav-item active">
            <i className="fas fa-vial"></i>
            <span>实验</span>
          </a>
        </nav>
      </div>

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

            <div className="compact-card">
              <div className="compact-card-header">
                <h3 className="compact-card-title">实验列表</h3>
                <div className="compact-card-hint">查看当前 / 历史实验并快速选择实验ID</div>
                <button className="compact-btn compact-btn-text compact-btn-sm" onClick={loadExperiments} disabled={listLoading}>
                  <i className="fas fa-sync-alt"></i>
                  <span>{listLoading ? '加载中...' : '刷新'}</span>
                </button>
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
                            <td>{exp.status}</td>
                            <td>{exp.created_at ? new Date(exp.created_at).toLocaleString() : '-'}</td>
                            <td style={{ fontFamily: 'monospace', fontSize: 12 }}>{exp.experiment_id}</td>
                            <td>
                              <button className="compact-btn compact-btn-outline compact-btn-xs" onClick={() => handleSelectExperiment(exp.experiment_id)}>
                                查看
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
                <div className="compact-card" style={{ marginTop: 12 }}>
                  <div className="compact-card-header">
                    <h3 className="compact-card-title">实验详情</h3>
                    <div className="compact-card-hint">时长 / 定义 / 变体素材</div>
                    <div className="compact-card-actions" style={{ display: 'flex', gap: 8 }}>
                      {selectedExp.status !== 'archived' && (
                        <>
                          <button className="compact-btn compact-btn-outline compact-btn-xs" onClick={() => handleActivate(selectedExp.experiment_id)}>
                            激活
                          </button>
                          <button className="compact-btn compact-btn-outline compact-btn-xs" onClick={() => handleStop(selectedExp.experiment_id)}>
                            停止
                          </button>
                        </>
                      )}
                      <button className="compact-btn compact-btn-outline compact-btn-xs" onClick={() => loadMetrics(selectedExp.experiment_id)}>
                        刷新指标
                      </button>
                      <button className="compact-btn compact-btn-text compact-btn-xs" onClick={() => { setSelectedExp(null); setMetrics(null); }}>
                        收起
                      </button>
                    </div>
                  </div>
                  <div className="compact-card-body">
                    <div className="compact-form-grid">
                      <div className="compact-form-group">
                        <label className="compact-label">实验ID</label>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <code style={{ fontSize: 12 }}>{selectedExp.experiment_id}</code>
                          <button className="compact-btn compact-btn-outline compact-btn-xs" onClick={() => copyToClipboard(selectedExp.experiment_id)}>
                            复制
                          </button>
                        </div>
                      </div>
                      <div className="compact-form-group">
                        <label className="compact-label">状态</label>
                        <div>{selectedExp.status}</div>
                      </div>
                      <div className="compact-form-group">
                        <label className="compact-label">时长</label>
                        <div>{formatDuration(selectedExp.start_at || selectedExp.created_at, selectedExp.end_at)}</div>
                      </div>
                      <div className="compact-form-group">
                        <label className="compact-label">创建时间</label>
                        <div>{formatTime(selectedExp.created_at)}</div>
                      </div>
                      <div className="compact-form-group">
                        <label className="compact-label">开始时间</label>
                        <div>{formatTime(selectedExp.start_at)}</div>
                      </div>
                      <div className="compact-form-group">
                        <label className="compact-label">结束时间</label>
                        <div>{formatTime(selectedExp.end_at)}</div>
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
                                {info?.selling_points && info.selling_points.length > 0 && (
                                  <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                                    卖点：{info.selling_points.slice(0, 2).join(' / ')}
                                  </div>
                                )}
                                {info?.cta_text && (
                                  <div style={{ fontSize: 12, color: '#111', marginTop: 4 }}>
                                    CTA：{info.cta_text}
                                  </div>
                                )}
                              </div>
                            </div>
                          </div>
                        );
                      })}
                    </div>

                    {metrics && (
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
                              {metrics.variants.map((v) => (
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

            <div className="compact-card">
              <div className="compact-card-header">
                <h3 className="compact-card-title">创建实验</h3>
                <div className="compact-card-hint">配置变体和权重，创建后再激活</div>
              </div>
              <div className="compact-card-body">
                <div className="compact-form-grid">
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">实验名称</span>
                      <span className="label-required">*</span>
                    </label>
                    <input className="compact-input" value={name} onChange={(e) => setName(e.target.value)} />
                  </div>
                  <div className="compact-form-group">
                    <label className="compact-label">
                      <span className="label-text">商品名</span>
                    </label>
                    <select className="compact-input" value={productName} onChange={(e) => setProductName(e.target.value)}>
                      {productOptions.length === 0 && <option value="">暂无商品，请先创建任务</option>}
                      {productOptions.map((p) => (
                        <option key={p} value={p}>
                          {p}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>

                <div className="compact-section-title" style={{ marginTop: 8 }}>变体配置</div>
                <div className="compact-form-grid">
                  {variants.map((v, idx) => (
                    <div key={idx} className="compact-card" style={{ padding: 10, border: '1px solid #f0f0f0' }}>
                      <div className="compact-form-group">
                        <label className="compact-label">创意ID</label>
                        <select
                          className="compact-input"
                          value={v.creative_id || ''}
                          onChange={(e) => updateVariant(idx, 'creative_id', parseInt(e.target.value, 10) || 0)}
                        >
                          <option value="">请选择创意</option>
                          {filteredCreativeOptions.map((opt) => (
                            <option key={opt.id} value={opt.id}>
                              {opt.label}
                            </option>
                          ))}
                        </select>
                        {v.creative_id !== 0 && (
                          <div className="compact-thumb" style={{ marginTop: 6, display: 'flex', alignItems: 'center', gap: 8 }}>
                            {creativeOptions.find((opt) => opt.id === v.creative_id)?.thumb ? (
                              <img
                                src={creativeOptions.find((opt) => opt.id === v.creative_id)!.thumb}
                                alt="预览"
                                style={{ width: 80, height: 80, objectFit: 'cover', borderRadius: 6, border: '1px solid #eee' }}
                              />
                            ) : (
                              <div style={{ width: 80, height: 80, borderRadius: 6, border: '1px dashed #ccc', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999', fontSize: 12 }}>
                                无缩略图
                              </div>
                            )}
                            <div style={{ fontSize: 12, color: '#555' }}>
                              {getCreativeLabel(v.creative_id)}
                            </div>
                          </div>
                        )}
                      </div>
                      <div className="compact-form-group">
                        <label className="compact-label">权重 (0-1)</label>
                        <input
                          className="compact-input"
                          type="number"
                          step="0.1"
                          value={v.weight}
                          onChange={(e) => updateVariant(idx, 'weight', parseFloat(e.target.value) || 0)}
                        />
                      </div>
                      {variants.length > 2 && (
                        <button className="compact-btn compact-btn-text compact-btn-xs" onClick={() => removeVariant(idx)}>
                          删除
                        </button>
                      )}
                    </div>
                  ))}
                </div>
                <div className="compact-form-actions" style={{ marginTop: 8 }}>
                  <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={addVariant}>
                    <i className="fas fa-plus"></i>
                    <span>添加变体</span>
                  </button>
                  <button className="compact-btn compact-btn-primary" onClick={handleCreate} disabled={creating}>
                    <i className="fas fa-save"></i>
                    <span>{creating ? '创建中...' : '创建实验'}</span>
                  </button>
                </div>
              </div>
            </div>

          </div>
        </div>
      </div>
    </div>
  );
};

export default ExperimentsPage;
