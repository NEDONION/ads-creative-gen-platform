import React, { useEffect, useMemo, useState } from 'react';
import Sidebar from '../components/Sidebar';
import { traceAPI } from '../services/api';
import type { TraceItem } from '../types';

const statusColor = (status: string) => {
  switch (status) {
    case 'success':
      return 'status-completed';
    case 'failed':
      return 'status-failed';
    case 'running':
      return 'status-processing';
    default:
      return 'status-pending';
  }
};

const TracePage: React.FC = () => {
  const [traces, setTraces] = useState<TraceItem[]>([]);
  const [selected, setSelected] = useState<TraceItem | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [filters, setFilters] = useState({ trace_id: '', status: '' });

  const loadTraces = async () => {
    setLoading(true);
    try {
      const res = await traceAPI.list({
        trace_id: filters.trace_id || undefined,
        status: filters.status || undefined,
        product_name: undefined,
      });
      if (res.code === 0 && res.data) {
        setTraces(res.data.traces || []);
      } else {
        setMessage(res.message || '加载失败');
      }
    } catch (err) {
      setMessage((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const selectTrace = async (id: string) => {
    if (selected && selected.trace_id === id) {
      setSelected(null);
      return;
    }
    setSelected(null);
    try {
      const res = await traceAPI.detail(id);
      if (res.code === 0 && res.data) {
        setSelected(res.data);
      } else {
        setMessage(res.message || '获取详情失败');
      }
    } catch (err) {
      setMessage((err as Error).message);
    }
  };

  useEffect(() => {
    loadTraces();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const statusOptions = useMemo(
    () => [
      { label: '全部', value: '' },
      { label: 'success', value: 'success' },
      { label: 'failed', value: 'failed' },
      { label: 'running', value: 'running' },
    ],
    []
  );

  return (
    <div className="app">
      <Sidebar />
      <div className="main-content">
        <div className="header">
          <h1 className="page-title">调用链路</h1>
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
              <div className="compact-card-header" style={{ alignItems: 'flex-end', gap: 12 }}>
                <div>
                  <h3 className="compact-card-title">调用列表</h3>
                  <div className="compact-card-hint">trace 概览</div>
                </div>
                <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                  <input
                    className="compact-input"
                    style={{ width: 160 }}
                    placeholder="Trace ID"
                    value={filters.trace_id}
                    onChange={(e) => setFilters((f) => ({ ...f, trace_id: e.target.value }))}
                  />
                  <select
                    className="compact-input"
                    style={{ width: 120 }}
                    value={filters.status}
                    onChange={(e) => setFilters((f) => ({ ...f, status: e.target.value }))}
                  >
                    {statusOptions.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                  <button className="compact-btn compact-btn-primary compact-btn-sm" onClick={loadTraces} disabled={loading}>
                    查询
                  </button>
                </div>
              </div>
              <div className="compact-card-body">
                {loading ? (
                  <div className="compact-loading">
                    <div className="loading"></div>
                  </div>
                ) : traces.length === 0 ? (
                  <div style={{ color: '#666', fontSize: 13 }}>暂无数据</div>
                ) : (
                  <div className="compact-table-wrapper">
                    <table className="compact-table">
                      <thead>
                        <tr>
                          <th>Trace ID</th>
                          <th>模型</th>
                          <th>状态</th>
                          <th>耗时</th>
                          <th>开始时间</th>
                          <th>来源</th>
                          <th>操作</th>
                        </tr>
                      </thead>
                      <tbody>
                        {traces.map((t) => {
                          const expanded = selected?.trace_id === t.trace_id;
                          return (
                            <React.Fragment key={t.trace_id}>
                              <tr className={expanded ? 'trace-row-expanded' : ''}>
                                <td style={{ fontFamily: 'monospace', fontSize: 12 }}>{t.trace_id}</td>
                                <td>{t.model_name} {t.model_version && `(${t.model_version})`}</td>
                                <td>
                                  <span className={`status-badge ${statusColor(t.status)}`}>{t.status}</span>
                                </td>
                                <td>{t.duration_ms} ms</td>
                                <td>{t.start_at ? new Date(t.start_at).toLocaleString() : '-'}</td>
                                <td>{t.source || '-'}</td>
                                <td>
                                  <button
                                    className={`compact-btn compact-btn-xs ${expanded ? 'compact-btn-outline' : 'compact-btn-primary'}`}
                                    onClick={() => selectTrace(t.trace_id)}
                                  >
                                    {expanded ? '收起' : '查看'}
                                  </button>
                                </td>
                              </tr>
                              {expanded && selected && (
                                <tr className="trace-detail-row">
                                  <td colSpan={7}>
                                    <div className="trace-detail">
                                      <div className="trace-detail-meta">
                                        <div>
                                          <div className="trace-meta-label">Trace ID</div>
                                          <div className="trace-meta-value" style={{ fontFamily: 'monospace' }}>{selected.trace_id}</div>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">状态</div>
                                          <span className={`status-badge ${statusColor(selected.status)}`}>{selected.status}</span>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">耗时</div>
                                          <div className="trace-meta-value">{selected.duration_ms} ms</div>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">模型</div>
                                          <div className="trace-meta-value">{selected.model_name} {selected.model_version && `(${selected.model_version})`}</div>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">开始</div>
                                          <div className="trace-meta-value">{selected.start_at ? new Date(selected.start_at).toLocaleString() : '-'}</div>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">结束</div>
                                          <div className="trace-meta-value">{selected.end_at ? new Date(selected.end_at).toLocaleString() : '-'}</div>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">来源</div>
                                          <div className="trace-meta-value">{selected.source || '-'}</div>
                                        </div>
                                      </div>
                                      <div className="trace-steps">
                                        {(selected.steps || []).filter((s) => s.step_name !== 'query_task').map((s, idx) => (
                                          <div key={idx} className="trace-step-card">
                                            <div className="trace-step-header">
                                              <div>
                                                <div className="trace-step-title">{s.step_name}</div>
                                                <div className="trace-step-subtitle">{s.component}</div>
                                              </div>
                                              <span className={`status-badge ${statusColor(s.status)}`}>{s.status}</span>
                                            </div>
                                            <div className="trace-step-meta">
                                              <span>耗时: {s.duration_ms} ms</span>
                                              <span>开始: {s.start_at ? new Date(s.start_at).toLocaleTimeString() : '-'}</span>
                                              <span>结束: {s.end_at ? new Date(s.end_at).toLocaleTimeString() : '-'}</span>
                                            </div>
                                            {s.input_preview && (
                                              <div className="trace-step-text"><strong>输入:</strong> {s.input_preview}</div>
                                            )}
                                            {s.output_preview && (
                                              <div className="trace-step-text"><strong>输出:</strong> {s.output_preview}</div>
                                            )}
                                            {s.error_message && (
                                              <div className="trace-step-text error"><strong>错误:</strong> {s.error_message}</div>
                                            )}
                                          </div>
                                        ))}
                                        {(selected.steps || []).length === 0 && <div className="trace-step-empty">暂无步骤数据</div>}
                                      </div>
                                    </div>
                                  </td>
                                </tr>
                              )}
                            </React.Fragment>
                          );
                        })}
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

export default TracePage;
