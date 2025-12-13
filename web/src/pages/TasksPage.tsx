import React, { useState, useEffect } from 'react';
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
  const pageSize = 15;

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
    const statusMap: Record<string, { class: string; text: string; icon: string }> = {
      pending: { class: 'compact-status-pending', text: '待处理', icon: 'fa-clock' },
      queued: { class: 'compact-status-pending', text: '排队中', icon: 'fa-hourglass-half' },
      processing: { class: 'compact-status-processing', text: '处理中', icon: 'fa-spinner' },
      completed: { class: 'compact-status-completed', text: '已完成', icon: 'fa-check-circle' },
      failed: { class: 'compact-status-failed', text: '失败', icon: 'fa-times-circle' },
      cancelled: { class: 'compact-status-pending', text: '已取消', icon: 'fa-ban' },
    };

    const statusInfo = statusMap[status] || { class: 'compact-status-pending', text: status, icon: 'fa-question-circle' };

    return (
      <span className={`compact-status ${statusInfo.class}`}>
        <i className={`fas ${statusInfo.icon}`}></i>
        <span>{statusInfo.text}</span>
      </span>
    );
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

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
            <span>素材管理</span>
          </a>
          <a href="/tasks" className="nav-item active">
            <i className="fas fa-tasks"></i>
            <span>任务管理</span>
          </a>
        </nav>
      </div>

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">{showDetail ? '任务详情' : '任务管理'}</h1>
          <div className="user-info">
            <div className="avatar">A</div>
            <span>管理员</span>
          </div>
        </div>

        <div className="content">
          <div className="compact-layout">
            {!showDetail ? (
              <>
                <div className="compact-toolbar">
                  <div className="compact-toolbar-left">
                    <div className="compact-stats-text">
                      共 <strong>{total}</strong> 个任务
                    </div>
                  </div>
                  <div className="compact-toolbar-right">
                    <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={loadTasks} disabled={loading}>
                      <i className="fas fa-sync"></i>
                      <span>刷新</span>
                    </button>
                  </div>
                </div>

                {error && (
                  <div className="compact-alert compact-alert-error">
                    <i className="fas fa-exclamation-circle"></i>
                    <span>{error}</span>
                  </div>
                )}

                {loading ? (
                  <div className="compact-loading">
                    <div className="loading"></div>
                    <div className="compact-loading-text">加载中...</div>
                  </div>
                ) : tasks.length === 0 ? (
                  <div className="compact-empty">
                    <i className="fas fa-tasks"></i>
                    <div className="compact-empty-text">暂无任务</div>
                    <div className="compact-empty-hint">创建创意生成任务后将显示在这里</div>
                  </div>
                ) : (
                  <>
                    <div className="compact-table-wrapper">
                      <table className="compact-table">
                        <thead>
                          <tr>
                            <th style={{ width: '100px' }}>任务ID</th>
                            <th>标题</th>
                            <th style={{ width: '100px' }}>状态</th>
                            <th style={{ width: '150px' }}>进度</th>
                            <th style={{ width: '120px' }}>创建时间</th>
                            <th style={{ width: '120px' }}>完成时间</th>
                            <th style={{ width: '80px' }}>操作</th>
                          </tr>
                        </thead>
                        <tbody>
                          {tasks.map((task) => (
                            <tr key={task.id}>
                              <td>
                                <code className="compact-code" title={task.id}>
                                  {task.id.substring(0, 8)}
                                </code>
                              </td>
                              <td className="compact-table-title">{task.title}</td>
                              <td>{getStatusBadge(task.status)}</td>
                              <td>
                                <div className="compact-progress-wrapper">
                                  <div className="compact-progress-bar">
                                    <div
                                      className="compact-progress-fill"
                                      style={{ width: `${task.progress}%` }}
                                    ></div>
                                  </div>
                                  <span className="compact-progress-text">{task.progress}%</span>
                                </div>
                              </td>
                              <td className="compact-table-date">{formatDate(task.created_at)}</td>
                              <td className="compact-table-date">
                                {task.completed_at ? formatDate(task.completed_at) : '-'}
                              </td>
                              <td>
                                <button
                                  className="compact-btn compact-btn-text compact-btn-xs"
                                  onClick={() => viewTask(task.id)}
                                >
                                  详情
                                </button>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>

                    {totalPages > 1 && (
                      <div className="compact-pagination">
                        <button
                          className="compact-page-btn"
                          disabled={currentPage <= 1}
                          onClick={() => setCurrentPage(currentPage - 1)}
                        >
                          <i className="fas fa-chevron-left"></i>
                        </button>

                        <div className="compact-page-numbers">
                          {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                            const pageNum = i + Math.max(1, Math.min(currentPage - 2, totalPages - 4));
                            if (pageNum > totalPages) return null;
                            return (
                              <button
                                key={pageNum}
                                className={`compact-page-btn ${currentPage === pageNum ? 'active' : ''}`}
                                onClick={() => setCurrentPage(pageNum)}
                              >
                                {pageNum}
                              </button>
                            );
                          })}
                        </div>

                        <button
                          className="compact-page-btn"
                          disabled={currentPage >= totalPages}
                          onClick={() => setCurrentPage(currentPage + 1)}
                        >
                          <i className="fas fa-chevron-right"></i>
                        </button>

                        <div className="compact-page-info">
                          第 {currentPage} / {totalPages} 页
                        </div>
                      </div>
                    )}
                  </>
                )}
              </>
            ) : (
              <>
                <div className="compact-toolbar">
                  <div className="compact-toolbar-left">
                    <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={closeTaskDetail}>
                      <i className="fas fa-arrow-left"></i>
                      <span>返回列表</span>
                    </button>
                  </div>
                </div>

                {selectedTask && (
                  <>
                    <div className="compact-detail-grid">
                      <div className="compact-detail-item">
                        <div className="compact-detail-label">标题</div>
                        <div className="compact-detail-value">{selectedTask.title}</div>
                      </div>
                      <div className="compact-detail-item">
                        <div className="compact-detail-label">状态</div>
                        <div className="compact-detail-value">{getStatusBadge(selectedTask.status)}</div>
                      </div>
                      <div className="compact-detail-item">
                        <div className="compact-detail-label">进度</div>
                        <div className="compact-detail-value">
                          <div className="compact-progress-wrapper">
                            <div className="compact-progress-bar">
                              <div
                                className="compact-progress-fill"
                                style={{ width: `${selectedTask.progress}%` }}
                              ></div>
                            </div>
                            <span className="compact-progress-text">{selectedTask.progress}%</span>
                          </div>
                        </div>
                      </div>
                      <div className="compact-detail-item">
                        <div className="compact-detail-label">创建时间</div>
                        <div className="compact-detail-value">{formatDate(selectedTask.created_at || '')}</div>
                      </div>
                      <div className="compact-detail-item">
                        <div className="compact-detail-label">完成时间</div>
                        <div className="compact-detail-value">
                          {selectedTask.completed_at ? formatDate(selectedTask.completed_at) : '-'}
                        </div>
                      </div>
                      <div className="compact-detail-item">
                        <div className="compact-detail-label">任务ID</div>
                        <div className="compact-detail-value">
                          <code className="compact-code">{selectedTask.task_id}</code>
                        </div>
                      </div>
                    </div>

                    {selectedTask.error && (
                      <div className="compact-alert compact-alert-error">
                        <i className="fas fa-exclamation-circle"></i>
                        <span>{selectedTask.error}</span>
                      </div>
                    )}

                    {selectedTask.creatives && selectedTask.creatives.length > 0 && (
                      <>
                        <div className="compact-section-title">
                          生成素材 ({selectedTask.creatives.length})
                        </div>
                        <div className="compact-assets-grid">
                          {selectedTask.creatives.map((creative) => (
                            <div key={creative.id} className="compact-asset-card">
                              <div className="compact-asset-image-wrapper">
                                <img
                                  src={creative.image_url}
                                  alt={creative.id}
                                  className="compact-asset-image"
                                  onError={(e) => {
                                    const target = e.target as HTMLImageElement;
                                    target.src =
                                      'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="240" height="180" viewBox="0 0 240 180"><rect width="240" height="180" fill="%23f5f5f5"/><text x="120" y="90" font-family="Arial" font-size="11" text-anchor="middle" fill="%23999">素材图片</text></svg>';
                                  }}
                                />
                              </div>
                              <div className="compact-asset-info">
                                <div className="compact-asset-meta">
                                  <span className="compact-asset-format">{creative.format}</span>
                                  <span className="compact-asset-size">
                                    {creative.width}×{creative.height}
                                  </span>
                                </div>
                                <div className="compact-asset-id" title={creative.id}>
                                  {creative.id.substring(0, 8)}...
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </>
                    )}
                  </>
                )}
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default TasksPage;
