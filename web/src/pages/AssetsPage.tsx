import React, { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import { creativeAPI } from '../services/api';
import type { AssetData } from '../types';

const AssetsPage: React.FC = () => {
  const [assets, setAssets] = useState<AssetData[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pageSize = 12;

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
    <Layout title="素材管理">
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">素材管理</h3>
          <div className="search-bar">
            <button className="btn btn-primary" onClick={loadAssets} disabled={loading}>
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
          ) : assets.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '40px', color: '#6b7280' }}>暂无素材</div>
          ) : (
            <>
              <div className="assets-grid">
                {assets.map((asset) => (
                  <div key={asset.id} className="asset-card">
                    <img
                      src={asset.public_url || asset.image_url}
                      alt={asset.id}
                      className="asset-image"
                      onError={(e) => {
                        const target = e.target as HTMLImageElement;
                        target.src = 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="280" height="200" viewBox="0 0 280 200"><rect width="280" height="200" fill="%23f0f0f0"/><text x="140" y="100" font-family="Arial" font-size="12" text-anchor="middle" fill="%23999">素材图片</text></svg>';
                      }}
                    />
                    <div className="asset-info">
                      <div className="asset-title">{asset.id.substring(0, 12)}...</div>
                      <div className="asset-meta">
                        <span>{asset.format}</span>
                        <span>{asset.width}×{asset.height}</span>
                      </div>
                    </div>
                  </div>
                ))}
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
    </Layout>
  );
};

export default AssetsPage;
