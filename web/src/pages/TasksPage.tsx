import React, { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import { creativeAPI } from '../services/api';
import type { TaskListItem, TaskDetailData } from '../types';

const TasksPage: React.FC = () => {
  const [tasks, setTasks] = useState<TaskListItem[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedTask, setSelectedTask] = useState<TaskDetailData | null>(null);
  const [showDetail, setShowDetail] = useState(false);
  const pageSize = 10;

  useEffect(() => {
    loadTasks();
  }, [currentPage]);

  const loadTasks = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await creativeAPI.listTasks({
        page: currentPage,
        page_size: pageSize,
      });

      if (response.code === 0 && response.data) {
        setTasks(response.data.tasks || []);
        setTotal(response.data.total);
        setTotalPages(response.data.total_pages);
      } else {
        setError(response.message || '获取任务列表失败');
      }
    } catch (err) {
      setError('加载任务失败: ' + (err as Error).message);
      console.error('Load tasks error:', err);
    } finally {
      setLoading(false);
    }
  };

  const viewTask = async (taskId: string) => {
    try {
      const response = await creativeAPI.getTask(taskId);
      if (response.code === 0 && response.data) {
        setSelectedTask(response.data);
        setShowDetail(true);
      } else {
        alert('获取任务详情失败: ' + response.message);
      }
    } catch (err) {
      alert('获取任务详情失败: ' + (err as Error).message);
      console.error('Get task detail error:', err);
    }
  };

  const closeTaskDetail = () => {
    setShowDetail(false);
    setSelectedTask(null);
  };

  const getStatusBadge = (status: string) => {
    const statusMap: Record<string, { class: string; text: string }> = {
      pending: { class: 'status-pending', text: '待处理' },
      queued: { class: 'status-pending', text: '排队中' },
      processing: { class: 'status-processing', text: '处理中' },
      completed: { class: 'status-completed', text: '已完成' },
      failed: { class: 'status-failed', text: '失败' },
      cancelled: { class: 'status-pending', text: '已取消' },
    };

    const statusInfo = statusMap[status] || { class: 'status-pending', text: status };

    return (
      <span className={`status-badge ${statusInfo.class}`}>
        <i className="fas fa-circle"></i> {statusInfo.text}
      </span>
    );
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <Layout title="任务管理">
      {!showDetail ? (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">任务管理</h3>
            <div className="search-bar">
              <button className="btn btn-primary" onClick={loadTasks} disabled={loading}>
                <i className="fas fa-sync"></i> 刷新
              </button>
            </div>
          </div>
          <div className="card-body">
            {error && (
              <div style={{ padding: '16px', background: '#fee2e2', color: '#ef4444', borderRadius: '8px', marginBottom: '16px' }}>
                {error}
              </div>
            )}

            {loading ? (
              <div style={{ textAlign: 'center', padding: '40px', color: '#6b7280' }}>
                <div className="loading"></div>
                <div style={{ marginTop: '12px' }}>加载中...</div>
              </div>
            ) : tasks.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '40px', color: '#6b7280' }}>暂无任务</div>
            ) : (
              <>
                <div className="table-container">
                  <table className="table">
                    <thead>
                      <tr>
                        <th>任务ID</th>
                        <th>标题</th>
                        <th>状态</th>
                        <th>进度</th>
                        <th>创建时间</th>
                        <th>完成时间</th>
                        <th>操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {tasks.map((task) => (
                        <tr key={task.id}>
                          <td title={task.id}>{task.id.substring(0, 8)}...</td>
                          <td>{task.title}</td>
                          <td>{getStatusBadge(task.status)}</td>
                          <td>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                              <div className="progress-bar" style={{ flex: 1 }}>
                                <div className="progress-fill" style={{ width: `${task.progress}%` }}></div>
                              </div>
                              <span>{task.progress}%</span>
                            </div>
                          </td>
                          <td>{formatDate(task.created_at)}</td>
                          <td>{task.completed_at ? formatDate(task.completed_at) : '-'}</td>
                          <td>
                            <button className="btn btn-outline btn-sm" onClick={() => viewTask(task.id)}>
                              <i className="fas fa-eye"></i> 详情
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {totalPages > 1 && (
                  <div className="pagination">
                    <button
                      className="page-btn"
                      disabled={currentPage <= 1}
                      onClick={() => setCurrentPage(currentPage - 1)}
                    >
                      上一页
                    </button>
                    {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                      const pageNum = i + Math.max(1, currentPage - 2);
                      if (pageNum > totalPages) return null;
                      return (
                        <button
                          key={pageNum}
                          className={`page-btn ${currentPage === pageNum ? 'active' : ''}`}
                          onClick={() => setCurrentPage(pageNum)}
                        >
                          {pageNum}
                        </button>
                      );
                    })}
                    <button
                      className="page-btn"
                      disabled={currentPage >= totalPages}
                      onClick={() => setCurrentPage(currentPage + 1)}
                    >
                      下一页
                    </button>
                    <div style={{ lineHeight: '32px', color: 'var(--gray)' }}>
                      共 {total} 项，第 {currentPage} 页，共 {totalPages} 页
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      ) : (
        <div className="card">
          <div className="card-header" style={{ background: '#f8fafc' }}>
            <h3 className="card-title">{selectedTask?.title || '任务详情'}</h3>
            <button className="btn btn-outline btn-sm" onClick={closeTaskDetail}>
              <i className="fas fa-times"></i> 关闭
            </button>
          </div>
          <div className="card-body">
            {selectedTask && (
              <>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '24px', marginBottom: '24px' }}>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--primary)', marginBottom: '4px' }}>
                      {getStatusBadge(selectedTask.status)}
                    </div>
                    <div style={{ color: 'var(--gray)', fontSize: '0.9rem' }}>状态</div>
                  </div>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--primary)', marginBottom: '4px' }}>
                      {selectedTask.progress}%
                    </div>
                    <div style={{ color: 'var(--gray)', fontSize: '0.9rem' }}>进度</div>
                  </div>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--dark)', marginBottom: '4px' }}>
                      {formatDate(selectedTask.created_at || '')}
                    </div>
                    <div style={{ color: 'var(--gray)', fontSize: '0.9rem' }}>创建时间</div>
                  </div>
                  <div style={{ textAlign: 'center' }}>
                    <div style={{ fontSize: '1rem', fontWeight: 700, color: 'var(--dark)', marginBottom: '4px' }}>
                      {selectedTask.completed_at ? formatDate(selectedTask.completed_at) : '-'}
                    </div>
                    <div style={{ color: 'var(--gray)', fontSize: '0.9rem' }}>完成时间</div>
                  </div>
                </div>

                {selectedTask.error && (
                  <div className="form-group">
                    <label className="form-label">错误信息</label>
                    <div style={{ padding: '12px 16px', minHeight: '60px', background: '#fef2f2', color: '#dc2626', borderRadius: '8px', border: '1px solid #fecaca' }}>
                      {selectedTask.error}
                    </div>
                  </div>
                )}

                {selectedTask.creatives && selectedTask.creatives.length > 0 && (
                  <div className="form-group">
                    <label className="form-label">生成素材 ({selectedTask.creatives.length})</label>
                    <div className="assets-grid">
                      {selectedTask.creatives.map((creative) => (
                        <div key={creative.id} className="asset-card">
                          <img
                            src={creative.image_url}
                            alt={creative.id}
                            className="asset-image"
                            onError={(e) => {
                              const target = e.target as HTMLImageElement;
                              target.src = 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="280" height="200" viewBox="0 0 280 200"><rect width="280" height="200" fill="%23f0f0f0"/><text x="140" y="100" font-family="Arial" font-size="12" text-anchor="middle" fill="%23999">素材图片</text></svg>';
                            }}
                          />
                          <div className="asset-info">
                            <div className="asset-title">{creative.id.substring(0, 8)}...</div>
                            <div className="asset-meta">
                              <span>{creative.format}</span>
                              <span>{creative.width}×{creative.height}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      )}
    </Layout>
  );
};

export default TasksPage;
