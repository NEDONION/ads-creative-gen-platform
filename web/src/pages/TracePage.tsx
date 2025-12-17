import React, { useEffect, useMemo, useState } from 'react';
import Sidebar from '../components/Sidebar';
import { traceAPI } from '../services/api';
import type { TraceItem } from '../types';
import Header from '../components/Header';
import { useI18n } from '../i18n';

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
  const { t } = useI18n();
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
        setMessage(res.message || t('activityError'));
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
        setMessage(res.message || t('activityError'));
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
      { label: t('all'), value: '' },
      { label: 'success', value: 'success' },
      { label: 'failed', value: 'failed' },
      { label: 'running', value: 'running' },
    ],
    [t]
  );

  return (
    <div className="app">
      <Sidebar />
      <div className="main-content">
        <Header title={t('headerTraces')} />

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
                  <h3 className="compact-card-title">{t('traceListTitle')}</h3>
                  <div className="compact-card-hint">{t('traceListHint')}</div>
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
                    {t('query')}
                  </button>
                  <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={loadTraces} disabled={loading}>
                    <i className="fas fa-sync"></i>
                    <span style={{ marginLeft: 4 }}>{t('refresh')}</span>
                  </button>
                </div>
              </div>
              <div className="compact-card-body">
                {loading ? (
                  <div className="compact-loading">
                    <div className="loading"></div>
                  </div>
                ) : traces.length === 0 ? (
                  <div style={{ color: '#666', fontSize: 13 }}>{t('noData')}</div>
                ) : (
                  <div className="compact-table-wrapper">
                    <table className="compact-table">
                      <thead>
                        <tr>
                          <th>{t('traceId')}</th>
                          <th>{t('model')}</th>
                          <th>{t('status')}</th>
                          <th>{t('runTime')}</th>
                          <th>{t('startTime')}</th>
                          <th>{t('source')}</th>
                          <th>{t('ops')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {traces.map((trace) => {
                          const expanded = selected?.trace_id === trace.trace_id;
                          return (
                            <React.Fragment key={trace.trace_id}>
                              <tr className={expanded ? 'trace-row-expanded' : ''}>
                                <td style={{ fontFamily: 'monospace', fontSize: 12 }}>{trace.trace_id}</td>
                                <td>{trace.model_name} {trace.model_version && `(${trace.model_version})`}</td>
                                <td>
                                  <span className={`status-badge ${statusColor(trace.status)}`}>{trace.status}</span>
                                </td>
                                <td>{trace.duration_ms} ms</td>
                                <td>{trace.start_at ? new Date(trace.start_at).toLocaleString() : '-'}</td>
                                <td>{trace.source || '-'}</td>
                                <td>
                                  <button
                                    className={`compact-btn compact-btn-xs ${expanded ? 'compact-btn-outline' : 'compact-btn-primary'}`}
                                    onClick={() => selectTrace(trace.trace_id)}
                                  >
                                    {expanded ? t('collapse') : t('view')}
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
                                          <div className="trace-meta-label">{t('status')}</div>
                                          <span className={`status-badge ${statusColor(selected.status)}`}>{selected.status}</span>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">{t('runTime')}</div>
                                          <div className="trace-meta-value">{selected.duration_ms} ms</div>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">{t('model')}</div>
                                          <div className="trace-meta-value">{selected.model_name} {selected.model_version && `(${selected.model_version})`}</div>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">{t('start')}</div>
                                          <div className="trace-meta-value">{selected.start_at ? new Date(selected.start_at).toLocaleString() : '-'}</div>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">{t('end')}</div>
                                          <div className="trace-meta-value">{selected.end_at ? new Date(selected.end_at).toLocaleString() : '-'}</div>
                                        </div>
                                        <div>
                                          <div className="trace-meta-label">{t('source')}</div>
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
                                              <span>{t('runTime')}: {s.duration_ms} ms</span>
                                              <span>{t('start')}: {s.start_at ? new Date(s.start_at).toLocaleTimeString() : '-'}</span>
                                              <span>{t('end')}: {s.end_at ? new Date(s.end_at).toLocaleTimeString() : '-'}</span>
                                            </div>
                                            {s.input_preview && (
                                              <div className="trace-step-text"><strong>{t('input')}:</strong> {s.input_preview}</div>
                                            )}
                                            {s.output_preview && (
                                              <div className="trace-step-text"><strong>{t('output')}:</strong> {s.output_preview}</div>
                                            )}
                                            {s.error_message && (
                                              <div className="trace-step-text error"><strong>{t('error')}:</strong> {s.error_message}</div>
                                            )}
                                          </div>
                                        ))}
                                        {(selected.steps || []).length === 0 && <div className="trace-step-empty">{t('noSteps')}</div>}
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
