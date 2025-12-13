import React, { useState, useEffect } from 'react';
import { creativeAPI } from '../services/api';
import type { AssetData } from '../types';

const AssetsPage: React.FC = () => {
  const [assets, setAssets] = useState<AssetData[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pageSize = 20;

  useEffect(() => {
    loadAssets();
  }, [currentPage]);

  const loadAssets = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await creativeAPI.listAssets({
        page: currentPage,
        page_size: pageSize,
      });

      if (response.code === 0 && response.data) {
        setAssets(response.data.assets || []);
        setTotal(response.data.total);
        setTotalPages(response.data.total_pages);
      } else {
        setError(response.message || '获取素材列表失败');
      }
    } catch (err) {
      setError('加载素材失败: ' + (err as Error).message);
      console.error('Load assets error:', err);
    } finally {
      setLoading(false);
    }
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
          <a href="/assets" className="nav-item active">
            <i className="fas fa-images"></i>
            <span>素材管理</span>
          </a>
          <a href="/tasks" className="nav-item">
            <i className="fas fa-tasks"></i>
            <span>任务管理</span>
          </a>
        </nav>
      </div>

      <div className="main-content">
        <div className="header">
          <h1 className="page-title">素材管理</h1>
          <div className="user-info">
            <div className="avatar">A</div>
            <span>管理员</span>
          </div>
        </div>

        <div className="content">
          <div className="compact-layout">
            <div className="compact-toolbar">
              <div className="compact-toolbar-left">
                <div className="compact-stats-text">
                  共 <strong>{total}</strong> 个素材
                </div>
              </div>
              <div className="compact-toolbar-right">
                <button className="compact-btn compact-btn-outline compact-btn-sm" onClick={loadAssets} disabled={loading}>
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
            ) : assets.length === 0 ? (
              <div className="compact-empty">
                <i className="fas fa-images"></i>
                <div className="compact-empty-text">暂无素材</div>
                <div className="compact-empty-hint">生成创意后素材将显示在这里</div>
              </div>
            ) : (
              <>
                <div className="compact-assets-grid">
                  {assets.map((asset) => (
                    <div key={asset.id} className="compact-asset-card">
                      <div className="compact-asset-image-wrapper">
                        <img
                          src={asset.public_url || asset.image_url}
                          alt={asset.id}
                          className="compact-asset-image"
                          onError={(e) => {
                            const target = e.target as HTMLImageElement;
                            target.src = 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="240" height="180" viewBox="0 0 240 180"><rect width="240" height="180" fill="%23f5f5f5"/><text x="120" y="90" font-family="Arial" font-size="11" text-anchor="middle" fill="%23999">素材图片</text></svg>';
                          }}
                        />
                      </div>
                      <div className="compact-asset-info">
                        <div className="compact-asset-meta">
                          <span className="compact-asset-format">{asset.format}</span>
                          <span className="compact-asset-size">{asset.width}×{asset.height}</span>
                        </div>
                        <div className="compact-asset-id" title={asset.id}>
                          {asset.id.substring(0, 8)}...
                        </div>
                      </div>
                    </div>
                  ))}
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
          </div>
        </div>
      </div>
    </div>
  );
};

export default AssetsPage;
